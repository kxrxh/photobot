package http

import (
	"context"
	"time"

	"csort.ru/analysis-service/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

const healthProbeTimeout = 2 * time.Second

type RabbitMQHealthChecker interface {
	IsConnected() bool
}

type HealthHandler struct {
	kalibrDB    *database.DB
	requestsDB  *database.DB
	redisClient *redis.Client
	rabbitMQ    RabbitMQHealthChecker
}

func NewHealthHandler(
	kalibrDB *database.DB,
	requestsDB *database.DB,
	redisClient *redis.Client,
	rabbitMQ RabbitMQHealthChecker,
) *HealthHandler {
	return &HealthHandler{
		kalibrDB:    kalibrDB,
		requestsDB:  requestsDB,
		redisClient: redisClient,
		rabbitMQ:    rabbitMQ,
	}
}

func (h *HealthHandler) Handle(
	c fiber.Ctx,
) error { //nolint:contextcheck // Fiber handler signature does not accept context.Context.
	ctx, cancel := context.WithTimeout(c.Context(), healthProbeTimeout)
	defer cancel()

	checks := make(map[string]string)
	allOk := true

	if err := h.kalibrDB.Ping(ctx); err != nil {
		checks["database_kalibr"] = err.Error()
		allOk = false
	} else {
		checks["database_kalibr"] = "ok"
	}

	if err := h.requestsDB.Ping(ctx); err != nil {
		checks["database_requests"] = err.Error()
		allOk = false
	} else {
		checks["database_requests"] = "ok"
	}

	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		checks["redis"] = err.Error()
		allOk = false
	} else {
		checks["redis"] = "ok"
	}

	if h.rabbitMQ != nil && !h.rabbitMQ.IsConnected() {
		checks["rabbitmq"] = "not connected"
		allOk = false
	} else if h.rabbitMQ != nil {
		checks["rabbitmq"] = "ok"
	}

	resp := fiber.Map{
		"service": "analysis-service",
		"time":    time.Now().Format(time.RFC3339),
		"checks":  checks,
	}
	if allOk {
		resp["status"] = "healthy"
		return c.JSON(resp)
	}
	resp["status"] = "unhealthy"
	return c.Status(fiber.StatusServiceUnavailable).JSON(resp)
}
