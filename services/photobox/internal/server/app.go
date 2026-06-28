package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"csort.ru/coffeebot/internal/authz"
	"csort.ru/coffeebot/internal/config"
	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/middleware"
	"csort.ru/coffeebot/internal/minio"
	redisclient "csort.ru/coffeebot/internal/redis"
	"csort.ru/coffeebot/internal/server/service"
	"csort.ru/coffeebot/internal/transport/response"

	"github.com/bytedance/sonic"
	fibotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Server struct {
	app         *fiber.App
	log         zerolog.Logger
	redisClient *redis.Client
}

func New(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	minioClient *minio.Client,
	identityClient authz.IdentityClient,
	httpPort int,
	rateLimitCfg config.RateLimitConfig,
	redisClient *redis.Client,
	log zerolog.Logger,
) (*Server, error) {
	if dbPool == nil {
		return nil, errors.New("dbPool is nil")
	}
	if minioClient == nil {
		return nil, errors.New("minioClient is nil")
	}
	if identityClient == nil {
		return nil, errors.New("identityClient is nil")
	}

	srvLog := log.With().Str("component", "server").Logger()

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		ErrorHandler: response.NewFiberErrorHandler(
			log.With().Str("component", "transport.response").Logger(),
		),
		Services: service.Compose(dbPool, identityClient),
	})

	setupMiddleware(ctx, app, httpPort, log, srvLog)
	app.Use(middleware.RateLimit(rateLimitCfg, redisClient))

	services := initializeServices(dbPool, minioClient, identityClient, log)

	handlers := initializeHandlers(services, log)

	setupRoutes(ctx, app, handlers, services.IdentityClient, log)

	return &Server{
		app:         app,
		log:         srvLog,
		redisClient: redisClient,
	}, nil
}

func setupMiddleware(
	ctx context.Context,
	app *fiber.App,
	httpPort int,
	baseLog, srvLog zerolog.Logger,
) {
	app.Use(fibotel.Middleware(
		fibotel.WithoutMetrics(true),
		fibotel.WithPort(httpPort),
	))

	app.Use(middleware.AccessLog(ctx, baseLog.With().Str("component", "http_access").Logger()))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, err any) {
			if e, ok := err.(error); ok {
				srvLog.Error().Msg(e.Error())
			}
		},
	}))

	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
}

func (s *Server) Start(port int) error {
	return s.app.Listen(fmt.Sprintf(":%d", port))
}

func (s *Server) Shutdown() error {
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	return s.app.Shutdown()
}

// App exposes the Fiber instance (e.g. httptest, adaptor).
func (s *Server) App() *fiber.App {
	return s.app
}

// Test runs an in-process request against the Fiber app.
func (s *Server) Test(req *http.Request, config ...fiber.TestConfig) (*http.Response, error) {
	return s.app.Test(req, config...)
}

// NewFromConfig builds pool, MinIO, identity (JWKS refresh), and Server; Fiber Services close them on shutdown.
func NewFromConfig(cfg *config.Config, log zerolog.Logger) (*Server, error) {
	dbPool, err := database.NewPool(
		cfg.DatabaseURL(),
		cfg.InternalDB.MaxConns,
		cfg.InternalDB.MinConns,
		cfg.InternalDB.MaxConnLifetime,
		cfg.InternalDB.MaxConnIdleTime,
	)
	if err != nil {
		return nil, fmt.Errorf("database pool: %w", err)
	}

	minioClient, err := minio.NewClient(context.Background(), minio.MinioClientConfig{
		Host:           cfg.Minio.Host,
		Port:           cfg.Minio.Port,
		AccessKey:      cfg.Minio.AccessKey,
		SecretKey:      cfg.Minio.SecretKey,
		Bucket:         cfg.Minio.Bucket,
		UseSSL:         cfg.Minio.UseSSL,
		PublicEndpoint: cfg.Minio.PublicEndpoint,
	}, log.With().Str("component", "minio.Client").Logger())
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("minio client: %w", err)
	}

	identityClient := authz.NewClient(
		cfg.Identity.ServiceURL,
		log.With().Str("component", "authz.Client").Logger(),
	)
	if err := identityClient.Start(context.Background()); err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("identity client: %w", err)
	}

	redisClient := redisclient.NewRateLimitClient(cfg.Redis, log)

	srv, err := New(
		context.Background(),
		dbPool,
		minioClient,
		identityClient,
		cfg.Server.Port,
		cfg.RateLimit,
		redisClient,
		log,
	)
	if err != nil {
		identityClient.Stop()
		dbPool.Close()
		return nil, err
	}

	return srv, nil
}
