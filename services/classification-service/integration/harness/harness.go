//go:build integration

package harness

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"csort.ru/classification-service/internal/config"
	"csort.ru/classification-service/internal/correlation"
	"csort.ru/classification-service/internal/server"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	serviceID              = "classification-test-service"
	serviceSecret          = "classification-test-secret"
	identityMockBaseURL    = "http://identity.integration.test/api/v1"
	correlationMockBaseURL = "http://correlation.integration.test"
)

type jwtSigner struct {
	mu     sync.RWMutex
	key    *rsa.PrivateKey
	jwks   jose.JSONWebKeySet
	signer jose.Signer
}

func newJWTSigner() (*jwtSigner, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	jwk := jose.JSONWebKey{
		Key:       key.Public(),
		KeyID:     "test-key-id",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithHeader("kid", "test-key-id"),
	)
	if err != nil {
		return nil, err
	}
	return &jwtSigner{
		key:    key,
		jwks:   jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}},
		signer: signer,
	}, nil
}

func (j *jwtSigner) getJWKS() jose.JSONWebKeySet {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.jwks
}

func (j *jwtSigner) SignUserToken(
	telegramID, maxID *int64,
	sub string,
	roles []string,
) (string, error) {
	j.mu.RLock()
	signer := j.signer
	j.mu.RUnlock()

	now := time.Now()
	exp := now.Add(500 * time.Minute)
	claims := map[string]any{
		"iss":   "mock-auth",
		"sub":   sub,
		"exp":   exp.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"roles": roles,
		"type":  "access",
	}
	if telegramID != nil {
		claims["telegram_id"] = *telegramID
	}
	if maxID != nil {
		claims["max_id"] = *maxID
	}
	if sub != "service" {
		uid := int32(1)
		claims["user_id"] = uid
	}
	return jwt.Signed(signer).Claims(claims).Serialize()
}

func setupHTTPMocks(signer *jwtSigner, svcID, svcSecret string) {
	httpmock.RegisterResponder("GET", identityMockBaseURL+"/auth/.well-known/jwks.json",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, signer.getJWKS())
		})

	httpmock.RegisterResponder("GET", identityMockBaseURL+"/health",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(http.StatusOK, `{"status":"healthy"}`), nil
		})

	httpmock.RegisterResponder("GET", correlationMockBaseURL+"/health",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(http.StatusOK, `{"status":"healthy"}`), nil
		})

	httpmock.RegisterResponder("POST", identityMockBaseURL+"/auth/login",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Grant-Type") != "client_credentials" {
				return httpmock.NewStringResponse(400, `{"error":"invalid grant type"}`), nil
			}
			var creds struct {
				ServiceID     string `json:"service_id"`
				ServiceSecret string `json:"service_secret"`
			}
			if err := json.NewDecoder(req.Body).Decode(&creds); err != nil {
				return httpmock.NewStringResponse(400, `{"error":"invalid body"}`), nil
			}
			if creds.ServiceID != svcID || creds.ServiceSecret != svcSecret {
				return httpmock.NewStringResponse(401, `{"error":"invalid credentials"}`), nil
			}
			tok, err := signer.SignUserToken(nil, nil, "service", []string{"service"})
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return httpmock.NewJsonResponse(200, map[string]any{
				"success": true,
				"result": map[string]string{
					"access_token":  tok,
					"refresh_token": tok,
				},
			})
		})

	httpmock.RegisterResponder("POST", identityMockBaseURL+"/auth/refresh",
		func(req *http.Request) (*http.Response, error) {
			var body struct {
				RefreshToken string `json:"refresh_token"`
			}
			if err := json.NewDecoder(req.Body).
				Decode(&body); err != nil ||
				body.RefreshToken == "" {
				resp, errResp := httpmock.NewJsonResponse(400, map[string]any{
					"error": map[string]string{"message": "missing refresh_token"},
				})
				if errResp != nil {
					return nil, errResp
				}
				return resp, nil
			}
			tok, err := signer.SignUserToken(nil, nil, "service", []string{"service"})
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return httpmock.NewJsonResponse(200, map[string]any{
				"success": true,
				"result": map[string]string{
					"access_token":  tok,
					"refresh_token": tok,
				},
			})
		})

	corrPat := "=~^" + strings.ReplaceAll(correlationMockBaseURL, ".", `\.`) + ".*"
	httpmock.RegisterResponder("POST", corrPat, func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		var creq correlation.CorrelationRequest
		if err := json.Unmarshal(body, &creq); err != nil {
			return httpmock.NewStringResponse(400, `{"error":"invalid json"}`), nil
		}
		if len(creq.Fractions) == 0 {
			return httpmock.NewStringResponse(400, `{"error":"fractions required"}`), nil
		}
		if len(creq.ParameterGroups) == 0 {
			return httpmock.NewStringResponse(400, `{"error":"parameter_groups required"}`), nil
		}
		out := []correlation.CorrelationWithTest{
			{
				CorrelationBase: correlation.CorrelationBase{
					Name:       "integration-mock",
					Conditions: []correlation.CorrelationCondition{},
				},
			},
		}
		return httpmock.NewJsonResponse(200, out)
	})

	identityPat := "=~^" + strings.ReplaceAll(
		identityMockBaseURL,
		".",
		`\.`,
	) + "/users/by-messenger-id/.*"
	httpmock.RegisterResponder("GET", identityPat, func(req *http.Request) (*http.Response, error) {
		resp, _ := httpmock.NewJsonResponse(200, map[string]any{
			"success": true,
			"result": map[string]any{
				"id":          1,
				"telegram_id": TestUserTelegramID,
				"full_name":   "Test User",
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			},
		})
		return resp, nil
	})
}

