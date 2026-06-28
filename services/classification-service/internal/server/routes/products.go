package routes

import "csort.ru/classification-service/internal/auth"

func defineProducts(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST("/", h.ProductHandler.CreateProduct, ProtectedWithRoles(authClient, "admin"))
	r.GET("/", h.ProductHandler.ListProducts)
	r.GET("/:id", h.ProductHandler.GetProduct)
	r.PUT("/:id", h.ProductHandler.UpdateProduct, ProtectedWithRoles(authClient, "admin"))
	r.DELETE("/:id", h.ProductHandler.DeleteProduct, ProtectedWithRoles(authClient, "admin"))
}
