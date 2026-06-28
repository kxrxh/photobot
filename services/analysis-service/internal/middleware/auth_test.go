package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apifiber "csort.ru/analysis-service/internal/apierrors/fiber"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestWithRoles(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apifiber.ErrorHandler})
	app.Use(func(c fiber.Ctx) error {
		c.Locals(UserRolesKey, []string{"user", "admin"})
		return c.Next()
	})

	t.Run("allows access when user has required role", func(t *testing.T) {
		app.Get("/admin", WithRoles("admin"), func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("allows access when user has any of required roles", func(t *testing.T) {
		app.Get("/user", WithRoles("user", "viewer"), func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("denies access when user lacks required role", func(t *testing.T) {
		app.Get("/worker", WithRoles("worker"), func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/worker", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("allows access when no roles required", func(t *testing.T) {
		app.Get("/open", WithRoles(), func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/open", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("denies access when user roles not set", func(t *testing.T) {
		appNoRoles := fiber.New(fiber.Config{ErrorHandler: apifiber.ErrorHandler})
		appNoRoles.Get("/admin", WithRoles("admin"), func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		resp, err := appNoRoles.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}
