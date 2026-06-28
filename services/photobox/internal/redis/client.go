package redis

import (
	"context"
	"fmt"
	"time"

	"csort.ru/coffeebot/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// NewRateLimitClient creates a Redis client for middleware rate limiting.
// If Redis is unavailable, it returns nil so callers can fallback to in-memory storage.
func NewRateLimitClient(cfg config.RedisConfig, log zerolog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		log.Warn().Err(err).Msg("redis unavailable, falling back to in-memory rate limiter")
		return nil
	}

	return client
}
