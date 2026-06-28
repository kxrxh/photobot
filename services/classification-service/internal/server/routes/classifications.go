package routes

import "csort.ru/classification-service/internal/auth"

func defineClassifications(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST("/", h.ClassificationHandler.CreateCompleteClassification)
	r.GET("/", h.ClassificationHandler.ListClassifications)
	r.GET("/:id", h.ClassificationHandler.GetCompleteClassification)
	r.PUT("/:id", h.ClassificationHandler.UpdateClassification)
	r.DELETE("/:id", h.ClassificationHandler.DeleteClassification)
	r.PUT("/:id/public", h.ClassificationHandler.MakeClassificationPublic,
		ProtectedWithRoles(authClient, "admin", "classification_editor"))
	r.PUT("/:id/private", h.ClassificationHandler.MakeClassificationPrivate,
		ProtectedWithRoles(authClient, "admin", "classification_editor"))
}
