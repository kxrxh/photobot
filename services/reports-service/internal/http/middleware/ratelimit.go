package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// APIRateLimitGlobal applies a per-user quota to every /api/reports request.
func APIRateLimitGlobal(cfg config.RateLimitConfig, storage fiber.Storage) fiber.Handler {
	return newReportsRateLimit(storage, cfg.GlobalMax, cfg.GlobalWindow, "global")
}

// APIRateLimitGenerate applies an additional per-user quota only to report generation (PUT).
func APIRateLimitGenerate(cfg config.RateLimitConfig, storage fiber.Storage) fiber.Handler {
	return newReportsRateLimit(storage, cfg.GenerateMax, cfg.GenerateWindow, "gen")
}

// APIRateLimitDownload applies an additional per-user quota only to download-url (GET).
func APIRateLimitDownload(cfg config.RateLimitConfig, storage fiber.Storage) fiber.Handler {
	return newReportsRateLimit(storage, cfg.DownloadMax, cfg.DownloadWindow, "dl")
}

func newReportsRateLimit(
	storage fiber.Storage,
	max int,
	window time.Duration,
	bucket string,
) fiber.Handler {
	if max <= 0 || window <= 0 {
		return func(c fiber.Ctx) error { return c.Next() }
	}
	keyGen := func(c fiber.Ctx) string {
		return fmt.Sprintf("reports:rl:%s:%s", bucket, rateLimitSubject(c))
	}
	return limiter.New(limiter.Config{
		Storage:           storage,
		LimiterMiddleware: limiter.SlidingWindow{},
		Max:               max,
		Expiration:        window,
		KeyGenerator:      keyGen,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				JSON(domain.ErrorResponse{Error: "Rate limit exceeded. Try again later."})
		},
	})
}

func rateLimitSubject(c fiber.Ctx) string {
	sub := JWTSubjectFromFiber(c)
	if sub == "" {
		return rateLimitSubjectFallback(c)
	}
	return sub
}

func rateLimitSubjectFallback(c fiber.Ctx) string {
	tok := BearerTokenFromFiber(c)
	if tok == "" {
		return "ip:" + c.IP()
	}
	h := sha256.Sum256([]byte(tok))
	return "tok:" + strings.ToLower(hex.EncodeToString(h[:8]))
}
