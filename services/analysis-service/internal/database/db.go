package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cfg *DatabaseConfig) (*DB, error) {
	pool, err := connect(ctx, *cfg)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func connect(ctx context.Context, dbCfg DatabaseConfig) (*pgxpool.Pool, error) {
	sslMode := dbCfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbCfg.User, dbCfg.Password),
		Host:     fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port),
		Path:     "/" + dbCfg.Name,
		RawQuery: "sslmode=" + url.QueryEscape(sslMode),
	}
	dsn := u.String()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolCfg.MaxConns = dbCfg.MaxConns
	poolCfg.MinConns = dbCfg.MinConns
	poolCfg.MaxConnLifetime = time.Duration(dbCfg.MaxConnLifetime) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(dbCfg.MaxConnIdleTime) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.Pool == nil {
		return errors.New("database not initialized")
	}
	return db.Pool.Ping(ctx)
}
