package routes

import "csort.ru/classification-service/internal/auth"

func defineUserActiveClassifications(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST("/", h.UserActiveClassificationHandler.SetUserActiveClassification)
	r.GET(
		"/:messenger_user_id",
		h.UserActiveClassificationHandler.GetUserActiveClassificationWithDetails,
		ProtectedWithRoles(authClient, "admin", "service"),
	)
	r.GET("/", h.UserActiveClassificationHandler.GetUserActiveClassification)
	r.DELETE("/", h.UserActiveClassificationHandler.DeleteUserActiveClassification)
}
