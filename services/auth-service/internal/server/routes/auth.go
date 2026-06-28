package routes

func defineAuth(r *Registry, h *Handlers) {
	r.POST("/register", h.AuthHandler.RegisterUser)
	r.POST("/register-web", h.AuthHandler.RegisterWeb)
	r.POST("/refresh", h.AuthHandler.Refresh)
	r.POST("/login", h.AuthHandler.Login)
	r.POST("/forgot-password", h.AuthHandler.ForgotPassword)
	r.POST("/reset-password", h.AuthHandler.ResetPassword)
	r.POST("/reset-password-recovery", h.AuthHandler.ResetPasswordRecovery)
	r.POST(
		"/change-password",
		h.AuthHandler.ChangePassword,
		Protected(),
	)
	r.POST(
		"/setup-web-access",
		h.AuthHandler.SetupWebAccess,
		Protected(),
	)
	r.POST(
		"/link-code",
		h.AuthHandler.RequestLinkCode,
		Protected(),
	)
	r.POST(
		"/link-with-code",
		h.AuthHandler.LinkWithCode,
		Protected(),
	)
	r.POST(
		"/link-with-code-from-web",
		h.AuthHandler.LinkWithCodeFromWeb,
		Protected(),
	)
	r.GET("/.well-known/jwks.json", h.AuthHandler.GetJWKS)
}
