package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/domain"
	"csort.ru/reports-service/internal/logger"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	bearerTokenKey = "bearerToken"
	jwtSubjectKey  = "jwtSubject"
	jwtClaimsKey   = "jwtMapClaims"
)

func NewJWKS(cfg config.AuthConfig) (*keyfunc.JWKS, error) {
	jwksURL := strings.TrimSuffix(cfg.ServiceURL, "/") + "/auth/.well-known/jwks.json"
	opts := keyfunc.Options{
		RefreshInterval: cfg.JWKSRefreshInterval,
		RefreshTimeout:  time.Minute,
		RefreshErrorHandler: func(err error) {
			logger.Logger.Warn().Err(err).Str("url", jwksURL).Msg("jwks background refresh failed")
		},
	}
	return keyfunc.Get(jwksURL, opts)
}

func Auth(jwks *keyfunc.JWKS) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return writeUnauthorized(c, "Unauthorized: Missing token")
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenStr == "" {
			return writeUnauthorized(c, "Unauthorized: Missing token")
		}

		claims, err := verifyToken(c.Context(), tokenStr, jwks)
		if err != nil {
			return writeUnauthorized(c, "Unauthorized: Invalid token")
		}

		c.Locals(bearerTokenKey, tokenStr)
		c.Locals(jwtSubjectKey, subjectFromClaims(claims))
		c.Locals(jwtClaimsKey, claims)
		return c.Next()
	}
}

func BearerTokenFromFiber(c fiber.Ctx) string {
	v, _ := c.Locals(bearerTokenKey).(string)
	return v
}

func JWTMapClaimsFromFiber(c fiber.Ctx) jwt.MapClaims {
	v, _ := c.Locals(jwtClaimsKey).(jwt.MapClaims)
	return v
}

func JWTSubjectFromFiber(c fiber.Ctx) string {
	v, _ := c.Locals(jwtSubjectKey).(string)
	return v
}

func subjectFromClaims(claims jwt.MapClaims) string {
	sub, _ := claims["sub"].(string)
	return strings.TrimSpace(sub)
}

func verifyToken(ctx context.Context, tokenStr string, jwks *keyfunc.JWKS) (jwt.MapClaims, error) {
	claims, err := parseToken(tokenStr, jwks)
	if err == nil {
		return claims, nil
	}
	first := err
	if refreshErr := jwks.Refresh(
		ctx,
		keyfunc.RefreshOptions{IgnoreRateLimit: true},
	); refreshErr != nil {
		return nil, fmt.Errorf("%w: jwks refresh: %w", first, refreshErr)
	}
	claims, err = parseToken(tokenStr, jwks)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func parseToken(tokenStr string, jwks *keyfunc.JWKS) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func writeUnauthorized(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Error: msg})
}
