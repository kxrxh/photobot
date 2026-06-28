package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"csort.ru/coffeebot/internal/config"
	"csort.ru/coffeebot/internal/logger"
	"csort.ru/coffeebot/internal/observability"
	"csort.ru/coffeebot/internal/server"

	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Error loading configuration")
	}

	logger.InitLogger(cfg.Debug)

	log := logger.Logger

	otelShutdown, err := observability.InitOTEL(context.Background(), "photobot-backend")
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("otel init failed")
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	app, err := server.NewFromConfig(cfg, log)
	if err != nil {
		exitAfterOTEL(log, otelShutdown, err, "Error creating server")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info().Msg("Shutting down...")
		if err := app.Shutdown(); err != nil {
			exitAfterOTEL(log, otelShutdown, err, "Server shutdown failed")
		}
		log.Info().Msg("Server gracefully stopped")
		os.Exit(0)
	}()

	if err := app.Start(cfg.Server.Port); err != nil {
		exitAfterOTEL(log, otelShutdown, err, "Error starting server")
	}
}

func exitAfterOTEL(
	log zerolog.Logger,
	otelShutdown func(context.Context) error,
	err error,
	msg string,
) {
	log.Error().Err(err).Msg(msg)
	_ = otelShutdown(context.Background())
	os.Exit(1)
}
