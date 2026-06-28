package response

import (
	"encoding/json"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
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

// NewFiberErrorHandler returns a Fiber error handler that logs with the given logger.
func NewFiberErrorHandler(log zerolog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		statusCode := fiber.StatusInternalServerError
		message := "Internal Server Error"
		var details any

		if e, ok := err.(*fiber.Error); ok {
			statusCode = e.Code
			message = e.Message
			if statusCode == fiber.StatusNotFound {
				log.Debug().
					Err(e).
					Int("status", statusCode).
					Str("path", c.Path()).
					Msg("Fiber 404 error caught")
			} else {
				log.Warn().
					Err(e).
					Int("status", statusCode).
					Str("path", c.Path()).
					Msg("Fiber error caught")
			}
		} else {
			details = err.Error()
			log.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error from handler chain")
		}

		c.Status(statusCode).Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Response().SetBody(nil)
		return c.JSON(ErrorResponse{
			Success: false,
			Error: ErrorInfo{
				Code:    statusCode,
				Message: message,
				Details: details,
				Path:    c.Path(),
			},
		})
	}
}
