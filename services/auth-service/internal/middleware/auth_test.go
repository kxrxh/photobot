package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"csort.ru/auth-service/internal/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtected(t *testing.T) {
	auth.InitTestKeys(t)

	app := fiber.New()
	app.Get("/protected", Protected(), func(c fiber.Ctx) error {
		userID := c.Locals(auth.LocalsUserID)
		return c.JSON(fiber.Map{"user_id": userID})
	})
	app.Get("/admin", Protected(auth.AdminRole), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	t.Run("missing Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("malformed Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid JWT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid JWT allows access", func(t *testing.T) {
		userID := int32(42)
		token, err := auth.GenerateJWT(&auth.GenerationParams{
			UserID: &userID,
			Roles:  []string{"user"},
			GTY:    auth.GrantTypePassword,
		}, auth.AccessToken, 5*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("insufficient role returns 403", func(t *testing.T) {
		userID := int32(1)
		token, err := auth.GenerateJWT(&auth.GenerationParams{
			UserID: &userID,
			Roles:  []string{"user"},
			GTY:    auth.GrantTypePassword,
		}, auth.AccessToken, 5*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("admin role allows access", func(t *testing.T) {
		userID := int32(1)
		token, err := auth.GenerateJWT(&auth.GenerationParams{
			UserID: &userID,
			Roles:  []string{auth.AdminRole},
			GTY:    auth.GrantTypePassword,
		}, auth.AccessToken, 5*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("service role does not bypass role check", func(t *testing.T) {
		svcID := "test-service"
		token, err := auth.GenerateJWT(&auth.GenerationParams{
			ServiceID: &svcID,
			Roles:     []string{auth.ServiceRole},
			GTY:       auth.GrantTypeService,
		}, auth.AccessToken, 5*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestCheckRole(t *testing.T) {
	t.Run("no roles in locals returns 403", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/r", CheckRole(auth.AdminRole), func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		resp, _ := app2.Test(req)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("wrong role returns 403", func(t *testing.T) {
		app2 := fiber.New()
		app2.Use(func(c fiber.Ctx) error {
			c.Locals(auth.LocalsUserRoles, []string{"user"})
			return c.Next()
		})
		app2.Get("/r", CheckRole(auth.AdminRole), func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		resp, _ := app2.Test(req)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("matching role allows access", func(t *testing.T) {
		app2 := fiber.New()
		app2.Use(func(c fiber.Ctx) error {
			c.Locals(auth.LocalsUserRoles, []string{"user", auth.AdminRole})
			return c.Next()
		})
		app2.Get("/r", CheckRole(auth.AdminRole), func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		resp, _ := app2.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("empty allowed roles allows any", func(t *testing.T) {
		app2 := fiber.New()
		app2.Use(func(c fiber.Ctx) error {
			c.Locals(auth.LocalsUserRoles, []string{"user"})
			return c.Next()
		})
		app2.Get("/r", CheckRole(), func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"ok": true})
		})
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		resp, _ := app2.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
