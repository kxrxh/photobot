package main

import (
	"context"
	"fmt"
	"time"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/http/handlers"
	"csort.ru/reports-service/internal/http/router"
	"csort.ru/reports-service/internal/logger"
	"csort.ru/reports-service/internal/observability"
	"csort.ru/reports-service/internal/redislimit"
	"csort.ru/reports-service/internal/render"
	"csort.ru/reports-service/internal/reports"
	"csort.ru/reports-service/internal/storage"

	"github.com/gofiber/fiber/v3"
)

func main() {
	logger.InitLogger()
	appCtx := context.Background()

	otelShutdown, err := observability.InitOTEL(appCtx, "reports-service")
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("otel init failed")
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	cfg, err := config.Load()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("config")
	}
	logger.Logger.Info().
		Int("listen_port", cfg.Server.Port).
		Bool("trust_proxy", cfg.Server.TrustProxy).
		Bool("redis_rate_limit", cfg.Redis.URL != "").
		Bool("minio_presign_rewrite", cfg.MinIO.PublicBaseURL != "").
		Msg("reports service config loaded")

	mainTpl := render.MainHTMLPath(cfg.Templates.Dir)
	if err := render.LoadReportTemplate(mainTpl); err != nil {
		logger.Logger.Fatal().Err(err).Str("path", mainTpl).Msg("load report template")
	}

	minioSvc, err := storage.NewMinIOService(cfg.MinIO)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("minio")
	}
	if cfg.MinIO.PublicBaseURL != "" {
		logger.Logger.Info().
			Str("minio_public_base_url", cfg.MinIO.PublicBaseURL).
			Msg("minio presign rewrite enabled")
	}

	analysisClient := analysis.New(cfg.AnalysisService.Host)
	pdf := reports.NewPDFConverter(5)
	warmupCtx, warmupCancel := context.WithTimeout(appCtx, 3*time.Minute)
	if err := pdf.WarmUp(warmupCtx, cfg.Templates.Dir, reports.PDFWarmupHTML); err != nil {
		logger.Logger.Warn().
			Err(err).
			Msg("pdf chrome warmup failed; first report may be slow or produce a bad file until Chrome is ready")
	} else {
		logger.Logger.Info().Msg("pdf chrome warmup completed")
	}
	warmupCancel()
	reportSvc := reports.NewService(cfg, analysisClient, minioSvc, pdf)
	reportsHandler := handlers.NewReportsHandler(reportSvc, minioSvc, cfg)

	var rateLimitStorage fiber.Storage
	if cfg.Redis.URL != "" {
		redisCtx, redisCancel := context.WithTimeout(appCtx, 5*time.Second)
		st, err := redislimit.OpenLimiterStorage(redisCtx, cfg.Redis.URL)
		redisCancel()
		if err != nil {
			logger.Logger.Fatal().Err(err).Msg("redis (rate limit)")
		}
		rateLimitStorage = st
	}

	app, err := router.New(cfg, reportsHandler, minioSvc, rateLimitStorage)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("router")
	}
	app.Hooks().OnPreShutdown(func() error {
		pdf.Close()
		return nil
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Logger.Info().Str("addr", addr).Msg("reports api listening")
	if err := app.Listen(addr); err != nil {
		logger.Logger.Fatal().Err(err).Msg("listen")
	}
}
