package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPinger struct {
	pingErr error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.pingErr
}

func newHealthHandlerFromPinger(p Pinger) *HealthHandler {
	return &HealthHandler{pinger: p}
}

func TestNewHealthHandler_Healthy(t *testing.T) {
	app := fiber.New()
	handler := newHealthHandlerFromPinger(&mockPinger{pingErr: nil})
	app.Get("/health", handler.GetHealth)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestNewHealthHandler_Unhealthy(t *testing.T) {
	app := fiber.New()
	handler := newHealthHandlerFromPinger(&mockPinger{pingErr: assert.AnError})
	app.Get("/health", handler.GetHealth)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}
