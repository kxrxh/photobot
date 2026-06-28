package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apierr "csort.ru/analysis-service/internal/apierrors/fiber"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/database"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/repository/kalibr"
	"csort.ru/analysis-service/internal/repository/requests"
	svc "csort.ru/analysis-service/internal/server/services"
	"csort.ru/analysis-service/internal/storage"
	http "csort.ru/analysis-service/internal/transport/http"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	redispkg "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	app *fiber.App

	maxQueuedRequests int
	container         *Container
	dbKalibr          *database.DB
	dbRequests        *database.DB
	cancel            context.CancelFunc
}

func New(
	ctx context.Context,
	cfg *config.Config,
) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}

	ctx, cancel := context.WithCancel(ctx)

	kalibrDB, err := database.New(ctx, &database.DatabaseConfig{
		Host:            cfg.KalibrDB.Host,
		Port:            cfg.KalibrDB.Port,
		User:            cfg.KalibrDB.User,
		Password:        cfg.KalibrDB.Password,
		Name:            cfg.KalibrDB.Name,
		SSLMode:         cfg.KalibrDB.SSLMode,
		MaxConns:        cfg.KalibrDB.MaxConns,
		MinConns:        cfg.KalibrDB.MinConns,
		MaxConnLifetime: cfg.KalibrDB.MaxConnLifetime,
		MaxConnIdleTime: cfg.KalibrDB.MaxConnIdleTime,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("connect kalibr database: %w", err)
	}
	requestsDB, err := database.New(ctx, &database.DatabaseConfig{
		Host:            cfg.RequestsDB.Host,
		Port:            cfg.RequestsDB.Port,
		User:            cfg.RequestsDB.User,
		Password:        cfg.RequestsDB.Password,
		Name:            cfg.RequestsDB.Name,
		SSLMode:         cfg.RequestsDB.SSLMode,
		MaxConns:        cfg.RequestsDB.MaxConns,
		MinConns:        cfg.RequestsDB.MinConns,
		MaxConnLifetime: cfg.RequestsDB.MaxConnLifetime,
		MaxConnIdleTime: cfg.RequestsDB.MaxConnIdleTime,
	})
	if err != nil {
		cancel()
		kalibrDB.Close()
		return nil, fmt.Errorf("connect requests database: %w", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		cancel()
		requestsDB.Close()
		kalibrDB.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	dbKalibr := kalibr.New(kalibrDB.Pool)
	dbRequests := requests.New(requestsDB.Pool)

	tempStorageClient, err := storage.New(ctx, &storage.ImageStorageConfig{
		Host:         cfg.ImageStorage.Host,
		Port:         cfg.ImageStorage.Port,
		RootUser:     cfg.ImageStorage.RootUser,
		RootPassword: cfg.ImageStorage.RootPassword,
		Bucket:       cfg.ImageStorage.TempBucket,
		UseSSL:       cfg.ImageStorage.UseSSL,
		ExternalHost: cfg.ImageStorage.ExternalHost,
	})
	if err != nil {
		cancel()
		_ = redisClient.Close()
		requestsDB.Close()
		kalibrDB.Close()
		return nil, fmt.Errorf("failed to initialize MinIO temp images client: %w", err)
	}

	analysisStorageClient, err := storage.NewForRead(ctx, &storage.ImageStorageConfig{
		Host:         cfg.ImageStorage.Host,
		Port:         cfg.ImageStorage.Port,
		RootUser:     cfg.ImageStorage.RootUser,
		RootPassword: cfg.ImageStorage.RootPassword,
		Bucket:       cfg.ImageStorage.AnalysisBucket,
		UseSSL:       cfg.ImageStorage.UseSSL,
		ExternalHost: cfg.ImageStorage.ExternalHost,
	})
	if err != nil {
		cancel()
		_ = redisClient.Close()
		requestsDB.Close()
		kalibrDB.Close()
		return nil, fmt.Errorf("failed to initialize MinIO analysis images client: %w", err)
	}

	serviceContainer, err := initializeServiceContainer(
		ctx,
		dbKalibr,
		dbRequests,
		requestsDB.Pool,
		redisClient,
		tempStorageClient,
		analysisStorageClient,
		cfg,
	)
	if err != nil {
		cancel()
		_ = redisClient.Close()
		requestsDB.Close()
		kalibrDB.Close()
		return nil, err
	}
	handlers := initializeHandlers(serviceContainer, cfg.ShareLink)

	lifecycleServices := []fiber.Service{
		svc.NewKalibrDatabaseService(kalibrDB),
		svc.NewRequestsDatabaseService(requestsDB),
		svc.NewRedisService(redisClient),
		svc.NewRabbitMQService(serviceContainer.RabbitMQClient, 5*time.Minute),
		svc.NewStorageService(svc.TempMinIOServiceName, tempStorageClient),
		svc.NewStorageService(
			svc.AnalysisMinIOServiceName,
			analysisStorageClient,
		),
		svc.NewAuthClientService(serviceContainer.AuthClient),
		svc.NewOutboxRelayWorkerService(serviceContainer.OutboxRelay),
		svc.NewOutboxCleanupWorkerService(serviceContainer.OutboxRelay),
		svc.NewRequestsCleanupWorkerService(serviceContainer.RequestsService),
		svc.NewStuckProcessingCleanupService(
			serviceContainer.RequestsService,
			serviceContainer.RequestsPool,
		),
		svc.NewWebSocketHubService(serviceContainer.WebSocketHub),
	}

	app := fiber.New(fiber.Config{
		ErrorHandler:                    apierr.ErrorHandler,
		JSONEncoder:                     sonic.Marshal,
		JSONDecoder:                     sonic.Unmarshal,
		BodyLimit:                       200 * 1024 * 1024, // matches transport maxTotalSize
		StreamRequestBody:               true,
		EnableSplittingOnParsers:        false,
		Services:                        lifecycleServices,
		ServicesStartupContextProvider:  context.Background,
		ServicesShutdownContextProvider: context.Background,
		TrustProxy:                      true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: []string{
				"172.16.0.0/12",
				"192.168.0.0/16",
				"127.0.0.1",
			},
		},
	})

	app.Hooks().OnPreShutdown(func() error {
		serviceContainer.AuthClient.EndJWKSBackground()
		return nil
	})

	//nolint:contextcheck // Fiber middleware stack; request context is per-request
	setupMiddleware(app, redisClient, cfg.AnalysisAPI)

	s := &Server{
		app:               app,
		maxQueuedRequests: cfg.App.MaxQueuedRequests,
		container:         serviceContainer,
		dbKalibr:          kalibrDB,
		dbRequests:        requestsDB,
		cancel:            cancel,
	}

	healthHandler := http.NewHealthHandler(
		kalibrDB,
		requestsDB,
		serviceContainer.RedisClient,
		serviceContainer.RabbitMQClient,
	)

	//nolint:contextcheck // Route registration only; request context is created and handled by Fiber handlers.
	defineRoutes(
		app,
		&handlers,
		healthHandler.Handle,
		serviceContainer.RedisClient,
		cfg.App.MaxQueuedRequests,
	)

	if err := svc.StartServices(ctx, lifecycleServices); err != nil {
		cancel()
		_ = s.app.Shutdown()
		return nil, fmt.Errorf("start services: %w", err)
	}

	return s, nil
}

func setupMiddleware(
	app *fiber.App,
	redisClient *redis.Client,
	analysisAPIURL string,
) {
	app.Use(middleware.RequestTrace())

	app.Use(logger.FiberRequestLogger())

	app.Use(middleware.AccessLog())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			if err, ok := e.(error); ok {
				logger.Logger.Error().Err(err).Msg("panic recovered")
			}
		},
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Messenger-Platform",
			"X-Bot-Name",
		},
	}))

	store := redispkg.NewFromConnection(redisClient)
	app.Use(limiter.New(limiter.Config{
		Max:               120,
		Expiration:        1 * time.Minute,
		Storage:           store,
		LimiterMiddleware: limiter.SlidingWindow{},
		Next: func(c fiber.Ctx) bool {
			if c.Method() != fiber.MethodGet {
				return false
			}
			path := c.Path()
			if path == "/api/v1/health" {
				return true
			}
			return strings.Contains(path, "/analyses/") && strings.Contains(path, "/images/")
		},
		KeyGenerator: func(c fiber.Ctx) string {
			return "analysis:" + c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		},
	}))

	app.Use(compress.New())

	app.Use(middleware.BaseURL(analysisAPIURL))
}

func (s *Server) Start(port int) error {
	return s.app.Listen(fmt.Sprintf(":%d", port))
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) Shutdown() error {
	logger.Logger.Info().Msg("graceful shutdown started")

	shutdownErr := s.app.Shutdown()
	if shutdownErr != nil {
		logger.Logger.Error().Err(shutdownErr).Msg("fiber shutdown failed")
	}
	if s.cancel != nil {
		s.cancel()
	}

	logger.Logger.Info().Msg("graceful shutdown completed")
	return shutdownErr
}
