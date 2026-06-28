package main

import (
	"context"
	"fmt"
	"os"

	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/devseed"
	"csort.ru/auth-service/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println("dev seed complete")
}

func run() error {
	logger.InitLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer pool.Close()

	db := database.New(pool)
	botSvc, err := bot.NewService(db, cfg.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("bot service: %w", err)
	}

	if err := devseed.Run(context.Background(), db, botSvc); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	return nil
}
