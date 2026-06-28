package routes

import (
	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/middleware"
)

func Protected(authClient *auth.Client) Option {
	return func(rt *Route) {
		rt.Middleware = append(rt.Middleware, middleware.JWTAuth(authClient))
	}
}

func ProtectedWithRoles(authClient *auth.Client, roles ...string) Option {
	return func(rt *Route) {
		rt.Middleware = append(
			rt.Middleware,
			middleware.JWTAuth(authClient),
			middleware.WithRoles(roles...),
		)
	}
}
