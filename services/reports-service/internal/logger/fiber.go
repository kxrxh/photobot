package logger

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type fiberRequestLoggerKey struct{}

func FiberRequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(fiberRequestLoggerKey{}, WithTrace(c.Context(), Logger))
		return c.Next()
	}
}

func HTTP(c fiber.Ctx) zerolog.Logger {
	if v, ok := c.Locals(fiberRequestLoggerKey{}).(zerolog.Logger); ok {
		return v
	}
	return WithTrace(c.Context(), Logger)
}
