//go:build integration

package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/api/reports"
	"csort.ru/analysis-service/internal/api/worker"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/messaging"
	"csort.ru/analysis-service/internal/repository/kalibr"
	repo_requests "csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/storage"
	"csort.ru/analysis-service/internal/transport/ws"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type IntegrationInfraExtras struct {
	ReportsClient *reports.Client
}

func IntegrationBuildInfra(
	ctx context.Context,
	cfg *config.Config,
	requestsPool *pgxpool.Pool,
	redisClient *redis.Client,
	tempStorageClient *storage.Client,
	analysisStorageClient *storage.Client,
	extras IntegrationInfraExtras,
) (*infraDeps, error) {
	if extras.ReportsClient == nil {
		return nil, errors.New("integration: ReportsClient is required")
	}
	rabbitMQClient, err := messaging.NewClient(&cfg.RabbitMQ, false, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: %w", err)
	}
	authClient, err := auth.NewClient(auth.Config{
		BaseURL:             cfg.Security.ServiceUrl,
		ServiceID:           cfg.Security.ServiceId,
		ServiceSecret:       cfg.Security.ServiceSecret,
		Timeout:             5 * time.Second,
		JWKSRefreshInterval: cfg.Security.JWKSRefreshInterval,
	})
	if err != nil {
		return nil, err
	}
	analysisClient := worker.NewClient(worker.Config{
		BaseURL:    cfg.AnalysisAPI,
		Timeout:    30 * time.Second,
		MaxRetries: 2,
	})
	return &infraDeps{
		rabbitMQClient:        rabbitMQClient,
		authClient:            authClient,
		analysisClient:        analysisClient,
		reportsClient:         extras.ReportsClient,
		webSocketHub:          ws.NewHub(ctx, redisClient),
		tempStorageClient:     tempStorageClient,
		analysisStorageClient: analysisStorageClient,
		redisClient:           redisClient,
		requestsPool:          requestsPool,
	}, nil
}

func IntegrationWireContainer(
	ctx context.Context,
	dbKalibr *kalibr.Queries,
	dbRequests *repo_requests.Queries,
	requestsPool *pgxpool.Pool,
	redisClient *redis.Client,
	tempStorageClient *storage.Client,
	analysisStorageClient *storage.Client,
	cfg *config.Config,
	extras IntegrationInfraExtras,
) (*Container, error) {
	infra, err := IntegrationBuildInfra(
		ctx,
		cfg,
		requestsPool,
		redisClient,
		tempStorageClient,
		analysisStorageClient,
		extras,
	)
	if err != nil {
		return nil, err
	}
	core := buildCoreServices(dbKalibr, dbRequests, infra, cfg)
	return assembleContainer(infra, core), nil
}

func IntegrationInitHandlers(c *Container, share config.ShareLinkConfig) Handlers {
	return initializeHandlers(c, share)
}

func IntegrationMountRoutes(
	app *fiber.App,
	h *Handlers,
	healthHandler fiber.Handler,
	redisClient *redis.Client,
	maxQueuedRequests int,
) {
	defineRoutes(app, h, healthHandler, redisClient, maxQueuedRequests)
}

func IntegrationSetupMiddleware(
	app *fiber.App,
	redisClient *redis.Client,
	analysisAPIURL string,
) {
	setupMiddleware(app, redisClient, analysisAPIURL)
}
