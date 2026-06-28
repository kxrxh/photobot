package response

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	apifiber "csort.ru/analysis-service/internal/apierrors/fiber"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

type marshalFailPayload struct {
	Fn func()
}

func TestSuccessHelpers(t *testing.T) {
	app := fiber.New()

	cases := []struct {
		name       string
		path       string
		handler    func(c fiber.Ctx) error
		statusCode int
		expected   string
	}{
		{
			name:       "OK wraps payload",
			path:       "/ok",
			handler:    func(c fiber.Ctx) error { return OK(c, fiber.Map{"foo": "bar"}) },
			statusCode: fiber.StatusOK,
			expected:   `{"foo":"bar"}`,
		},
		{
			name:       "Accepted wraps payload",
			path:       "/accepted",
			handler:    func(c fiber.Ctx) error { return Accepted(c, fiber.Map{"id": 42}) },
			statusCode: fiber.StatusAccepted,
			expected:   `{"id":42}`,
		},
		{
			name:       "JSON wraps payload",
			path:       "/json",
			handler:    func(c fiber.Ctx) error { return JSON(c, fiber.StatusAccepted, fiber.Map{"queued": true}) },
			statusCode: fiber.StatusAccepted,
			expected:   `{"queued":true}`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app.Get(tc.path, tc.handler)

			req := httptest.NewRequest(fiber.MethodGet, tc.path, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.statusCode, resp.StatusCode)
			assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), "application/json")

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			var res struct {
				Success bool            `json:"success"`
				Result  json.RawMessage `json:"result"`
			}
			err = json.Unmarshal(body, &res)
			assert.NoError(t, err)
			assert.True(t, res.Success)
			assert.JSONEq(t, tc.expected, string(res.Result))
		})
	}
}

func TestNoContentHelpers(t *testing.T) {
	app := fiber.New()

	app.Get("/no-content", NoContent)
	app.Get("/json-no-content", func(c fiber.Ctx) error {
		return JSON(c, fiber.StatusNoContent, fiber.Map{"ignored": true})
	})

	for _, path := range []string{"/no-content", "/json-no-content"} {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil))
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Empty(t, body)
	}
}

func TestJSONMarshalFailureReturnsAPIErrorEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apifiber.ErrorHandler})
	app.Get("/marshal-fail", func(c fiber.Ctx) error {
		return JSON(c, fiber.StatusOK, marshalFailPayload{
			Fn: func() {},
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/marshal-fail", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), "application/json")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var envelope struct {
		Success bool `json:"success"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Path    string `json:"path"`
		} `json:"error"`
	}
	err = json.Unmarshal(body, &envelope)
	assert.NoError(t, err)
	assert.False(t, envelope.Success)
	assert.Equal(t, fiber.StatusInternalServerError, envelope.Error.Code)
	assert.Equal(t, "Failed to encode response", envelope.Error.Message)
	assert.Equal(t, "/marshal-fail", envelope.Error.Path)
}
