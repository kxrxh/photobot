package redislimit

import (
	"context"
	"fmt"

	storage "github.com/gofiber/storage/redis/v3"
	redis "github.com/redis/go-redis/v9"
)

func OpenLimiterStorage(ctx context.Context, redisURL string) (*storage.Storage, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis url: %w", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return storage.NewFromConnection(client), nil
}
