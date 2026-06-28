package routes

import "csort.ru/classification-service/internal/auth"

func defineCorrelation(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST("/correlation", h.CorrelationHandler.CalculateCorrelation, Protected(authClient))
}
