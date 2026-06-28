package server

import (
	"context"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/api/classification"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/image"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/messaging"
	"csort.ru/analysis-service/internal/messenger/max"
	"csort.ru/analysis-service/internal/messenger/telegram"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/report"
	"csort.ru/analysis-service/internal/repository/kalibr"
	repo_requests "csort.ru/analysis-service/internal/repository/requests"
	"csort.ru/analysis-service/internal/requests"
	"csort.ru/analysis-service/internal/storage"
	"csort.ru/analysis-service/internal/transport/http"
	"csort.ru/analysis-service/internal/transport/ws"
	validatepkg "csort.ru/analysis-service/internal/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	AnalysisService   *analysis.Service
	ObjectsService    *objects.Service
	RequestsService   *requests.Service
	OutboxRelay       *requests.OutboxRelay
	ReportService     *report.Service
	ImageService      *image.Service
	AuthClient        *auth.Client
	ClassificationAPI *classification.Client
	RabbitMQClient    *messaging.Client
	WebSocketHub      *ws.Hub
	TempStorageClient *storage.Client
	RedisClient       *redis.Client
	RequestsPool      *pgxpool.Pool
}

type coreServices struct {
	analysisService *analysis.Service
	objectsService  *objects.Service
	requestsService *requests.Service
	outboxRelay     *requests.OutboxRelay
	reportService   *report.Service
	imageService    *image.Service
}

func buildCoreServices(
	dbKalibr *kalibr.Queries,
	dbRequests *repo_requests.Queries,
	infra *infraDeps,
	cfg *config.Config,
) *coreServices {
	reportService := report.NewReportService(infra.reportsClient)
	requestsService := requests.NewRequestsService(requests.RequestsServiceParams{
		Repo:           dbRequests,
		Pool:           infra.requestsPool,
		AnalysisClient: infra.analysisClient,
		Classification: infra.classificationClient,
		WebSocketHub:   infra.webSocketHub,
		MinIOClient:    infra.tempStorageClient,
	})

	return &coreServices{
		analysisService: analysis.NewAnalysisService(dbKalibr, infra.analysisClient),
		objectsService: objects.NewObjectsService(
			dbKalibr,
			dbRequests,
			infra.tempStorageClient,
		),
		requestsService: requestsService,
		outboxRelay: requests.NewOutboxRelay(
			dbRequests,
			infra.rabbitMQClient,
			&cfg.RabbitMQ,
			&cfg.OutboxRelay,
		),
		reportService: reportService,
		imageService:  image.NewImageService(dbKalibr, infra.analysisStorageClient),
	}
}

func initializeServiceContainer(
	ctx context.Context,
	dbKalibr *kalibr.Queries,
	dbRequests *repo_requests.Queries,
	requestsPool *pgxpool.Pool,
	redisClient *redis.Client,
	tempStorageClient *storage.Client,
	analysisStorageClient *storage.Client,
	cfg *config.Config,
) (*Container, error) {
	infra, err := buildInfra(
		ctx,
		cfg,
		requestsPool,
		redisClient,
		tempStorageClient,
		analysisStorageClient,
	)
	if err != nil {
		return nil, err
	}
	core := buildCoreServices(dbKalibr, dbRequests, infra, cfg)
	return assembleContainer(infra, core), nil
}

type Handlers struct {
	AnalysisHandler  *http.AnalysisHandler
	ObjectsHandler   *http.ObjectsHandler
	RequestsHandler  *http.RequestsHandler
	ReportHandler    *http.ReportHandler
	ImageHandler     *http.ImageHandler
	WebSocketHandler *http.WebSocketHandler
	OwnershipHandler *http.OwnershipHandler
	AuthClient       *auth.Client
}

