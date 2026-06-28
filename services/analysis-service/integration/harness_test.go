//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/api/reports"
	apifiber "csort.ru/analysis-service/internal/apierrors/fiber"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/database"
	"csort.ru/analysis-service/internal/repository/kalibr"
	repo_requests "csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/server"
	"csort.ru/analysis-service/internal/storage"
	transporthttp "csort.ru/analysis-service/internal/transport/http"
	"github.com/alicebob/miniredis/v2"
	"github.com/bytedance/sonic"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

const testTelegramUserID = int64(9001)

type flowTestStack struct {
	App     *fiber.App
	Pool    *pgxpool.Pool
	Token   string
	Cleanup func()
}

type jwksFixture struct {
	PrivateKey *rsa.PrivateKey
	Server     *httptest.Server
}

func newJWKSFixture(t *testing.T) *jwksFixture {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwk := jose.JSONWebKey{
		Key:       priv.Public(),
		KeyID:     "integration-key",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	}))
	t.Cleanup(srv.Close)
	return &jwksFixture{PrivateKey: priv, Server: srv}
}

func (j *jwksFixture) mintAccessToken(t *testing.T) string {
	t.Helper()
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: j.PrivateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "integration-key"),
	)
	require.NoError(t, err)

	var uid int32 = 1
	tg := testTelegramUserID
	custom := auth.CustomClaims{
		UserID:     &uid,
		TelegramID: &tg,
		Roles:      []string{},
		Type:       "access",
	}
	std := jwt.Claims{
		Subject:  "integration-user",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	raw, err := jwt.Signed(sig).Claims(std).Claims(custom).Serialize()
	require.NoError(t, err)
	return string(raw)
}

func newReportsStubServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/reports/") && r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"analysisId":"stub","message":"generated"}`))
			return
		}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/download-url") &&
			strings.HasPrefix(r.URL.Path, "/api/reports/") {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			format := r.URL.Query().Get("format")
			blobURL := scheme + "://" + r.Host + "/_stub_report_blob?format=" + url.QueryEscape(format)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"url":%q,"expiresInSeconds":3600}`, blobURL)
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/_stub_report_blob" {
			format := r.URL.Query().Get("format")
			switch format {
			case "pdf":
				w.Header().Set("Content-Type", "application/pdf")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("%PDF-stub"))
				return
			default:
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("id,value\n1,ok\n"))
				return
			}
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newWorkerStubServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/merge" && r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/recalculate" && r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Response":"new-analysis-from-recalc"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newFlowTestStack(t *testing.T) *flowTestStack {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	jwks := newJWKSFixture(t)
	reportsSrv := newReportsStubServer(t)
	workerSrv := newWorkerStubServer(t)

	pgContainer := runPostgresIT(t, ctx, "flow_it")

	rmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12.11-management-alpine")
	require.NoError(t, err)

	minioContainer, err := minio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(readKalibrSchema(t)))
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(readRequestsInitSchema(t)))
	require.NoError(t, err)

	amqpURL, err := rmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	minioHostPort, err := minioContainer.ConnectionString(ctx)
	require.NoError(t, err)
	minioHost, minioPortStr, err := net.SplitHostPort(minioHostPort)
	require.NoError(t, err)
	minioPort, err := strconv.Atoi(minioPortStr)
	require.NoError(t, err)

	tempBucket := "temp-images-it"
	analysisBucket := "analysis-images-it"

	tempClient, err := storage.New(ctx, &storage.ImageStorageConfig{
		Host:         minioHost,
		Port:         minioPort,
		RootUser:     minioContainer.Username,
		RootPassword: minioContainer.Password,
		Bucket:       tempBucket,
		UseSSL:       false,
	})
	require.NoError(t, err)

	analysisClient, err := storage.NewForRead(ctx, &storage.ImageStorageConfig{
		Host:         minioHost,
		Port:         minioPort,
		RootUser:     minioContainer.Username,
		RootPassword: minioContainer.Password,
		Bucket:       analysisBucket,
		UseSSL:       false,
	})
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	redisHost, redisPort, _ := net.SplitHostPort(mr.Addr())
	redisPortNum, _ := strconv.Atoi(redisPort)
	redisClient := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(redisHost, strconv.Itoa(redisPortNum)),
	})
	require.NoError(t, redisClient.Ping(ctx).Err())

	kalibrDB := &database.DB{Pool: pool}
	requestsDB := &database.DB{Pool: pool}

	reportsHTTP, err := reports.NewClient(reports.Config{
		BaseURL:      reportsSrv.URL,
		Timeout:      10 * time.Second,
		MaxRetries:   1,
		GetToken:     func() string { return "integration-reports-token" },
		RefreshToken: func(context.Context) error { return nil },
	})
	require.NoError(t, err)

	cfg := &config.Config{
		KalibrDB:   config.DBConfig{Host: "unused"},
		RequestsDB: config.DBConfig{Host: "unused"},
		App:        config.AppConfig{Port: 6040, MaxQueuedRequests: 10},
		RabbitMQ: config.RabbitMQConfig{
			URL:               amqpURL,
			RequestExchange:   "request_exchange",
			RequestQueue:      "request_queue",
			RequestRoutingKey: "request",
		},
		OutboxRelay: config.OutboxRelayConfig{
			BatchSize:            10,
			PollIntervalSec:      1,
			CleanupIntervalHr:    1,
			MessageTTLHours:      24,
			MarkPublishedRetries: 3,
		},
		Redis: config.RedisConfig{Host: redisHost, Port: redisPortNum},
		ImageStorage: config.ImageStorageConfig{
			Host:           minioHost,
			Port:           minioPort,
			RootUser:       minioContainer.Username,
			RootPassword:   minioContainer.Password,
			TempBucket:     tempBucket,
			AnalysisBucket: analysisBucket,
			UseSSL:         false,
		},
		AnalysisAPI: workerSrv.URL,
		ReportsAPI:  reportsSrv.URL,
		Security: config.SecurityConfig{
			ServiceUrl:    strings.TrimRight(jwks.Server.URL, "/"),
			ServiceId:     "it-service",
			ServiceSecret: "it-secret",
		},
		ShareLink: config.ShareLinkConfig{
			HMACSecret:      "0123456789abcdef0123456789abcdef",
			MaxClockSkewSec: 60,
		},
	}

	container, err := server.IntegrationWireContainer(
		ctx,
		kalibr.New(pool),
		repo_requests.New(pool),
		pool,
		redisClient,
		tempClient,
		analysisClient,
		cfg,
		server.IntegrationInfraExtras{ReportsClient: reportsHTTP},
	)
	require.NoError(t, err)

	require.NoError(t, container.AuthClient.UpdateJWKS(ctx))

	handlers := server.IntegrationInitHandlers(container, cfg.ShareLink)
	health := transporthttp.NewHealthHandler(
		kalibrDB,
		requestsDB,
		redisClient,
		container.RabbitMQClient,
	)

	app := fiber.New(fiber.Config{
		ErrorHandler:      apifiber.ErrorHandler,
		JSONEncoder:       sonic.Marshal,
		JSONDecoder:       sonic.Unmarshal,
		BodyLimit:         200 * 1024 * 1024,
		StreamRequestBody: true,
	})
	server.IntegrationSetupMiddleware(app, redisClient, "http://127.0.0.1/api/v1")
	server.IntegrationMountRoutes(app, &handlers, health.Handle, redisClient, cfg.App.MaxQueuedRequests)

	tok := jwks.mintAccessToken(t)

	cleanup := func() {
		_ = app.Shutdown()
		redisClient.Close()
		pool.Close()
		_ = pgContainer.Terminate(context.WithoutCancel(ctx))
		_ = rmqContainer.Terminate(context.WithoutCancel(ctx))
		_ = minioContainer.Terminate(context.WithoutCancel(ctx))
	}
	t.Cleanup(cleanup)

	return &flowTestStack{
		App:     app,
		Pool:    pool,
		Token:   tok,
		Cleanup: cleanup,
	}
}

func mustRequest(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func decodeAPIResult(t *testing.T, resp *http.Response, into any) {
	t.Helper()
	defer resp.Body.Close()
	var wrap struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&wrap))
	require.True(t, wrap.Success, "API success=false")
	require.NoError(t, json.Unmarshal(wrap.Result, into))
}
