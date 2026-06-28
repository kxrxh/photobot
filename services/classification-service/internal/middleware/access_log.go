package middleware

import (
	"time"

	"csort.ru/classification-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func AccessLog() fiber.Handler {
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

		z := logger.HTTP(c)

		var evt *zerolog.Event
		switch {
		case status >= 500:
			evt = z.Error()
		case status >= 400:
			evt = z.Warn()
		default:
			evt = z.Info()
		}

		evt = evt.Ctx(c.Context()).
			Str("log.type", "http_access").
			Str("http.method", method).
			Str("http.route", routeTpl).
			Str("url.path", path).
			Str("url.full", url).
			Int("http.status_code", status).
			Int64("http.latency_ms", latency.Milliseconds()).
			Str("client.address", ip)

		if chainErr != nil {
			evt = evt.Err(chainErr)
		}

		evt.Msgf("HTTP %s %s -> %d in %s", method, url, status, latency.Round(time.Microsecond))
		return chainErr
	}
}
