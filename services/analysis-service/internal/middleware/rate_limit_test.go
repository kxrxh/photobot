package middleware

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newRedisForRateLimitTest(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})
	return client, mr
}

func TestDecrementRateLimit_DoesNotGoNegative(t *testing.T) {
	client, _ := newRedisForRateLimitTest(t)
	ctx := context.Background()
	key := "limiter:task:telegram:123"

	require.NoError(t, DecrementRateLimit(ctx, client, key))
	require.NoError(t, DecrementRateLimit(ctx, client, key))

	value, err := client.Get(ctx, key).Result()
	require.ErrorIs(t, err, redis.Nil)
	require.Empty(t, value)
}

func TestDecrementRateLimit_DecrementsAndDeletesAtZero(t *testing.T) {
	client, mr := newRedisForRateLimitTest(t)
	ctx := context.Background()
	key := "limiter:task:max:42"

	require.NoError(t, client.Set(ctx, key, 2, 0).Err())
	require.NoError(t, DecrementRateLimit(ctx, client, key))
	current, err := mr.Get(key)
	require.NoError(t, err)
	require.Equal(t, "1", current)

	require.NoError(t, DecrementRateLimit(ctx, client, key))
	require.False(t, mr.Exists(key))
}
