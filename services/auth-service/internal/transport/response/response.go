package response

import (
	"encoding/json"
	"net/http"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/logger"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

type SuccessResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
}

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Path    string `json:"path"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorInfo `json:"error"`
}

func OK(c fiber.Ctx, payload any) error {
	return JSON(c, fiber.StatusOK, payload)
}

func Created(c fiber.Ctx, payload any) error {
	return JSON(c, fiber.StatusCreated, payload)
}

func JSON(c fiber.Ctx, statusCode int, payload any) error {
	if statusCode == fiber.StatusNoContent {
		return NoContent(c)
	}

	encoded, err := sonic.Marshal(payload)
	if err != nil {
		return Fail(c, fiber.StatusInternalServerError, "Failed to encode response", nil)
	}

	return c.Status(statusCode).JSON(SuccessResponse{
		Success: true,
		Result:  json.RawMessage(encoded),
	})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c fiber.Ctx, statusCode int, message string, details any) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    statusCode,
			Message: message,
			Details: details,
			Path:    c.Path(),
		},
	})
}

func FiberErrorHandler(c fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	message := http.StatusText(statusCode)
	var details any

	if httpErr, ok := apperrors.FromError(err); ok {
		statusCode = httpErr.Code
		message = httpErr.Message
		if httpErr.Err != nil && statusCode < fiber.StatusInternalServerError {
			details = httpErr.Err.Error()
		}
	} else if fiberErr, ok := err.(*fiber.Error); ok {
		statusCode = fiberErr.Code
		message = fiberErr.Message
	}

	log := logger.Component(c, "transport.response.error_handler")
	logEvent := log.Warn()
	if statusCode >= fiber.StatusInternalServerError {
		logEvent = log.Error()
	}
	logEvent.
		Ctx(c.Context()).
		Err(err).
		Str("log.type", "http_error").
		Str("http.method", c.Method()).
		Str("url.path", c.Path()).
		Int("http.status_code", statusCode).
		Msg("request failed")

	return Fail(c, statusCode, message, details)
}
