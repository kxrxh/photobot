package routes

import "csort.ru/classification-service/internal/auth"

func defineParams(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST("/", h.ClassificationParamsHandler.Create, ProtectedWithRoles(authClient, "admin"))
	r.GET("/", h.ClassificationParamsHandler.List)
	r.DELETE("/:id", h.ClassificationParamsHandler.Delete, ProtectedWithRoles(authClient, "admin"))
}
