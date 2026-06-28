package analysis

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	cleanupLockKey       = "analysis:cleanup:lock"
	cleanupCheckInterval = 1 * time.Hour
)

func generateLockID(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	sum := h.Sum64()
	if sum > 1<<63-1 {
		sum %= (1 << 63)
	}
	return int64(sum)
}

type CleanupLock struct {
	dbPool *pgxpool.Pool
	lockID int64
	logger zerolog.Logger
}

func NewCleanupLock(dbPool *pgxpool.Pool, log zerolog.Logger) *CleanupLock {
	return &CleanupLock{
		dbPool: dbPool,
		lockID: generateLockID(cleanupLockKey),
		logger: log,
	}
}

func (l *CleanupLock) StartPeriodicCleanup(
	ctx context.Context,
	cleanupFunc func(context.Context) error,
) {
	if l.dbPool == nil {
		l.logger.Warn().
			Msg("Database pool not available, cleanup will run without distributed locking (not recommended for multiple instances)")
	}

	go func() {
		ticker := time.NewTicker(cleanupCheckInterval)
		defer ticker.Stop()

		l.tryRunCleanup(ctx, cleanupFunc)

		for {
			select {
			case <-ctx.Done():
				l.logger.Info().Msg("Context cancelled, stopping periodic cleanup")
				return
			case <-ticker.C:
				l.tryRunCleanup(ctx, cleanupFunc)
			}
		}
	}()
}

func (l *CleanupLock) tryRunCleanup(ctx context.Context, cleanupFunc func(context.Context) error) {
	if ctx.Err() != nil {
		return
	}

	if l.dbPool == nil {
		if err := cleanupFunc(ctx); err != nil {
			l.logger.Error().Err(err).Msg("Failed to run cleanup")
		} else {
			l.logger.Info().Msg("Successfully ran cleanup")
		}
		return
	}

	conn, err := l.dbPool.Acquire(ctx)
	if err != nil {
		l.logger.Error().Err(err).Msg("Failed to acquire database connection for cleanup lock")
		return
	}
	defer conn.Release()

	acquired, err := l.acquireLock(ctx, conn)
	if err != nil {
		l.logger.Error().Err(err).Msg("Failed to acquire cleanup lock")
		return
	}

	if !acquired {
		l.logger.Debug().Msg("Cleanup lock already held by another instance, skipping")
		return
	}

	l.logger.Info().Msg("Acquired cleanup lock, starting cleanup")

	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		if err := l.releaseLock(releaseCtx, conn); err != nil {
			l.logger.Warn().Err(err).Msg("Failed to release cleanup lock")
		}
	}()

	if err := cleanupFunc(ctx); err != nil {
		l.logger.Error().Err(err).Msg("Failed to run cleanup")
	} else {
		l.logger.Info().Msg("Successfully ran cleanup")
	}
}

func (l *CleanupLock) acquireLock(ctx context.Context, conn *pgxpool.Conn) (bool, error) {
	var acquired bool
	err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", l.lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("failed to acquire cleanup lock: %w", err)
	}
	return acquired, nil
}

func (l *CleanupLock) releaseLock(ctx context.Context, conn *pgxpool.Conn) error {
	var released bool
	err := conn.QueryRow(ctx, "SELECT pg_advisory_unlock($1)", l.lockID).Scan(&released)
	if err != nil {
		return fmt.Errorf("failed to release cleanup lock: %w", err)
	}
	if !released {
		l.logger.Debug().Msg("Lock was not held by this session during release attempt")
	}
	return nil
}
