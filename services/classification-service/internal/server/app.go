package server

import (
	"context"
	"errors"
	"fmt"

	"csort.ru/classification-service/internal/config"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/middleware"
	redisclient "csort.ru/classification-service/internal/redis"
	"csort.ru/classification-service/internal/server/routes"
	transporthttp "csort.ru/classification-service/internal/transport/http"
	"csort.ru/classification-service/internal/transport/response"

	"github.com/bytedance/sonic"
	fibotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	app         *fiber.App
	cancel      context.CancelFunc
	redisClient *redis.Client
}

func New(cfg *config.Config, dbPool *pgxpool.Pool) (*Server, error) {
	if dbPool == nil {
		return nil, errors.New("dbPool is nil")
	}
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}

	redisClient, err := redisclient.NewRateLimitClient(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis for rate limiting: %w", err)
	}
	closeRedis := func() {
		_ = redisClient.Close()
	}

	app := fiber.New(newFiberAppConfig(cfg))
	setupMiddleware(app, cfg.Server.Port, cfg.RateLimit, redisClient)

	services := initializeServices(dbPool, cfg)

	handlers := initializeHandlers(services)

	ctx, cancel := context.WithCancel(context.Background())

	if err := services.AuthServiceClient.Start(ctx); err != nil {
		cancel()
		closeRedis()
		return nil, fmt.Errorf("failed to start identity client: %w", err)
	}

	if err := services.AuthTokenManager.Start(ctx); err != nil {
		cancel()
		closeRedis()
		return nil, fmt.Errorf("failed to start auth token manager: %w", err)
	}

	if err := services.CorrelationTokenManager.Start(ctx); err != nil {
		cancel()
		closeRedis()
		return nil, fmt.Errorf("failed to start correlation token manager: %w", err)
	}

	healthHandler := transporthttp.NewHealthHandler(
		dbPool,
		cfg.AuthServiceURL,
		cfg.CorrelationServiceURL,
	)
	apiRoutes := routes.Define(&routes.Handlers{
		ClassificationHandler:           handlers.ClassificationHandler,
		ProductHandler:                  handlers.ProductHandler,
		UserActiveClassificationHandler: handlers.UserActiveClassificationHandler,
		MarkupHandler:                   handlers.MarkupHandler,
		CorrelationHandler:              handlers.CorrelationHandler,
		ClassificationParamsHandler:     handlers.ClassificationParamsHandler,
		OwnershipHandler:                handlers.OwnershipHandler,
	}, services.AuthServiceClient, healthHandler.GetHealth)
	registerRoutes(app, apiRoutes)

	return &Server{
		app:         app,
		cancel:      cancel,
		redisClient: redisClient,
	}, nil
}

func newFiberAppConfig(cfg *config.Config) fiber.Config {
	fc := fiber.Config{
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ErrorHandler: response.FiberErrorHandler,
	}
	if cfg.HTTP.TrustProxy {
		fc.TrustProxy = true
		fc.ProxyHeader = fiber.HeaderXForwardedFor
		if cfg.HTTP.TrustPrivateProxies {
			fc.TrustProxyConfig = fiber.TrustProxyConfig{Private: true}
		}
	}
	return fc
}

func setupMiddleware(
	app *fiber.App,
	httpPort int,
	rateLimitCfg config.RateLimitConfig,
	redisClient *redis.Client,
) {
	app.Use(fibotel.Middleware(
		fibotel.WithoutMetrics(true),
		fibotel.WithPort(httpPort),
	))

	app.Use(logger.FiberRequestLogger())

	app.Use(middleware.AccessLog())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			if err, ok := e.(error); ok {
				logger.Logger.Error().Err(err).Msg("handle request failed")
			}
		},
	}))

	app.Use(middleware.RateLimit(rateLimitCfg, redisClient))
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

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) Shutdown() error {
	s.cancel()
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	return s.app.Shutdown()
}
