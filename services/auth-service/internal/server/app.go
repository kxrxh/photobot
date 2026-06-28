package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/middleware"
	"csort.ru/auth-service/internal/server/routes"
	"csort.ru/auth-service/internal/transport/response"
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
	app *fiber.App
}

// New creates a new server instance
func New(cfg *config.Config, dbPool *pgxpool.Pool, redisClient *redis.Client) (*Server, error) {
	if dbPool == nil {
		return nil, errors.New("dbPool is nil")
	}
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}
	if redisClient == nil {
		return nil, errors.New("redisClient is nil")
	}

	app := fiber.New(fiber.Config{
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ErrorHandler: response.FiberErrorHandler,
		TrustProxy:   true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: []string{
				"172.16.0.0/12",
				"192.168.0.0/16",
				"127.0.0.1",
			},
		},
	})

	setupMiddleware(app, cfg, redisClient, cfg.Server.Port)

	dbQueries := database.New(dbPool)

	services := initializeServices(dbQueries, dbPool, cfg, redisClient)
	handlers := initializeHandlers(services)
	healthCacheTTL := time.Duration(cfg.Server.HealthCacheTTLSec) * time.Second
	healthHandler := NewHealthHandler(dbPool, redisClient, healthCacheTTL)
	apiRoutes := routes.Define(&handlers, healthHandler)
	registerRoutes(app, apiRoutes)

	return &Server{
		app: app,
	}, nil
}

func setupMiddleware(app *fiber.App, cfg *config.Config, redisClient *redis.Client, httpPort int) {
	app.Use(fibotel.Middleware(
		fibotel.WithoutMetrics(true),
		fibotel.WithPort(httpPort),
	))

	app.Use(logger.FiberRequestLogger())

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
	allowedOrigins := parseAllowedOrigins(cfg.Server.CORSAllowOrigins)
	allowCredentials := len(allowedOrigins) > 0 &&
		(len(allowedOrigins) != 1 || allowedOrigins[0] != "*")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: allowCredentials,
		AllowMethods: []string{
			"GET",
			"POST",
			"HEAD",
			"PUT",
			"DELETE",
			"PATCH",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Init-Data",
			"X-Messenger-Platform",
			"X-Bot-Name",
			"X-Grant-Type",
		},
	}))

	app.Use(middleware.AuthRateLimit(redisClient, cfg.RateLimit))
}

func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"*"}
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		out = append(out, origin)
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func (s *Server) Start(port int) error {
	return s.app.Listen(fmt.Sprintf(":%d", port))
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

func (s *Server) App() *fiber.App {
	return s.app
}
