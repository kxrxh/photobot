package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	GetAndDel(ctx context.Context, key string) (string, error)
}

type RedisTokenStore struct {
	client *redis.Client
}

var _ TokenStore = &RedisTokenStore{}

func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{client: client}
}

func (r *RedisTokenStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisTokenStore) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisTokenStore) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisTokenStore) GetAndDel(ctx context.Context, key string) (string, error) {
	return r.client.GetDel(ctx, key).Result()
}
