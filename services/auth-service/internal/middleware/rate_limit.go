package middleware

import (
	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	redisstorage "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
)

func AuthRateLimit(redisClient *redis.Client, cfg config.RateLimitConfig) fiber.Handler {
	if redisClient == nil {
		return func(c fiber.Ctx) error { return c.Next() }
	}

	store := redisstorage.NewFromConnection(redisClient)

	return limiter.New(limiter.Config{
		Storage:           store,
		Max:               cfg.Max,
		Expiration:        cfg.Expiration(),
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c fiber.Ctx) string {
			return "auth:" + c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return apperrors.New(
				fiber.StatusTooManyRequests,
				"Too many requests. Please try again later.",
			)
		},
	})
}
