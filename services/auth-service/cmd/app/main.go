package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/devseed"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/observability"
	"csort.ru/auth-service/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// fatalOtel logs the error, runs OpenTelemetry shutdown (log.Fatal skips defers), then exits.
func fatalOtel(
	log zerolog.Logger,
	otelShutdown func(context.Context) error,
	err error,
	msg string,
) {
	log.Error().Err(err).Msg(msg)
	if otelShutdown != nil {
		_ = otelShutdown(context.Background())
	}
	os.Exit(1)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger.InitLogger()
	log := logger.GetLogger("main")

	otelShutdown, err := observability.InitOTEL(context.Background(), "auth-service")
	if err != nil {
		log.Fatal().Err(err).Msg("otel init failed")
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	log.Info().Msg("Starting Auth Service...")

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		fatalOtel(log, otelShutdown, err, "Failed to parse database config")
	}
	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		fatalOtel(log, otelShutdown, err, "Failed to connect to database")
	}
	if err := dbPool.Ping(context.Background()); err != nil {
		dbPool.Close()
		fatalOtel(log, otelShutdown, err, "Failed to ping database")
	}
	log.Info().Msg("Database connection established")

	redisOpts := &redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		ConnMaxIdleTime: 5 * time.Minute,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PoolTimeout:     4 * time.Second,
	}
	if cfg.Redis.PoolSize > 0 {
		redisOpts.PoolSize = cfg.Redis.PoolSize
	}
	redisClient := redis.NewClient(redisOpts)
	if result := redisClient.Ping(context.Background()); result.Err() != nil {
		_ = redisClient.Close()
		dbPool.Close()
		fatalOtel(log, otelShutdown, result.Err(), "Failed to ping Redis")
	}
	log.Info().Msg("Redis connection established")

	keyManager := auth.GetKeyManager()
	if err := keyManager.Initialize(
		cfg.Security.RSAPrivateKeyPath,
		cfg.Security.RSAPublicKeyPath,
	); err != nil {
		fatalOtel(log, otelShutdown, err, "Failed to initialize JWT key manager")
	}
	log.Info().Msg("JWT key manager initialized")

	if cfg.DevMode {
		log.Warn().Msg("DEV_MODE enabled: signature bypass, relaxed rate limits, dev seed")
		if cfg.RateLimit.Max < 10000 {
			cfg.RateLimit.Max = 100000
		}
		dbQueries := database.New(dbPool)
		botSvc, botErr := bot.NewService(dbQueries, cfg.Security.EncryptionKey)
		if botErr != nil {
			log.Warn().Err(botErr).Msg("dev bot service init failed")
		} else if seedErr := devseed.Run(context.Background(), dbQueries, botSvc); seedErr != nil {
			log.Warn().Err(seedErr).Msg("dev seed failed")
		}
	}

	srv, err := server.New(cfg, dbPool, redisClient)
	if err != nil {
		dbPool.Close()
		_ = redisClient.Close()
		fatalOtel(log, otelShutdown, err, "Failed to initialize server")
	}
	defer dbPool.Close()
	defer func() { _ = redisClient.Close() }()

	serverDone := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("Server goroutine panicked")
				serverDone <- fmt.Errorf("panic: %v", r)
			}
		}()
		if cfg.Debug {
			log.Warn().Msg("Server started in debug mode.")
		}
		if err := srv.Start(cfg.Server.Port); err != nil {
			log.Error().Err(err).Msg("Server Listen returned")
			serverDone <- err
			return
		}
		serverDone <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("Shutting down server")
	case err := <-serverDone:
		if err != nil {
			log.Error().Err(err).Msg("Server stopped unexpectedly")
			_ = srv.Shutdown()
			dbPool.Close()
			_ = redisClient.Close()
			os.Exit(1) //nolint:gocritic // exitAfterDefer: cleanup done explicitly before exit
		}
		log.Info().Msg("Server stopped")
		return
	}

	if err := srv.Shutdown(); err != nil {
		log.Warn().Err(err).Msg("Server shutdown error")
	}
	log.Info().Msg("Server stopped")
}
