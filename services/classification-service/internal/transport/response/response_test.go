package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"csort.ru/classification-service/internal/httperr"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOK_WrapsPayload(t *testing.T) {
	app := fiber.New()
	app.Get("/data", func(c fiber.Ctx) error {
		return OK(c, fiber.Map{"id": 1, "name": "test"})
	})

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var body struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Success)
	var inner map[string]any
	require.NoError(t, json.Unmarshal(body.Result, &inner))
	assert.Equal(t, float64(1), inner["id"])
	assert.Equal(t, "test", inner["name"])
}

func TestOK_NilResult(t *testing.T) {
	app := fiber.New()
	app.Get("/empty", func(c fiber.Ctx) error {
		return OK(c, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/empty", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Success bool        `json:"success"`
		Result  interface{} `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Success)
	assert.Nil(t, body.Result)
}

func TestFiberErrorHandler_NilUnused(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: FiberErrorHandler})
	app.Get("/ok", func(c fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestFiberErrorHandler_HTTPError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: FiberErrorHandler})
	app.Get("/bad", func(c fiber.Ctx) error {
		return httperr.New(400, "Invalid request")
	})
	app.Get("/conflict", func(c fiber.Ctx) error {
		return httperr.New(409, "Resource already exists")
	})
	app.Get("/internal", func(c fiber.Ctx) error {
		return httperr.New(500, "Database error")
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantMsg    string
	}{
		{"400 Bad Request", "/bad", 400, "Invalid request"},
		{"409 Conflict", "/conflict", 409, "Resource already exists"},
		{"500 Internal", "/internal", 500, "Database error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

			var body struct {
				Success bool `json:"success"`
				Error   struct {
					Message string `json:"message"`
					Code    int    `json:"code"`
				} `json:"error"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.False(t, body.Success)
			assert.Equal(t, tt.wantMsg, body.Error.Message)
			assert.Equal(t, tt.wantStatus, body.Error.Code)
		})
	}
}

func TestFiberErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: FiberErrorHandler})
	app.Get("/404", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "Not Found")
	})

	req := httptest.NewRequest(http.MethodGet, "/404", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.False(t, body.Success)
	assert.Equal(t, "Not Found", body.Error.Message)
}

func TestFiberErrorHandler_PlainError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: FiberErrorHandler})
	app.Get("/err", func(c fiber.Ctx) error {
		return assert.AnError
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.False(t, body.Success)
	assert.Equal(t, "Internal Server Error", body.Error.Message)
}
