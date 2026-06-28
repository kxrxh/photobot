package testing

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"csort.ru/auth-service/internal/config"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// TB is a minimal interface satisfied by *testing.T for use with gomock and test helpers.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

const (
	pgDB       = "auth_test"
	pgUser     = "postgres"
	pgPassword = "postgres"
)

// schemaPath returns the path to database/schema.sql. It searches from cwd and parent dirs.
func schemaPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return "database/schema.sql"
	}
	for {
		try := filepath.Join(dir, "database", "schema.sql")
		if _, err := os.Stat(try); err == nil {
			return try
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "database/schema.sql"
}

// SetupPostgres starts a PostgreSQL container with schema and returns a connection pool.
func SetupPostgres(ctx context.Context, t TB) (*pgxpool.Pool, func()) {
	t.Helper()
	pool, cleanup, err := setupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	return pool, cleanup
}

func setupPostgres(ctx context.Context) (pool *pgxpool.Pool, cleanup func(), err error) {
	schema := schemaPath()
	initScripts := []string{schema}

	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase(pgDB),
		postgres.WithUsername(pgUser),
		postgres.WithPassword(pgPassword),
		postgres.WithOrderedInitScripts(initScripts...),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres run: %w", err)
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("postgres connection string: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("parse postgres config: %w", err)
	}

	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("postgres connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("postgres ping: %w", err)
	}

	cleanup = func() {
		pool.Close()
		_ = testcontainers.TerminateContainer(ctr)
	}

	return pool, cleanup, nil
}

// SetupRedis starts a Redis container and returns a redis client.
func SetupRedis(ctx context.Context, t TB) (*redis.Client, func()) {
	t.Helper()
	client, cleanup, err := setupRedis(ctx)
	if err != nil {
		t.Fatalf("failed to start redis: %v", err)
	}
	return client, cleanup
}

func setupRedis(ctx context.Context) (client *redis.Client, cleanup func(), err error) {
	ctr, err := tcredis.Run(ctx, "redis:8-alpine")
	if err != nil {
		return nil, nil, fmt.Errorf("redis run: %w", err)
	}

	endpoint, err := ctr.Endpoint(ctx, "")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("redis endpoint: %w", err)
	}

	client = redis.NewClient(&redis.Options{Addr: endpoint})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("redis ping: %w", err)
	}

	cleanup = func() {
		_ = client.Close()
		_ = testcontainers.TerminateContainer(ctr)
	}

	return client, cleanup, nil
}

func parseRedisEndpoint(endpoint string) (host string, port int) {
	h, pStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		parts := strings.Split(endpoint, ":")
		if len(parts) >= 2 {
			host = parts[0]
			if p, e := strconv.Atoi(parts[1]); e == nil {
				if host == "" {
					host = "localhost"
				}
				return host, p
			}
		}
		return "localhost", 6379
	}
	host = h
	port = 6379
	if p, e := strconv.Atoi(pStr); e == nil {
		port = p
	}
	if host == "" {
		host = "localhost"
	}
	return host, port
}

func SetupPostgresAndRedis(ctx context.Context, t TB) (*pgxpool.Pool, *redis.Client, func()) {
	t.Helper()

	pool, cleanupPG, err := setupPostgres(ctx)
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	client, cleanupRedis, err := setupRedis(ctx)
	if err != nil {
		cleanupPG()
		t.Fatalf("redis: %v", err)
	}

	cleanup := func() {
		cleanupRedis()
		cleanupPG()
	}

	return pool, client, cleanup
}

// SetupPostgresAndRedisNoTB is for package TestMain-style setup when *testing.T is unavailable.
func SetupPostgresAndRedisNoTB(ctx context.Context) (*pgxpool.Pool, *redis.Client, func(), error) {
	pool, cleanupPG, err := setupPostgres(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	client, cleanupRedis, err := setupRedis(ctx)
	if err != nil {
		cleanupPG()
		return nil, nil, nil, err
	}
	cleanup := func() {
		cleanupRedis()
		cleanupPG()
	}
	return pool, client, cleanup, nil
}

// ConfigFor builds a *config.Config from Postgres pool, Redis client, and keys directory.
// Use this instead of config.TestConfig so test helpers stay in internal/testing.
func ConfigFor(pool *pgxpool.Pool, redisClient *redis.Client, keysDir string) *config.Config {
	connConfig := pool.Config().ConnConfig
	host := connConfig.Host
	if host == "" {
		host = "localhost"
	}
	port := int(connConfig.Port)
	if port == 0 {
		port = 5432
	}
	user := connConfig.User
	if user == "" {
		user = "postgres"
	}
	password := connConfig.Password
	dbName := connConfig.Database
	if dbName == "" {
		dbName = pgDB
	}

	redisHost, redisPort := parseRedisEndpoint(redisClient.Options().Addr)

	return &config.Config{
		Server: config.ServerConfig{Port: 0},
		Database: config.DatabaseConfig{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			Name:     dbName,
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     redisHost,
			Port:     redisPort,
			Password: "",
			DB:       0,
		},
		Security: config.SecurityConfig{
			AdminLogin:           "admin",
			AdminPassword:        hashedTestPassword(),
			AccessExpiryMinutes:  15,
			RefreshExpiryMinutes: 10080,
			EncryptionKey:        "0123456789abcdef0123456789abcdef",
			RSAPrivateKeyPath:    filepath.Join(keysDir, "private.pem"),
			RSAPublicKeyPath:     filepath.Join(keysDir, "public.pem"),
		},
		Merge:     config.MergeConfig{},
		RateLimit: config.RateLimitConfig{Max: 100000, ExpirationSeconds: 60},
		Debug:     false,
	}
}

func hashedTestPassword() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test password hash: %v", err))
	}
	return string(hash)
}
