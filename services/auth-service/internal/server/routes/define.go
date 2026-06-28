package routes

import (
	"csort.ru/auth-service/internal/auth"
	"github.com/gofiber/fiber/v3"
)

func Define(h *Handlers, healthHandler fiber.Handler) []Route {
	r := NewRegistry()
	api := r.Group("/api/v1")
	api.GET("/health", healthHandler)

	defineUsers(api.Group("/users"), h)
	defineAuth(api.Group("/auth"), h)
	defineRoles(api.Group("/roles", Protected(auth.AdminRole)), h)
	defineServices(api.Group("/services", Protected(auth.AdminRole)), h)
	defineBotsAdmin(api.Group("/bots", Protected(auth.AdminRole)), h)
	defineBotsService(api.Group("/bots", Protected(auth.ServiceRole)), h)

	return r.List()
}
