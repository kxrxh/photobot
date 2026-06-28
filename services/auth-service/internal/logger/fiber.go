package logger

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type fiberRequestLoggerKey struct{}

// FiberRequestLogger stores a trace-enriched root logger on Fiber locals. Register immediately after OTEL middleware.
func FiberRequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(fiberRequestLoggerKey{}, WithTrace(c.Context(), Logger))
		return c.Next()
	}
}

// HTTP returns the request-scoped logger (trace_id/span_id when OTEL recorded a span). Falls back to WithTrace(c.Context(), Logger).
func HTTP(c fiber.Ctx) zerolog.Logger {
	if v, ok := c.Locals(fiberRequestLoggerKey{}).(zerolog.Logger); ok {
		return v
	}
	return WithTrace(c.Context(), Logger)
}

// Component is HTTP(c) with a component field, replacing per-request GetLogger+WithTrace at call sites.
func Component(c fiber.Ctx, component string) zerolog.Logger {
	return HTTP(c).With().Str("component", component).Logger()
}
