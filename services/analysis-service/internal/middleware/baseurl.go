package middleware

import (
	"csort.ru/analysis-service/internal/httputil"
	"csort.ru/analysis-service/internal/logger"
	"github.com/gofiber/fiber/v3"
)

var baseurlLog = logger.GetLogger("middleware.baseurl")

func BaseURL(fallback string) fiber.Handler {
	return func(c fiber.Ctx) error {
		parent := c.Context()
		res := httputil.DerivePublicAPIBaseURL(c, fallback)
		baseurlLog.Debug().
			Str("base_url", res.BaseURL).
			Str("source", res.Source).
			Str("host", res.Host).
			Str("scheme", res.Scheme).
			Str("forwarded_prefix", res.ForwardedPrefix).
			Msg("resolved public API base URL")

		c.SetContext(httputil.WithBaseURL(parent, res.BaseURL))
		return c.Next()
	}
}
