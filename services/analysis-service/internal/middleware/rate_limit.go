package middleware

import (
	"context"
	"fmt"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

const (
	RateLimitKeyLocals = "rate_limit_key"

	redisCheckTimeout = 5 * time.Second
	decrementTimeout  = 3 * time.Second
)

var rateLimitLog = logger.GetLogger("middleware.rate_limit")

const requestScriptSrc = `
local current = tonumber(redis.call("GET", KEYS[1])) or 0
if current >= tonumber(ARGV[1]) then
	return 0
else
	redis.call("INCR", KEYS[1])
	redis.call("EXPIRE", KEYS[1], ARGV[2])
	return 1
end
`

var requestScript = redis.NewScript(requestScriptSrc)

const decrementScriptSrc = `
local current = tonumber(redis.call("GET", KEYS[1])) or 0
if current <= 0 then
	return 0
end
local nextValue = redis.call("DECR", KEYS[1])
if tonumber(nextValue) <= 0 then
	redis.call("DEL", KEYS[1])
	return 0
end
return tonumber(nextValue)
`

var decrementScript = redis.NewScript(decrementScriptSrc)

type RateLimitConfig struct {
	MaxQueuedRequests int
	RedisClient       *redis.Client
	KeyPrefix         string
	TaskTimeout       time.Duration
	FailOpen          bool // Allow requests on Redis errors
}

func deriveRateLimitKey(identity *auth.Identity) string {
	if identity == nil {
		return ""
	}
	if identity.MaxID != nil {
		return fmt.Sprintf("max:%d", *identity.MaxID)
	}
	if identity.TelegramID != nil {
		return fmt.Sprintf("telegram:%d", *identity.TelegramID)
	}
	return ""
}

func RateLimit(config RateLimitConfig) fiber.Handler {
	if config.RedisClient == nil {
		rateLimitLog.Fatal().Msg("RateLimit: RedisClient is required")
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "limiter:task:"
	}
	if config.TaskTimeout == 0 {
		config.TaskTimeout = 1 * time.Hour
	}
	if config.MaxQueuedRequests <= 0 {
		config.MaxQueuedRequests = 10
	}

	return func(c fiber.Ctx) error {
		identity, ok := c.Locals(UserDataKey).(*auth.Identity)
		if !ok || identity == nil {
			rateLimitLog.Warn().
				Msg("No identity found in context; RateLimit must run after JWTAuth")
			return apierrors.New(fiber.StatusUnauthorized, "Authentication required")
		}

		rateLimitKey := deriveRateLimitKey(identity)
		if rateLimitKey == "" {
			rateLimitLog.Warn().Msg("Identity has no messenger id (telegram or max)")
			return apierrors.New(
				fiber.StatusBadRequest,
				"user has no messenger id (telegram or max)",
			)
		}

		key := config.KeyPrefix + rateLimitKey
		ctx, cancel := context.WithTimeout(c.Context(), redisCheckTimeout)
		defer cancel()

		status, err := requestScript.Run(ctx, config.RedisClient,
			[]string{key},
			config.MaxQueuedRequests,
			int(config.TaskTimeout.Seconds()),
		).Int()
		if err != nil {
			rateLimitLog.Error().
				Err(err).
				Str("rate_limit_key", rateLimitKey).
				Msg("Redis error checking rate limit")
			if config.FailOpen {
				return c.Next()
			}
			return apierrors.New(fiber.StatusServiceUnavailable, "Service temporarily unavailable")
		}

		if status == 0 {
			rateLimitLog.Warn().Str("rate_limit_key", rateLimitKey).Msg("User rate limit exceeded")
			return apierrors.New(
				fiber.StatusTooManyRequests,
				"Too many pending tasks. Please wait for them to finish.",
			)
		}

		c.Locals(RateLimitKeyLocals, key)
		return c.Next()
	}
}

func DecrementRateLimit(
	ctx context.Context,
	redisClient *redis.Client,
	key string,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, decrementTimeout)
	defer cancel()

	_, err := decrementScript.Run(ctx, redisClient, []string{key}).Int()
	if err != nil {
		rateLimitLog.Error().Err(err).Str("key", key).Msg("Failed to decrement rate limit")
		return err
	}

	return nil
}