func serviceTokenFromMock(svcID, svcSecret string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"service_id":     svcID,
		"service_secret": svcSecret,
	})
	req, err := http.NewRequest(
		http.MethodPost,
		identityMockBaseURL+"/auth/login",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grant-Type", "client_credentials")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth login failed: %d %s", resp.StatusCode, string(b))
	}
	var result struct {
		Success bool `json:"success"`
		Result  struct {
			AccessToken string `json:"access_token"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if !result.Success || result.Result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	return result.Result.AccessToken, nil
}

func getPortFromConnStr(connStr string) int {
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		connStr = strings.TrimPrefix(connStr, "postgres://")
		connStr = strings.TrimPrefix(connStr, "postgresql://")
	}
	parts := strings.Split(connStr, "@")
	if len(parts) != 2 {
		return 5432
	}
	hostPart := parts[1]
	if idx := strings.Index(hostPart, "/"); idx >= 0 {
		hostPart = hostPart[:idx]
	}
	hp := strings.Split(hostPart, ":")
	if len(hp) < 2 {
		return 5432
	}
	port, _ := strconv.Atoi(hp[1])
	if port <= 0 {
		return 5432
	}
	return port
}

const TestUserTelegramID int64 = 12345

const TestUserID int32 = 1

type TestEnv struct {
	PostgresCtr               *postgres.PostgresContainer
	DBPool                    *pgxpool.Pool
	Server                    *server.Server
	HTTPServer                *httptest.Server
	BaseURL                   string
	Config                    *config.Config
	ServiceToken              string
	UserToken                 string
	AdminToken                string
	ClassificationEditorToken string
}

func SetupTestEnv(ctx context.Context) (*TestEnv, func(), error) {
	signer, err := newJWTSigner()
	if err != nil {
		return nil, nil, err
	}
	setupHTTPMocks(signer, serviceID, serviceSecret)

	_, harnessFile, _, _ := runtime.Caller(0)
	schemaPath := filepath.Join(filepath.Dir(harnessFile), "..", "..", "database", "schema.sql")
	initScripts := []string{schemaPath}

	postgresCtr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("classification_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithOrderedInitScripts(initScripts...),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, err
	}

	pgConnStr, err := postgresCtr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(pgConnStr)
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}

	mr, err := miniredis.Run()
	if err != nil {
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	redisHost, redisPortStr, err := net.SplitHostPort(mr.Addr())
	if err != nil {
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil {
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}

	cfg := config.TestConfig(
		"localhost",
		getPortFromConnStr(pgConnStr),
		"postgres", "postgres", "classification_test",
		identityMockBaseURL,
		correlationMockBaseURL,
	)
	cfg.Redis.Host = redisHost
	cfg.Redis.Port = redisPort

	srv, err := server.New(cfg, dbPool)
	if err != nil {
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}

	ts := httptest.NewServer(adaptor.FiberApp(srv.App()))

	serviceToken, err := serviceTokenFromMock(serviceID, serviceSecret)
	if err != nil {
		ts.Close()
		_ = srv.Shutdown()
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	tgID := TestUserTelegramID
	userToken, err := signer.SignUserToken(&tgID, nil, "test-user", []string{"user"})
	if err != nil {
		ts.Close()
		_ = srv.Shutdown()
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	adminToken, err := signer.SignUserToken(&tgID, nil, "admin-user", []string{"admin"})
	if err != nil {
		ts.Close()
		_ = srv.Shutdown()
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}
	editorToken, err := signer.SignUserToken(
		&tgID,
		nil,
		"editor-user",
		[]string{"classification_editor"},
	)
	if err != nil {
		ts.Close()
		_ = srv.Shutdown()
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
		return nil, nil, err
	}

	env := &TestEnv{
		PostgresCtr:               postgresCtr,
		DBPool:                    dbPool,
		Server:                    srv,
		HTTPServer:                ts,
		BaseURL:                   ts.URL,
		Config:                    cfg,
		ServiceToken:              serviceToken,
		UserToken:                 userToken,
		AdminToken:                adminToken,
		ClassificationEditorToken: editorToken,
	}

	cleanup := func() {
		ts.Close()
		_ = srv.Shutdown()
		mr.Close()
		dbPool.Close()
		_ = testcontainers.TerminateContainer(postgresCtr)
	}

	return env, cleanup, nil
}

func SeedProduct(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(ctx, `INSERT INTO products (name) VALUES ($1) RETURNING id`, name).
		Scan(&id)
	require.NoError(t, err)
	return id
}

func SeedClassification(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	productID uuid.UUID,
	createdBy int32,
	name string,
) uuid.UUID {
	t.Helper()
	var classID uuid.UUID
	err := pool.QueryRow(
		ctx,
		`INSERT INTO classifications (name, created_by, is_public, product_id) VALUES ($1, $2, false, $3) RETURNING id`,
		name,
		createdBy,
		productID,
	).Scan(&classID)
	require.NoError(t, err)

	var fractionID uuid.UUID
	err = pool.QueryRow(
		ctx,
		`INSERT INTO fractions (name, classification_id, order_index) VALUES ($1, $2, 0) RETURNING id`,
		"f1",
		classID,
	).Scan(&fractionID)
	require.NoError(t, err)

	var conditionID uuid.UUID
	err = pool.QueryRow(
		ctx,
		`INSERT INTO conditions (name, fraction_id, operator, connection, order_index) VALUES ($1, $2, 'OR', 'AND', 0) RETURNING id`,
		"c1",
		fractionID,
	).Scan(&conditionID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO params (name, condition_id, operator, value) VALUES ($1, $2, '>=', 0)`,
		"p1", conditionID,
	)
	require.NoError(t, err)

	return classID
}

func SeedParam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO classification_params (name) VALUES ($1) RETURNING id`,
		name,
	).Scan(&id)
	require.NoError(t, err)
	return id
}
