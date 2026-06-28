package apifiber

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"csort.ru/analysis-service/internal/apierrors"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlerJSON(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/x", func(c fiber.Ctx) error {
		return apierrors.New(fiber.StatusUnauthorized, "nope")
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Path    string `json:"path"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.False(t, body.Success)
	assert.Equal(t, fiber.StatusUnauthorized, body.Error.Code)
	assert.Equal(t, "nope", body.Error.Message)
	assert.Equal(t, "/x", body.Error.Path)
}
