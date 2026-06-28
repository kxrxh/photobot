//go:build integration

package integration

import (
	"context"
	"testing"

	repo_requests "csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/requests"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestRequestsServiceList_EmptyDB(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	pgContainer := runPostgresIT(t, ctx, "requests_it")
	t.Cleanup(func() {
		_ = pgContainer.Terminate(context.WithoutCancel(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, string(readRequestsInitSchema(t)))
	require.NoError(t, err)

	q := repo_requests.New(pool)
	svc := requests.NewRequestsService(requests.RequestsServiceParams{
		Repo: q,
		Pool: pool,
	})

	out, err := svc.List(ctx, requests.GetRequestsRequest{
		UserID: "user-1",
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Empty(t, out)
}
