package middleware

import (
	"strings"
	"time"

	"csort.ru/reports-service/internal/logger"

	fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog"
)

// AccessLog emits one structured line per HTTP request (method, path, status, latency, subject when authenticated).
// Skips OPTIONS and GET /health to avoid noise from probes and CORS preflight.
func AccessLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		if c.Method() == fiber.MethodGet && c.Path() == "/health" {
			return c.Next()
		}

		start := time.Now()
		chainErr := c.Next()
		if chainErr != nil {
			if err := c.App().ErrorHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		latency := time.Since(start)
		status := c.Response().StatusCode()

		z := logger.HTTP(c)
		zc := z.With()
		if sub := strings.TrimSpace(JWTSubjectFromFiber(c)); sub != "" {
			zc = zc.Str("jwt_subject", sub)
		}
		if rid := requestid.FromContext(c); rid != "" {
			zc = zc.Str("request_id", rid)
		}
		base := zc.Logger()

		var evt *zerolog.Event
		switch {
		case status >= 500:
			evt = base.Error()
		case status >= 400:
			evt = base.Warn()
		default:
			evt = base.Info()
		}

		routeTpl := c.Path()
		if r := c.Route(); r.Path != "" {
			routeTpl = r.Path
		}

		latencyRounded := latency.Round(time.Microsecond)
		evt = evt.Ctx(c.Context()).
			Str("log.type", "http_access").
			Str(fiberzerolog.FieldMethod, c.Method()).
			Str(fiberzerolog.FieldPath, c.Path()).
			Str(fiberzerolog.FieldRoute, routeTpl).
			Int(fiberzerolog.FieldStatus, status).
			Str(fiberzerolog.FieldIP, c.IP()).
			Str(fiberzerolog.FieldLatency, latencyRounded.String()).
			Int64("http.latency_ms", latency.Milliseconds()).
			Int(fiberzerolog.FieldBytesSent, len(c.Response().Body())).
			Str("url.full", c.OriginalURL())

		if chainErr != nil {
			evt = evt.Err(chainErr)
		}

		evt.Msgf(
			"HTTP %s %s -> %d in %s",
			c.Method(),
			c.OriginalURL(),
			status,
			latencyRounded,
		)
		return chainErr
	}
}
