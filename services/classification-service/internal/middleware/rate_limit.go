package middleware

import (
	"csort.ru/classification-service/internal/config"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	redisstorage "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
)

func RateLimit(cfg config.RateLimitConfig, redisClient *redis.Client) fiber.Handler {
	limiterCfg := limiter.Config{
		Max:               cfg.Max,
		Expiration:        cfg.Expiration(),
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage:           redisstorage.NewFromConnection(redisClient),
		KeyGenerator: func(c fiber.Ctx) string {
			return "classification:" + c.IP()
		},
		Next: func(c fiber.Ctx) bool {
			return c.Path() == "/health"
		},
		LimitReached: func(c fiber.Ctx) error {
			return response.Fail(
				c,
				fiber.StatusTooManyRequests,
				"Too many requests. Please try again later.",
				nil,
			)
		},
	}
	return limiter.New(limiterCfg)
}
