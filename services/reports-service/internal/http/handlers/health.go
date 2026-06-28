package handlers

import (
	"context"
	"time"

	"csort.ru/reports-service/internal/storage"

	"github.com/gofiber/fiber/v3"
)

const healthCheckTimeout = 3 * time.Second

// Health returns a handler that reports process liveness and MinIO connectivity.
func Health(minio *storage.MinIOService) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), healthCheckTimeout)
		defer cancel()

		if err := minio.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).
				JSON(map[string]any{
					"status": "unhealthy",
					"minio": map[string]any{
						"status": "error",
						"error":  err.Error(),
					},
				})
		}

		return c.JSON(map[string]any{
			"status": "ok",
			"minio": map[string]any{
				"status": "ok",
			},
		})
	}
}
