package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type DatabaseConfig struct {
	Host            string        `env:"HOST, default=localhost"        validate:"required"`
	Port            int           `env:"PORT, default=5432"             validate:"required,gte=1,lte=65535"`
	User            string        `env:"USER, default=user"`
	Password        string        `env:"PASSWORD, default=password"` //nolint:gosec // G117: config struct for DB credentials
	Name            string        `env:"NAME, default=db"               validate:"required"`
	SSLMode         string        `env:"SSLMODE, default=disable"`
	MaxConns        int32         `env:"MAX_CONNS, default=10"`
	MinConns        int32         `env:"MIN_CONNS, default=2"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME, default=1h"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME, default=5m"`
}

type MinioConfig struct {
	Host           string `env:"HOST, default=localhost"  validate:"required"`
	Port           int    `env:"PORT, default=9000"       validate:"required,gte=1,lte=65535"`
	AccessKey      string `env:"ROOT_USER"                validate:"required"` //nolint:gosec // G117: config struct for MinIO credentials
	SecretKey      string `env:"ROOT_PASSWORD"            validate:"required"`
	Bucket         string `env:"BUCKET, default=photobot" validate:"required"`
	UseSSL         bool   `env:"USE_SSL, default=false"`
	PublicEndpoint string `env:"PUBLIC_ENDPOINT"`
}

type ServerConfig struct {
	Port int `env:"SERVER_PORT, default=8000" validate:"required,gte=1,lte=65535"`
}

type IdentityConfig struct {
	ServiceURL string `env:"AUTH_SERVICE_URL, default=http://localhost:8080/api/v1" validate:"required,url"`
}

type RedisConfig struct {
	Host     string `env:"HOST, default=localhost" validate:"required"`
	Port     int    `env:"PORT, default=6379"      validate:"required,gte=1,lte=65535"`
	Password string `env:"PASSWORD"` //nolint:gosec // G117: config struct for Redis credentials
	DB       int    `env:"DB, default=0"`
}

type RateLimitConfig struct {
	Max               int `env:"MAX, default=120"           validate:"gte=1,lte=100000"`
	ExpirationSeconds int `env:"EXPIRATION_SEC, default=60" validate:"gte=1,lte=86400"`
}

func (r RateLimitConfig) Expiration() time.Duration {
	return time.Duration(r.ExpirationSeconds) * time.Second
}

type Config struct {
	Server     ServerConfig    `env:", prefix="`
	InternalDB DatabaseConfig  `env:", prefix=DB_"`
	Minio      MinioConfig     `env:", prefix=MINIO_"`
	Identity   IdentityConfig  `env:", prefix="`
	Redis      RedisConfig     `env:", prefix=REDIS_"`
	RateLimit  RateLimitConfig `env:", prefix=RATE_LIMIT_"`
	Debug      bool            `env:"DEBUG, default=false"`
}

func LoadConfig() (*Config, error) {
	ctx := context.Background()

	if _, err := os.Stat(".env.dev"); err == nil {
		if err := godotenv.Load(".env.dev"); err != nil {
			return nil, fmt.Errorf("failed to load .env.dev: %w", err)
		}
	}

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.InternalDB.User),
		url.QueryEscape(c.InternalDB.Password),
		url.QueryEscape(c.InternalDB.Host),
		c.InternalDB.Port,
		url.QueryEscape(c.InternalDB.Name),
		url.QueryEscape(c.InternalDB.SSLMode),
	)
}

// DSN returns the data source name for the database connection.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		url.QueryEscape(c.Host),
		c.Port,
		url.QueryEscape(c.Name),
		url.QueryEscape(c.SSLMode),
	)
}
