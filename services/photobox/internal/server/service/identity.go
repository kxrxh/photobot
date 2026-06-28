package service

import (
	"context"

	"csort.ru/coffeebot/internal/authz"
	"github.com/gofiber/fiber/v3"
)

// IdentityJWKS stops the identity Client JWKS refresh on Fiber shutdown.
type IdentityJWKS struct {
	client *authz.Client
}

var _ fiber.Service = (*IdentityJWKS)(nil)

func NewIdentityJWKS(client *authz.Client) *IdentityJWKS {
	return &IdentityJWKS{client: client}
}

func (s *IdentityJWKS) String() string {
	return "Identity (JWKS)"
}

func (s *IdentityJWKS) Start(_ context.Context) error {
	return nil
}

func (s *IdentityJWKS) State(_ context.Context) (string, error) {
	return "running", nil
}

func (s *IdentityJWKS) Terminate(_ context.Context) error {
	s.client.Stop()
	return nil
}
