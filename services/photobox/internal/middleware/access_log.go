package middleware

import (
	"context"
	"time"

	"csort.ru/coffeebot/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

// AccessLog logs each HTTP request with structured fields and a readable message.
func AccessLog(ctx context.Context, base zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		chainErr := c.Next()
		if chainErr != nil {
			if err := c.App().ErrorHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		url := c.OriginalURL()
		ip := c.IP()

		routeTpl := path
		if r := c.Route(); r.Path != "" {
			routeTpl = r.Path
		}

		logCtx := ContextWithRequestSpan(ctx, c.Context())
		traced := logger.WithTrace(logCtx, base)

		var evt *zerolog.Event
		switch {
		case status >= 500:
			evt = traced.Error()
		case status >= 400:
			evt = traced.Warn()
		default:
			evt = traced.Info()
		}

		evt = evt.Ctx(logCtx).
			Str("log.type", "http_access").
			Str("http.method", method).
			Str("http.route", routeTpl).
			Str("url.path", path).
			Str("url.full", url).
			Int("http.status_code", status).
			Int64("http.latency_ms", latency.Milliseconds()).
			Int("http.response.body.size", len(c.Response().Body())).
			Str("client.address", ip)

		if chainErr != nil {
			evt = evt.Err(chainErr)
		}

		evt.Msgf("HTTP %s %s -> %d in %s", method, url, status, latency.Round(time.Microsecond))
		return chainErr
	}
}
