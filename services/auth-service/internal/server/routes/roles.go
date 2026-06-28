package routes

func defineRoles(r *Registry, h *Handlers) {
	r.POST("/", h.RoleHandler.CreateRole)
	r.GET("/", h.RoleHandler.ListRoles)
	r.GET("/:id", h.RoleHandler.GetRole)
	r.GET("/name/:name", h.RoleHandler.GetRoleByName)
	r.PUT("/:id", h.RoleHandler.UpdateRole)
	r.DELETE("/:id", h.RoleHandler.DeleteRole)
}
