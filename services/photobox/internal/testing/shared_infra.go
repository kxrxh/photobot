package testing

import (
	"context"
	"errors"
	"sync"

	"csort.ru/coffeebot/internal/minio"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	sharedOnce     sync.Once
	sharedPGConn   string
	sharedMinio    *minio.Client
	sharedCleanup  func()
	sharedStartErr error
)

func StartSharedInfra(ctx context.Context) error {
	sharedOnce.Do(func() {
		var (
			pgConn       string
			minioClient  *minio.Client
			pgCleanup    func()
			minioCleanup func()
			pgErr        error
			minioErr     error
			wg           sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			pgConn, pgCleanup, pgErr = startPostgres(ctx)
		}()
		go func() {
			defer wg.Done()
			minioClient, minioCleanup, minioErr = startMinIO(ctx)
		}()
		wg.Wait()

		if pgErr != nil {
			sharedStartErr = pgErr
			if minioCleanup != nil {
				minioCleanup()
			}
			return
		}
		if minioErr != nil {
			sharedStartErr = minioErr
			pgCleanup()
			return
		}

		sharedPGConn = pgConn
		sharedMinio = minioClient
		sharedCleanup = func() {
			pgCleanup()
			minioCleanup()
		}
	})
	return sharedStartErr
}

func StopSharedInfra() {
	if sharedCleanup != nil {
		sharedCleanup()
		sharedCleanup = nil
	}
	sharedPGConn = ""
	sharedMinio = nil
}

func newTestPool(ctx context.Context) (*pgxpool.Pool, error) {
	if sharedPGConn == "" {
		return nil, errors.New("shared postgres not started")
	}
	pool, err := pgxpool.New(ctx, sharedPGConn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

const truncateTablesSQL = `
TRUNCATE TABLE
	pending_weed_analyses,
	pending_weed_images,
	pending_weed_stats,
	pending_weeds,
	weed_notes,
	weed_analyses,
	weed_images,
	weed_stats,
	catalog_proposals,
	weeds
RESTART IDENTITY CASCADE
`

func ResetSharedDB(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("db pool is nil")
	}
	_, err := pool.Exec(ctx, truncateTablesSQL)
	return err
}
