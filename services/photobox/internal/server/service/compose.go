package service

import (
	"csort.ru/coffeebot/internal/authz"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compose registers Fiber Services so Terminate closes the DB pool and identity JWKS loop.
func Compose(dbPool *pgxpool.Pool, identityClient authz.IdentityClient) []fiber.Service {
	out := []fiber.Service{NewPGXPool(dbPool)}
	if c, ok := identityClient.(*authz.Client); ok {
		out = append(out, NewIdentityJWKS(c))
	}
	return out
}
