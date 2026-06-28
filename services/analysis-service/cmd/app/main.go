package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"csort.ru/analysis-service/internal/server"
)

func main() {
	appCtx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger.InitLogger()
	log := logger.GetLogger("main")

	otelShutdown, err := observability.InitOTEL(appCtx, "analysis-service")
	if err != nil {
		log.Fatal().Err(err).Msg("otel init failed")
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	log.Info().Msg("service started")

	app, err := server.New(appCtx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("create server failed")
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info().Msg("shutdown started")

		if err := app.Shutdown(); err != nil {
			log.Error().Err(err).Msg("server shutdown failed")
			return
		}
		log.Info().Msg("server stopped")
	}()

	if err := app.Start(cfg.App.Port); err != nil {
		log.Error().Err(err).Msg("start server failed")
		return
	}
}
