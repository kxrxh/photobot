package routes

import "csort.ru/auth-service/internal/auth"

func defineUsers(r *Registry, h *Handlers) {
	r.GET("/me", h.UserHandler.GetMe, Protected())
	r.GET(
		"/by-messenger-id/:id",
		h.UserHandler.GetUserByMessengerId,
		Protected(auth.AdminRole, auth.ServiceRole),
	)
	r.GET("/:id", h.UserHandler.GetUser, Protected(auth.AdminRole))
	r.PUT("/me", h.UserHandler.UpdateMe, Protected())
	r.PUT("/:id", h.UserHandler.UpdateUser, Protected(auth.AdminRole))
	r.GET("/", h.UserHandler.ListUsers, Protected(auth.AdminRole))
	r.GET("/me/roles", h.RoleHandler.GetMyRoles, Protected())
	r.GET("/:user_id/roles", h.RoleHandler.GetUserRoles, Protected(auth.AdminRole))
	r.POST("/roles", h.RoleHandler.AssignRole, Protected(auth.AdminRole))
	r.DELETE("/roles", h.RoleHandler.RevokeRole, Protected(auth.AdminRole))
}
