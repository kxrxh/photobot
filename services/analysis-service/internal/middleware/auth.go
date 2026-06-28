package middleware

import (
	"context"
	"slices"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/logger"
	"github.com/gofiber/fiber/v3"
)

const (
	AuthorizationHeader = "Authorization"
	UserDataKey         = "user_data"
	UserRolesKey        = "user_roles"
)

var authMiddlewareLog = logger.GetLogger("middleware.auth")

func JWTAuth(authClient *auth.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		return jwtAuthFiber(c, authClient)
	}
}

//nolint:contextcheck // Fiber uses c.Context() for request scope; this is not an http.Handler with *http.Request.
func jwtAuthFiber(c fiber.Ctx, authClient *auth.Client) error {
	if c.Method() == fiber.MethodOptions {
		authMiddlewareLog.Debug().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Msg("OPTIONS request, skipping auth")
		return c.Next()
	}

	hdr := c.Get(AuthorizationHeader)
	if hdr == "" {
		authMiddlewareLog.Warn().
			Str("path", c.Path()).
			Msg("auth rejected: missing authorization header")
		return apierrors.New(
			fiber.StatusUnauthorized,
			"Authentication required: Missing Authorization header",
		)
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(hdr, bearerPrefix) || len(hdr) <= len(bearerPrefix) {
		authMiddlewareLog.Warn().
			Str("path", c.Path()).
			Msg("auth rejected: invalid authorization header format")
		return apierrors.New(
			fiber.StatusUnauthorized,
			"Invalid Authorization header format. Expected: Bearer <token>",
		)
	}
	tok := hdr[len(bearerPrefix):]

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	authStart := time.Now()
	resp, err := authClient.ValidateToken(ctx, tok)
	authDuration := time.Since(authStart)
	authMiddlewareLog.Debug().
		Str("path", c.Path()).
		Dur("auth_duration", authDuration).
		Msg("Token validation timing")
	if err != nil {
		authMiddlewareLog.Warn().
			Err(err).
			Str("path", c.Path()).
			Msg("auth rejected: token validation failed")
		return apierrors.New(fiber.StatusUnauthorized, "Invalid or expired token")
	}

	if !resp.Valid {
		authMiddlewareLog.Warn().Str("path", c.Path()).Msg("auth rejected: token invalid")
		return apierrors.New(fiber.StatusUnauthorized, "Invalid token")
	}
	c.Locals(UserDataKey, resp.Identity)
	c.Locals(UserRolesKey, resp.Roles)

	return c.Next()
}

func WithRoles(requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRoles, ok := c.Locals(UserRolesKey).([]string)
		if !ok {
			authMiddlewareLog.Error().Msg("auth rejected: could not determine roles")
			return apierrors.New(fiber.StatusForbidden, "Could not determine user roles")
		}

		if len(requiredRoles) == 0 {
			return c.Next()
		}

		for _, requiredRole := range requiredRoles {
			if slices.Contains(userRoles, requiredRole) {
				return c.Next()
			}
		}

		return apierrors.New(fiber.StatusForbidden, "Insufficient permissions")
	}
}
