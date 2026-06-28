package services

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
)

const RedisServiceName = "redis"

type RedisService struct {
	client    *redis.Client
	closeOnce sync.Once
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{client: client}
}

func (s *RedisService) Start(ctx context.Context) error {
	if s.client == nil {
		return errors.New("redis client is nil")
	}
	return s.client.Ping(ctx).Err()
}

func (s *RedisService) String() string { return RedisServiceName }

func (s *RedisService) State(ctx context.Context) (string, error) {
	if s.client == nil {
		return "unavailable", nil
	}
	if err := s.client.Ping(ctx).Err(); err != nil {
		return "disconnected", nil
	}
	return "connected", nil
}

func (s *RedisService) Terminate(context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		if s.client != nil {
			err = s.client.Close()
		}
	})
	return err
}
