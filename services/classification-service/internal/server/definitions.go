package server

import (
	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/classification"
	"csort.ru/classification-service/internal/config"
	"csort.ru/classification-service/internal/correlation"
	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/markup"
	"csort.ru/classification-service/internal/ownership"
	"csort.ru/classification-service/internal/params"
	"csort.ru/classification-service/internal/product"
	"csort.ru/classification-service/internal/transport/http"
	"csort.ru/classification-service/internal/user"
	validatepkg "csort.ru/classification-service/pkg/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	ClassificationHandler           *http.ClassificationHandler
	ProductHandler                  *http.ProductHandler
	UserActiveClassificationHandler *http.UserActiveClassificationHandler
	MarkupHandler                   *http.MarkupHandler
	CorrelationHandler              *http.CorrelationHandler
	ClassificationParamsHandler     *http.ClassificationParamsHandler
	OwnershipHandler                *http.OwnershipHandler
}

type Services struct {
	ClassificationService           *classification.ClassificationService
	ProductService                  *product.ProductService
	UserActiveClassificationService *user.UserActiveClassificationService
	MarkupService                   *markup.MarkupService
	CorrelationService              *correlation.CorrelationService
	AuthTokenManager                *auth.TokenManager
	CorrelationTokenManager         *auth.TokenManager
	AuthServiceClient               *auth.Client
	ClassificationParamsService     *params.ClassificationParamsService
	MergeService                    *ownership.Service
}

func initializeServices(dbPool *pgxpool.Pool, cfg *config.Config) *Services {
	dbQueries := database.New(dbPool)
	classificationService := classification.NewClassificationService(dbQueries, dbPool)
	productService := product.NewProductService(dbQueries)
	userActiveClassificationService := user.NewUserActiveClassificationService(
		dbQueries,
		classificationService,
	)
	authClient := auth.NewClient(
		cfg.AuthServiceURL,
		cfg.Security.ServiceID,
		cfg.Security.ServiceSecret,
	)
	markupService := markup.NewMarkupService(dbQueries, dbPool)

	authTokenManager := auth.NewTokenManager(authClient, "auth-service")
	correlationTokenManager := auth.NewTokenManager(authClient, "correlation-service")
	correlationService := correlation.NewCorrelationService(
		cfg.CorrelationServiceURL,
		correlationTokenManager,
	)
	classificationParamsService := params.NewClassificationParamsService(dbQueries)
	mergeService := ownership.NewService(dbQueries, dbPool)

	return &Services{
		ClassificationService:           classificationService,
		ProductService:                  productService,
		UserActiveClassificationService: userActiveClassificationService,
		MarkupService:                   markupService,
		CorrelationService:              correlationService,
		AuthTokenManager:                authTokenManager,
		CorrelationTokenManager:         correlationTokenManager,
		AuthServiceClient:               authClient,
		ClassificationParamsService:     classificationParamsService,
		MergeService:                    mergeService,
	}
}

func initializeHandlers(services *Services) *Handlers {
	v := validatepkg.Default
	return &Handlers{
		ClassificationHandler: http.NewClassificationHandler(services.ClassificationService),
		ProductHandler:        http.NewProductHandler(services.ProductService),
		UserActiveClassificationHandler: http.NewUserActiveClassificationHandler(
			services.UserActiveClassificationService,
			services.AuthServiceClient,
			services.AuthTokenManager,
		),
		MarkupHandler:      http.NewMarkupHandler(services.MarkupService, v),
		CorrelationHandler: http.NewCorrelationHandler(services.CorrelationService),
		ClassificationParamsHandler: http.NewClassificationParamsHandler(
			services.ClassificationParamsService,
			v,
		),
		OwnershipHandler: http.NewOwnershipHandler(services.MergeService),
	}
}
