package routes

import "csort.ru/auth-service/internal/middleware"

func Protected(roles ...string) Option {
	return func(rt *Route) {
		rt.Middleware = append(rt.Middleware, middleware.Protected(roles...))
	}
}
