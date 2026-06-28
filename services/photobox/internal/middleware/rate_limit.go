package middleware

import (
	"csort.ru/coffeebot/internal/config"
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
		KeyGenerator: func(c fiber.Ctx) string {
			return "photobox:" + c.IP()
		},
		Next: func(c fiber.Ctx) bool {
			return c.Path() == "/health"
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		},
	}
	if redisClient != nil {
		limiterCfg.Storage = redisstorage.NewFromConnection(redisClient)
	}
	return limiter.New(limiterCfg)
}
