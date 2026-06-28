package services

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

func StartServices(ctx context.Context, services []fiber.Service) error {
	for _, svc := range services {
		if err := svc.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}
