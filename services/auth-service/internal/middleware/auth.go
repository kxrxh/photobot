package middleware

import (
	"slices"
	"strings"

	"csort.ru/auth-service/internal/auth"
	"github.com/gofiber/fiber/v3"
)

// Protected middleware checks if the user is authenticated and has the required roles.
func Protected(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"valid": false, "error": "Missing authentication: Authorization header required"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"valid": false, "error": "Missing or malformed JWT"})
		}
		token := parts[1]

		claims, err := auth.ParseJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"valid": false, "error": "Invalid or expired JWT"})
		}

		if claims.Type != auth.AccessToken {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"valid": false, "error": "Invalid token type"})
		}

		if len(allowedRoles) > 0 {
			hasAllowedRole := false
			for _, role := range claims.Roles {
				if slices.Contains(allowedRoles, role) {
					hasAllowedRole = true
					break
				}
			}
			if !hasAllowedRole {
				return c.Status(fiber.StatusForbidden).
					JSON(fiber.Map{"error": "Insufficient permissions"})
			}
		}

		c.Locals(auth.LocalsUserID, claims.UserID)
		c.Locals(auth.LocalsTelegramID, claims.TelegramID)
		c.Locals(auth.LocalsMaxID, claims.MaxID)
		c.Locals(auth.LocalsUserRoles, claims.Roles)
		c.Locals(auth.LocalsGrantType, claims.GTY)

		return c.Next()
	}
}

// CheckRole checks if the user has a specific role.
func CheckRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		roles, ok := c.Locals(auth.LocalsUserRoles).([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).
				JSON(fiber.Map{"error": "Could not determine user roles"})
		}

		if len(allowedRoles) > 0 {
			for _, role := range roles {
				if slices.Contains(allowedRoles, role) {
					return c.Next()
				}
			}
			return c.Status(fiber.StatusForbidden).
				JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		return c.Next()
	}
}
