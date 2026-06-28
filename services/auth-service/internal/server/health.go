package server

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const healthProbeTimeout = 2 * time.Second

type healthCache struct {
	mu    sync.RWMutex
	resp  fiber.Map
	until time.Time
}

// NewHealthHandler probes DB and Redis (200 vs 503). With cacheTTL > 0, successful probes are cached briefly.
func NewHealthHandler(
	db *pgxpool.Pool,
	redisClient *redis.Client,
	cacheTTL time.Duration,
) fiber.Handler {
	var cache healthCache
	return func(c fiber.Ctx) error {
		if cacheTTL > 0 {
			cache.mu.RLock()
			if time.Now().Before(cache.until) && cache.resp != nil {
				resp := fiber.Map{
					"service": cache.resp["service"],
					"time":    time.Now().Format(time.RFC3339),
					"checks":  cache.resp["checks"],
					"status":  "healthy",
				}
				cache.mu.RUnlock()
				return c.JSON(resp)
			}
			cache.mu.RUnlock()
		}

		ctx, cancel := context.WithTimeout(c.Context(), healthProbeTimeout)
		defer cancel()

		checks := make(map[string]string)
		allOk := true

		if err := db.Ping(ctx); err != nil {
			checks["database"] = err.Error()
			allOk = false
		} else {
			checks["database"] = "ok"
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			allOk = false
		} else {
			checks["redis"] = "ok"
		}

		resp := fiber.Map{
			"service": "auth-service",
			"time":    time.Now().Format(time.RFC3339),
			"checks":  checks,
		}
		if allOk {
			resp["status"] = "healthy"
			if cacheTTL > 0 {
				cache.mu.Lock()
				cache.resp = fiber.Map{"service": "auth-service", "checks": checks}
				cache.until = time.Now().Add(cacheTTL)
				cache.mu.Unlock()
			}
			return c.JSON(resp)
		}
		resp["status"] = "unhealthy"
		return c.Status(fiber.StatusServiceUnavailable).JSON(resp)
	}
}
