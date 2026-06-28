package middleware

import (
	"context"
	"errors"
	"slices"

	"csort.ru/coffeebot/internal/authz"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel/trace"
)

func ContextWithRequestSpan(parent, request context.Context) context.Context {
	if request == nil {
		return parent
	}
	span := trace.SpanFromContext(request)
	if !span.SpanContext().IsValid() {
		return parent
	}
	return trace.ContextWithSpan(parent, span)
}

// GetCurrentIdentity returns JWT identity from Fiber locals (set by JWTAuth).
func GetCurrentIdentity(c fiber.Ctx) (*authz.Identity, error) {
	idnt, ok := c.Locals(UserDataKey).(*authz.Identity)
	if !ok || idnt == nil {
		return nil, errors.New("identity not found in context")
	}
	return idnt, nil
}

// GetUserRoles returns role names from JWT claims (set by JWTAuth).
func GetUserRoles(c fiber.Ctx) ([]string, error) {
	roles, ok := c.Locals(UserRolesKey).([]string)
	if !ok {
		return nil, errors.New("user roles not found in context")
	}
	return roles, nil
}

func HasAdminOrModeratorRole(c fiber.Ctx) bool {
	userRoles, ok := c.Locals(UserRolesKey).([]string)
	if !ok {
		return false
	}
	return slices.Contains(userRoles, "admin") || slices.Contains(userRoles, "moderator")
}
