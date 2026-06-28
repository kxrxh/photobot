package server

import (
	"context"
	"fmt"
	"time"

	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/api/classification"
	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/api/reports"
	"csort.ru/analysis-service/internal/api/worker"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/messaging"
	"csort.ru/analysis-service/internal/storage"
	"csort.ru/analysis-service/internal/transport/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type infraDeps struct {
	rabbitMQClient        *messaging.Client
	authClient            *auth.Client
	analysisClient        *worker.Client
	reportsClient         *reports.Client
	classificationClient  *classification.Client
	webSocketHub          *ws.Hub
	tempStorageClient     *storage.Client
	analysisStorageClient *storage.Client
	redisClient           *redis.Client
	requestsPool          *pgxpool.Pool
}

func buildInfra(
	ctx context.Context,
	cfg *config.Config,
	requestsPool *pgxpool.Pool,
	redisClient *redis.Client,
	tempStorageClient *storage.Client,
	analysisStorageClient *storage.Client,
) (*infraDeps, error) {
	rabbitMQClient, err := messaging.NewClient(&cfg.RabbitMQ, false, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("initialize rabbitmq client: %w", err)
	}

	authClient, err := auth.NewClient(auth.Config{
		BaseURL:             cfg.Security.ServiceUrl,
		ServiceID:           cfg.Security.ServiceId,
		ServiceSecret:       cfg.Security.ServiceSecret,
		JWKSRefreshInterval: cfg.Security.JWKSRefreshInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize auth client: %w", err)
	}

	tokenManager, err := common.NewTokenManager(ctx, common.TokenManagerConfig{
		Interval:                4 * time.Minute,
		ObtainTokens:            authClient.ObtainTokens,
		RefreshWithRefreshToken: authClient.RefreshWithRefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize reports token manager: %w", err)
	}

	reportsClient, err := reports.NewClient(reports.Config{
		BaseURL:      cfg.ReportsAPI,
		Timeout:      30 * time.Minute,
		MaxRetries:   3,
		GetToken:     tokenManager.GetToken,
		RefreshToken: tokenManager.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize reports client: %w", err)
	}

	const analysisWorkerHTTPTimeout = 30 * time.Minute

	analysisClient := worker.NewClient(worker.Config{
		BaseURL:    cfg.AnalysisAPI,
		Timeout:    analysisWorkerHTTPTimeout,
		MaxRetries: 2,
		RetryDelay: 200 * time.Millisecond,
	})

	classificationClient, err := classification.NewClient(classification.Config{
		BaseURL:      cfg.ClassificationAPI,
		Timeout:      10 * time.Second,
		MaxRetries:   2,
		RetryDelay:   200 * time.Millisecond,
		GetToken:     authClient.GetToken,
		RefreshToken: authClient.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize classification client: %w", err)
	}

	return &infraDeps{
		rabbitMQClient:        rabbitMQClient,
		authClient:            authClient,
		analysisClient:        analysisClient,
		reportsClient:         reportsClient,
		classificationClient:  classificationClient,
		webSocketHub:          ws.NewHub(ctx, redisClient),
		tempStorageClient:     tempStorageClient,
		analysisStorageClient: analysisStorageClient,
		redisClient:           redisClient,
		requestsPool:          requestsPool,
	}, nil
}

func assembleContainer(infra *infraDeps, core *coreServices) *Container {
	return &Container{
		AnalysisService:   core.analysisService,
		ObjectsService:    core.objectsService,
		RequestsService:   core.requestsService,
		OutboxRelay:       core.outboxRelay,
		ReportService:     core.reportService,
		ImageService:      core.imageService,
		AuthClient:        infra.authClient,
		ClassificationAPI: infra.classificationClient,
		RabbitMQClient:    infra.rabbitMQClient,
		WebSocketHub:      infra.webSocketHub,
		TempStorageClient: infra.tempStorageClient,
		RedisClient:       infra.redisClient,
		RequestsPool:      infra.requestsPool,
	}
}
