package middleware

import (
	"context"
	"slices"
	"strings"
	"time"

	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/transport/response"
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
		if c.Method() == fiber.MethodOptions {
			authMiddlewareLog.Debug().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Msg("OPTIONS request, skipping auth")
			return c.Next()
		}
		authHeader := c.Get(AuthorizationHeader)
		if authHeader == "" {
			authMiddlewareLog.Warn().Str("path", c.Path()).Msg("Missing Authorization header")
			return response.Fail(c, fiber.StatusUnauthorized,
				"Authentication required: Missing Authorization header", nil)
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			authMiddlewareLog.Warn().
				Str("authHeader", authHeader).
				Msg("Invalid Authorization header format")
			return response.Fail(c, fiber.StatusUnauthorized,
				"Invalid Authorization header format. Expected: Bearer <token>", nil)
		}
		token := parts[1]
		ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
		defer cancel()

		validationResp, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			authMiddlewareLog.Warn().
				Err(err).
				Str("path", c.Path()).
				Str("error", err.Error()).
				Msg("Token validation failed")
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)
		}

		if !validationResp.Valid {
			authMiddlewareLog.Warn().Str("path", c.Path()).Msg("Token validation returned invalid")
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid token", nil)
		}
		c.Locals(UserDataKey, validationResp.Identity)
		c.Locals(UserRolesKey, validationResp.Roles)

		return c.Next()
	}
}

func WithRoles(requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRoles, ok := c.Locals(UserRolesKey).([]string)
		if !ok {
			authMiddlewareLog.Error().Msg("Could not determine user roles")
			return response.Fail(c, fiber.StatusForbidden, "Could not determine user roles", nil)
		}

		if len(requiredRoles) == 0 {
			return c.Next()
		}

		for _, requiredRole := range requiredRoles {
			if slices.Contains(userRoles, requiredRole) {
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "Insufficient permissions", nil)
	}
}
