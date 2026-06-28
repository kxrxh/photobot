package routes

import (
	"csort.ru/classification-service/internal/auth"
	"github.com/gofiber/fiber/v3"
)

func Define(h *Handlers, authClient *auth.Client, healthHandler fiber.Handler) []Route {
	r := NewRegistry()
	r.GET("/health", healthHandler)

	api := r.Group("/api/v1")
	defineOwnership(api, h, authClient)
	defineProducts(api.Group("/products", Protected(authClient)), h, authClient)
	defineClassifications(
		api.Group("/classifications", Protected(authClient)),
		h,
		authClient,
	)
	defineUserActiveClassifications(
		api.Group("/user-active-classifications", Protected(authClient)),
		h,
		authClient,
	)
	defineMarkup(api.Group("/markup", Protected(authClient)), h)
	defineParams(api.Group("/params", Protected(authClient)), h, authClient)
	defineCorrelation(api, h, authClient)

	return r.List()
}
