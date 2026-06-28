package middleware

import (
	"fmt"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/api/common"
	apierr "csort.ru/analysis-service/internal/apierrors/fiber"
	"csort.ru/analysis-service/internal/logger"
	fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func AccessLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		if c.Method() == fiber.MethodGet && c.Path() == "/api/v1/ws" {
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
		if identity, ok := c.Locals(UserDataKey).(*auth.Identity); ok && identity != nil {
			var parts []string
			if identity.MaxID != nil {
				parts = append(parts, fmt.Sprintf("max:%d", *identity.MaxID))
			}
			if identity.TelegramID != nil {
				parts = append(parts, fmt.Sprintf("telegram:%d", *identity.TelegramID))
			}
			if len(parts) > 0 {
				zc = zc.Str("user_id", strings.Join(parts, ","))
			}
		}
		if msg, ok := c.Locals(apierr.ErrorMessageLocalsKey).(string); ok && msg != "" {
			zc = zc.Str(apierr.ErrorMessageLocalsKey, msg)
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
			if upstream := common.Extract(chainErr); upstream != nil {
				evt = evt.
					Str("upstream.peer.service", upstream.PeerService).
					Int("upstream.http.response.status_code", upstream.StatusCode).
					Str("upstream.url.full", upstream.Endpoint).
					Str("upstream.http.request.method", upstream.Method)
				if preview := upstream.BodyPreview(common.DefaultBodyPreviewLen); preview != "" {
					evt = evt.Str("upstream.http.response.body.preview", preview)
				}
			}
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