func initializeHandlers(services *Container, share config.ShareLinkConfig) Handlers {
	v := validatepkg.Default
	packAuth := http.NewReportPackAuthorizer(share, services.AuthClient, services.AnalysisService)
	analysisHandler := http.NewAnalysisHandler(
		logger.GetLogger("transport.http.analysis"),
		services.AnalysisService,
		services.ImageService,
		services.RequestsService,
		services.TempStorageClient,
		v,
	)
	objectsHandler := http.NewObjectsHandler(
		logger.GetLogger("transport.http.objects"),
		services.ObjectsService,
	)
	requestsHandler := http.NewRequestsHandler(
		services.RequestsService,
		services.RedisClient,
		v,
	)
	telegramClient := telegram.NewClient(nil)
	maxClient := max.NewClient(nil)
	reportHandler := http.NewReportHandler(
		logger.GetLogger("transport.http.report"),
		services.ReportService,
		services.AnalysisService,
		services.AuthClient,
		telegramClient,
		maxClient,
	)
	imageHandler := http.NewImageHandler(
		logger.GetLogger("transport.http.image"),
		services.ImageService,
		services.AnalysisService,
		packAuth,
	)
	webSocketHandler := http.NewWebSocketHandler(
		logger.GetLogger("transport.http.websocket"),
		services.WebSocketHub,
		services.AuthClient,
	)
	ownershipHandler := http.NewOwnershipHandler(
		logger.GetLogger("transport.http.ownership"),
		services.RequestsService,
	)

	return Handlers{
		AnalysisHandler:  analysisHandler,
		ObjectsHandler:   objectsHandler,
		RequestsHandler:  requestsHandler,
		ReportHandler:    reportHandler,
		ImageHandler:     imageHandler,
		WebSocketHandler: webSocketHandler,
		OwnershipHandler: ownershipHandler,
		AuthClient:       services.AuthClient,
	}
}

func defineRoutes(
	app *fiber.App,
	h *Handlers,
	healthHandler fiber.Handler,
	redisClient *redis.Client,
	maxQueuedRequests int,
) {
	api := app.Group("/api/v1")

	api.Get("/health", healthHandler)

	analyses := api.Group("/analyses")
	analyses.Get("/:id/images/source/:index", h.ImageHandler.DownloadSourceImage)
	analyses.Get("/:id/images/analysis/:index", h.ImageHandler.DownloadOutputImage)
	analyses.Get("/:id/images/objects/archive", h.ImageHandler.DownloadObjectImagesArchive)
	analyses.Get("/:id/images/objects/:objectId", h.ImageHandler.DownloadObjectImage)
	analyses.Post(
		"/ownership-transfers",
		middleware.JWTAuth(h.AuthClient),
		middleware.WithRoles("service"),
		h.OwnershipHandler.TransferOwnership,
	)

	protected := analyses.Group("", middleware.JWTAuth(h.AuthClient))
	protected.Post(
		"/notify",
		middleware.WithRoles("service"),
		h.RequestsHandler.NotifyProcessingCompletion,
	)
	protected.Get("/", h.AnalysisHandler.GetAnalyses)
	protected.Get("/:id", h.AnalysisHandler.GetAnalysisByID)
	protected.Post(
		"/",
		middleware.RateLimit(middleware.RateLimitConfig{
			MaxQueuedRequests: maxQueuedRequests,
			RedisClient:       redisClient,
		}),
		h.AnalysisHandler.CreateAnalysis,
	)
	protected.Get("/:id/objects", h.AnalysisHandler.GetAnalysisObjects)
	protected.Post("/merge", h.AnalysisHandler.MergeAnalyses)
	protected.Post("/:id/report/send-to-chat", h.ReportHandler.SendToChat)

	api.Group("/objects", middleware.JWTAuth(h.AuthClient)).
		Post("/search", h.ObjectsHandler.SearchObjects)

	requests := api.Group("/requests", middleware.JWTAuth(h.AuthClient))
	requests.Get("/", h.RequestsHandler.GetRequests)
	requests.Get("/:requestId", h.RequestsHandler.GetRequestByID)
	requests.Get("/objects/:requestId", h.ObjectsHandler.GetObjectsByRequestId)
	requests.Post("/confirm", h.RequestsHandler.ConfirmRequest)

	api.Group("/ws").
		Get("/",
			h.WebSocketHandler.UpgradeMiddleware,
			h.WebSocketHandler.HandleWebSocket)
}
