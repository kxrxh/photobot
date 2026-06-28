package server

import (
	"csort.ru/classification-service/internal/server/routes"
	"github.com/gofiber/fiber/v3"
)

func registerRoutes(app *fiber.App, r []routes.Route) {
	for _, rt := range r {
		handlers := []any{}
		for _, m := range rt.Middleware {
			handlers = append(handlers, m)
		}
		handlers = append(handlers, rt.Handler)
		app.Add([]string{rt.Method}, rt.Path, handlers[0], handlers[1:]...)
	}
}
