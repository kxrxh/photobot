package router

import (
	"time"

	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/http/handlers"
	"csort.ru/reports-service/internal/http/middleware"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/storage"

	"github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func New(
	cfg config.Config,
	reportsHandler *handlers.ReportsHandler,
	minio *storage.MinIOService,
	rateLimitStorage fiber.Storage,
) (*fiber.App, error) {
	fcfg := fiber.Config{
		ReadTimeout: time.Hour,
		BodyLimit:   10 * 1024 * 1024,
	}
	if cfg.Server.TrustProxy {
		fcfg.TrustProxy = true
		fcfg.ProxyHeader = fiber.HeaderXForwardedFor
		fcfg.TrustProxyConfig = fiber.TrustProxyConfig{Private: true, Loopback: true}
	}

	app := fiber.New(fcfg)

	if rateLimitStorage != nil {
		app.Hooks().OnPreShutdown(func() error {
			return rateLimitStorage.Close()
		})
	}

	app.Use(otel.Middleware(
		otel.WithoutMetrics(true),
		otel.WithPort(cfg.Server.Port),
	))
	app.Use(logger.FiberRequestLogger())
	app.Use(requestid.New())
	app.Use(middleware.AccessLog())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e interface{}) {
			if err, ok := e.(error); ok {
				logger.Logger.Error().Err(err).Msg("panic recovered")
			}
		},
	}))
	app.Use(compress.New())

	sec := helmet.ConfigDefault
	sec.CrossOriginEmbedderPolicy = "unsafe-none"
	sec.CrossOriginOpenerPolicy = "unsafe-none"
	sec.CrossOriginResourcePolicy = "cross-origin"
	app.Use(helmet.New(sec))

	allowedOrigins := parseAllowedOrigins(cfg.Server.CORSAllowOrigins)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: corsAllowCredentials(allowedOrigins),
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderAuthorization,
			"traceparent",
			"tracestate",
			"baggage",
		},
	}))

	app.Get("/health", handlers.Health(minio))

	app.Get("/api/reports/:analysisId/csv", reportsHandler.DownloadCSVPackSigned)

	jwks, err := middleware.NewJWKS(cfg.Auth)
	if err != nil {
		return nil, err
	}
	app.Hooks().OnPreShutdown(func() error {
		jwks.EndBackground()
		return nil
	})

	rl := cfg.RateLimit
	api := app.Group("/api/reports")
	api.Use(middleware.Auth(jwks))
	api.Use(middleware.APIRateLimitGlobal(rl, rateLimitStorage))
	api.Put(
		"/:analysisId",
		middleware.APIRateLimitGenerate(rl, rateLimitStorage),
		reportsHandler.Generate,
	)
	api.Get(
		"/:analysisId/download-url",
		middleware.APIRateLimitDownload(rl, rateLimitStorage),
		reportsHandler.DownloadURL,
	)

	return app, nil
}
