package httputil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestBaseURLFromRequest(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		base := BaseURLFromRequest(c, "https://fallback.example/api/v1")
		return c.SendString(base)
	})

	t.Run("derives from X-Forwarded headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "api.example.com")
		req.Header.Set("X-Forwarded-Prefix", "/analysis")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body := readBody(t, resp)
		assert.Equal(t, "https://api.example.com/analysis/api/v1", body)
	})

	t.Run("adds leading slash to prefix when missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "host")
		req.Header.Set("X-Forwarded-Prefix", "analysis")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		body := readBody(t, resp)
		assert.Equal(t, "https://host/analysis/api/v1", body)
	})

	t.Run("uses Host when X-Forwarded-Host absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "direct.example.com"
		req.Header.Set("X-Forwarded-Proto", "https")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		body := readBody(t, resp)
		assert.Equal(t, "https://direct.example.com/api/v1", body)
	})

	t.Run("no prefix yields base plus api/v1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "app.local"
		req.Header.Set("X-Forwarded-Proto", "http")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		body := readBody(t, resp)
		assert.Equal(t, "http://app.local/api/v1", body)
	})
}

func TestDerivePublicAPIBaseURL_Source(t *testing.T) {
	app := fiber.New()
	var got PublicAPIBaseResult
	app.Get("/fwd", func(c fiber.Ctx) error {
		got = DerivePublicAPIBaseURL(c, "https://fb/api/v1")
		return c.SendStatus(200)
	})

	t.Run("forwarded", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fwd", nil)
		req.Header.Set("X-Forwarded-Host", "api.example.com")
		req.Header.Set("X-Forwarded-Proto", "https")
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, "forwarded", got.Source)
		assert.Equal(t, "https://api.example.com/api/v1", got.BaseURL)
	})

	t.Run("direct_host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fwd", nil)
		req.Host = "direct.test"
		req.Header.Set("X-Forwarded-Proto", "http")
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, "direct_host", got.Source)
		assert.Equal(t, "http://direct.test/api/v1", got.BaseURL)
	})
}

func TestPublicAPIBaseFromHeaders_fallback(t *testing.T) {
	res := publicAPIBaseFromHeaders("", "", "https", "", "https://fallback.example/api/v1")
	assert.Equal(t, "fallback", res.Source)
	assert.Equal(t, "https://fallback.example/api/v1", res.BaseURL)
}

func TestWithBaseURL_BaseURLFromContext(t *testing.T) {
	ctx := context.Background()

	_, ok := BaseURLFromContext(ctx)
	assert.False(t, ok)

	ctx = WithBaseURL(ctx, "https://test/api/v1")
	got, ok := BaseURLFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "https://test/api/v1", got)
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return string(b)
}
