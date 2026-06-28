//go:build integration

package harness

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"

	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/server"
	testingpkg "csort.ru/auth-service/internal/testing"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TestEnv struct {
	DBPool     *pgxpool.Pool
	Redis      *redis.Client
	Server     *server.Server
	HTTPServer *httptest.Server
	BaseURL    string
	Config     *config.Config
	KeysDir    string
}

func SetupTestEnv(ctx context.Context) (*TestEnv, func(), error) {
	logger.InitLogger()

	pool, redisClient, cleanupInfra, err := testingpkg.SetupPostgresAndRedisNoTB(ctx)
	if err != nil {
		return nil, nil, err
	}

	keysDir, err := os.MkdirTemp("", "auth-integration-keys-*")
	if err != nil {
		cleanupInfra()
		return nil, nil, fmt.Errorf("keys temp dir: %w", err)
	}

	km := auth.GetKeyManager()
	if km.GetSigningKey() == nil {
		if err := km.Initialize(
			filepath.Join(keysDir, "private.pem"),
			filepath.Join(keysDir, "public.pem"),
		); err != nil {
			_ = os.RemoveAll(keysDir)
			cleanupInfra()
			return nil, nil, err
		}
	}

	cfg := testingpkg.ConfigFor(pool, redisClient, keysDir)

	srv, err := server.New(cfg, pool, redisClient)
	if err != nil {
		_ = os.RemoveAll(keysDir)
		cleanupInfra()
		return nil, nil, err
	}

	ts := httptest.NewServer(adaptor.FiberApp(srv.App()))

	env := &TestEnv{
		DBPool:     pool,
		Redis:      redisClient,
		Server:     srv,
		HTTPServer: ts,
		BaseURL:    ts.URL,
		Config:     cfg,
		KeysDir:    keysDir,
	}

	cleanup := func() {
		ts.Close()
		_ = srv.Shutdown()
		_ = os.RemoveAll(keysDir)
		cleanupInfra()
	}

	return env, cleanup, nil
}
