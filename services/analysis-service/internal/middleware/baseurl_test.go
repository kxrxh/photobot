package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"csort.ru/analysis-service/internal/httputil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestBaseURL(t *testing.T) {
	app := fiber.New()
	app.Use(BaseURL("https://fallback/api/v1"))
	app.Get("/", func(c fiber.Ctx) error {
		base, ok := httputil.BaseURLFromContext(c.Context())
		if !ok {
			return c.SendString("missing")
		}
		return c.SendString(base)
	})

	t.Run("sets base URL from X-Forwarded headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Host = "example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "api.example.com")
		req.Header.Set("X-Forwarded-Prefix", "/analysis")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "https://api.example.com/analysis/api/v1", string(b))
	})
}
