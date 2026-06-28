package routes

import (
	"csort.ru/auth-service/internal/transport/http"
)

type Handlers struct {
	UserHandler    *http.UserHandler
	RoleHandler    *http.RoleHandler
	BotHandler     *http.BotHandler
	AuthHandler    *http.AuthHandler
	ServiceHandler *http.ServiceHandler
}
