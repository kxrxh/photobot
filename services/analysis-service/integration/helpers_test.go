//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func analysisServiceRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func readRequestsInitSchema(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join(analysisServiceRoot(t), "databases", "requests", "schema.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	return b
}

func runPostgresIT(t *testing.T, ctx context.Context, database string) *postgres.PostgresContainer {
	t.Helper()
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(database),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	return pgContainer
}

func readKalibrSchema(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join(analysisServiceRoot(t), "databases", "kalibr", "schema.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read kalibr schema: %v", err)
	}
	return b
}
