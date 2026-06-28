package routes

import "csort.ru/classification-service/internal/auth"

func defineOwnership(r *Registry, h *Handlers, authClient *auth.Client) {
	r.POST(
		"/classifications/ownership-transfers",
		h.OwnershipHandler.TransferOwnership,
		ProtectedWithRoles(authClient, "service"),
	)
}
