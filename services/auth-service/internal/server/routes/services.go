package routes

func defineServices(r *Registry, h *Handlers) {
	r.POST("/", h.ServiceHandler.CreateService)
	r.GET("/", h.ServiceHandler.ListServices)
	r.DELETE("/:service_id", h.ServiceHandler.DeleteService)
}
