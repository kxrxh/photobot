package httputil

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type contextKey struct{}

var baseURLKey = contextKey{}

type PublicAPIBaseResult struct {
	BaseURL         string
	Source          string // forwarded | direct_host | fallback
	Scheme          string // http or https (empty when Source is fallback)
	Host            string // effective host used in BaseURL (empty when Source is fallback)
	ForwardedPrefix string // raw X-Forwarded-Prefix before normalization (if any)
}

func DerivePublicAPIBaseURL(c fiber.Ctx, fallback string) PublicAPIBaseResult {
	proto := c.Get("X-Forwarded-Proto")
	if proto != "http" && proto != "https" {
		proto = "https"
	}
	return publicAPIBaseFromHeaders(
		c.Get("X-Forwarded-Host"),
		c.Host(),
		proto,
		c.Get("X-Forwarded-Prefix"),
		fallback,
	)
}

func publicAPIBaseFromHeaders(
	forwardedHost, directHost, proto, rawPrefix, fallback string,
) PublicAPIBaseResult {
	host := forwardedHost
	source := "direct_host"
	if host != "" {
		source = "forwarded"
	} else {
		host = directHost
	}
	if host == "" {
		return PublicAPIBaseResult{BaseURL: fallback, Source: "fallback"}
	}
	prefix := rawPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	base := proto + "://" + host
	if prefix != "" {
		base += strings.TrimSuffix(prefix, "/")
	}
	return PublicAPIBaseResult{
		BaseURL:         base + "/api/v1",
		Source:          source,
		Scheme:          proto,
		Host:            host,
		ForwardedPrefix: rawPrefix,
	}
}

func BaseURLFromRequest(c fiber.Ctx, fallback string) string {
	return DerivePublicAPIBaseURL(c, fallback).BaseURL
}

func WithBaseURL(ctx context.Context, baseURL string) context.Context {
	return context.WithValue(ctx, baseURLKey, baseURL)
}

func BaseURLFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(baseURLKey)
	if s, ok := v.(string); ok && s != "" {
		return s, true
	}
	return "", false
}
