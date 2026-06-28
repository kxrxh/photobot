package testing

import (
	"context"

	"csort.ru/coffeebot/internal/config"
	"csort.ru/coffeebot/internal/logger"
	"csort.ru/coffeebot/internal/minio"
	"csort.ru/coffeebot/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TestAppDeps struct {
	DBPool      *pgxpool.Pool
	MinioClient *minio.Client
	Identity    *FakeIdentityClient
	GetToken    func(userID int32, roles []string) string
}

func NewTestApp(ctx context.Context, t TB) (*server.Server, *TestAppDeps, func()) {
	t.Helper()

	minioClient := sharedMinio
	if sharedPGConn == "" || minioClient == nil {
		t.Fatalf("shared test infra not started; call StartSharedInfra from TestMain")
	}

	pool, err := newTestPool(ctx)
	if err != nil {
		t.Fatalf("postgres pool: %v", err)
	}
	if err := ResetSharedDB(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("reset shared db: %v", err)
	}

	identityFake := NewFakeIdentityClient(DefaultIntegrationIdentityTokens())

	srv, err := server.New(
		ctx,
		pool,
		minioClient,
		identityFake,
		8080,
		config.RateLimitConfig{Max: 120, ExpirationSeconds: 60},
		nil,
		logger.Logger,
	)
	if err != nil {
		pool.Close()
		t.Fatalf("failed to create server: %v", err)
	}

	getToken := func(userID int32, roles []string) string {
		_, _ = userID, roles
		return "Bearer " + IntegrationTokenAdmin
	}

	deps := &TestAppDeps{
		DBPool:      pool,
		MinioClient: minioClient,
		Identity:    identityFake,
		GetToken:    getToken,
	}

	cleanup := func() {
		_ = srv.Shutdown()
	}

	return srv, deps, cleanup
}
