//go:build integration

package harness

import (
	"context"
	"net/http/httptest"
	"testing"

	"csort.ru/coffeebot/internal/server"
	testutil "csort.ru/coffeebot/internal/testing"

	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestEnv struct {
	DBPool       *pgxpool.Pool
	Server       *server.Server
	HTTPServer   *httptest.Server
	BaseURL      string
	GetToken func(userID int32, roles []string) string
}

func NewTestEnv(ctx context.Context, t *testing.T) *TestEnv {
	t.Helper()

	srv, deps, cleanup := testutil.NewTestApp(ctx, t)
	ts := httptest.NewServer(adaptor.FiberApp(srv.App()))

	env := &TestEnv{
		DBPool:     deps.DBPool,
		Server:     srv,
		HTTPServer: ts,
		BaseURL:    ts.URL,
		GetToken:   deps.GetToken,
	}

	t.Cleanup(func() {
		ts.Close()
		cleanup()
	})

	return env
}

func (e *TestEnv) Bearer(rawToken string) string {
	return "Bearer " + rawToken
}

func (e *TestEnv) AuthHeader(rawToken string) map[string]string {
	return map[string]string{"Authorization": e.Bearer(rawToken)}
}
