package service

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGXPool closes the pgx pool on Fiber shutdown (Start is a no-op).
type PGXPool struct {
	pool *pgxpool.Pool
}

var _ fiber.Service = (*PGXPool)(nil)

func NewPGXPool(pool *pgxpool.Pool) *PGXPool {
	return &PGXPool{pool: pool}
}

func (s *PGXPool) String() string {
	return "PostgreSQL"
}

func (s *PGXPool) Start(_ context.Context) error {
	return nil
}

func (s *PGXPool) State(ctx context.Context) (string, error) {
	if s.pool == nil {
		return "stopped", nil
	}
	if err := s.pool.Ping(ctx); err != nil {
		return "unreachable", err
	}
	return "connected", nil
}

func (s *PGXPool) Terminate(_ context.Context) error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}
