package routes

func defineBotsAdmin(r *Registry, h *Handlers) {
	r.POST("/", h.BotHandler.CreateBot)
	r.GET("/", h.BotHandler.ListBots)
	r.PUT("/:id", h.BotHandler.UpdateBot)
	r.DELETE("/:id", h.BotHandler.DeleteBot)
}

func defineBotsService(r *Registry, h *Handlers) {
	r.GET("/token", h.BotHandler.GetBotTokenByNameAndPlatform)
	r.GET("/:name", h.BotHandler.GetBotByName)
}
