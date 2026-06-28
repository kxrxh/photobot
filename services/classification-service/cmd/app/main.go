package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csort.ru/classification-service/internal/config"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/observability"
	"csort.ru/classification-service/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	log := logger.GetLogger("main.connectDB")

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Successfully connected to database")
	return dbPool, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load configuration: %w", err))
	}

	logger.InitLogger()

	os.Exit(run(cfg))
}

func run(cfg *config.Config) int {
	log := logger.GetLogger("main")

	otelShutdown, err := observability.InitOTEL(
		context.Background(),
		cfg.Observability,
		"classification-service",
	)
	if err != nil {
		log.Error().Err(err).Msg("otel init failed")
		return 1
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	dbPool, err := connectDB(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return 1
	}

	srv, err := server.New(cfg, dbPool)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create server")
		return 1
	}

	listenDone := make(chan error, 1)
	go func() {
		listenDone <- srv.Start(cfg.Server.Port)
	}()

	log.Info().Int("port", cfg.Server.Port).Msg("Server started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-listenDone:
		if err != nil {
			log.Error().Err(err).Msg("Failed to start server")
			return 1
		}
		log.Error().Msg("Server stopped unexpectedly")
		return 1
	case <-quit:
	}

	log.Info().Msg("Shutting down server...")

	if err := srv.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		return 1
	}

	if err := <-listenDone; err != nil {
		log.Error().Err(err).Msg("Server exit error")
		return 1
	}

	log.Info().Msg("Server exited")
	return 0
}
