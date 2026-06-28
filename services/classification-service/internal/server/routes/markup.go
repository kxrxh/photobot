package routes

func defineMarkup(r *Registry, h *Handlers) {
	r.POST("/", h.MarkupHandler.CreateMarkup)
	r.GET("/", h.MarkupHandler.ListMarkups)
	r.GET("/:id", h.MarkupHandler.GetMarkup)
	r.PUT("/:id", h.MarkupHandler.UpdateMarkup)
	r.DELETE("/:id", h.MarkupHandler.DeleteMarkup)
}
