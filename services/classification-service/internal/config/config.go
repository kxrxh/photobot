package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/CsortTeam/envkeys"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type DatabaseConfig struct {
	Host     string `env:"HOST, default=localhost"         validate:"required"`
	Port     int    `env:"PORT, default=5432"              validate:"required,gte=1,lte=65535"`
	User     string `env:"USER, default=postgres"`
	Password string `env:"PASSWORD, default=postgres"` //nolint:gosec // G117: config value from env
	Name     string `env:"NAME, default=classification_db" validate:"required"`
	SSLMode  string `env:"SSLMODE, default=disable"`
	MaxConns int32  `env:"MAX_CONNS, default=50"           validate:"gte=1,lte=200"`
	MinConns int32  `env:"MIN_CONNS, default=2"            validate:"gte=0,lte=200"`
}

type ServerConfig struct {
	Port int `env:"SERVER_PORT, default=6020" validate:"required,gte=1,lte=65535"`
}

type SecurityConfig struct {
	ServiceID     string `env:"SERVICE_ID, default=classification-service" validate:"required"`
	ServiceSecret string `env:"SERVICE_SECRET, required"                   validate:"required"`
}

type RedisConfig struct {
	Host     string `env:"HOST, default=localhost" validate:"required"`
	Port     int    `env:"PORT, default=6379"      validate:"required,gte=1,lte=65535"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB, default=0"`
}

type RateLimitConfig struct {
	Max               int `env:"MAX, default=120"           validate:"gte=1,lte=100000"`
	ExpirationSeconds int `env:"EXPIRATION_SEC, default=60" validate:"gte=1,lte=86400"`
}

func (r RateLimitConfig) Expiration() time.Duration {
	return time.Duration(r.ExpirationSeconds) * time.Second
}

type ObservabilityConfig struct {
	ServiceName        string `env:"OTEL_SERVICE_NAME, default=classification-service"`
	Endpoint           string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ResourceAttributes string `env:"OTEL_RESOURCE_ATTRIBUTES, default=service.namespace=photobot"`
	SDKDisabled        bool   `env:"OTEL_SDK_DISABLED, default=false"`
}

type HTTPConfig struct {
	TrustProxy          bool `env:"TRUST_PROXY, default=true"`
	TrustPrivateProxies bool `env:"TRUST_PRIVATE_PROXIES, default=true"`
}

type Config struct {
	Server                ServerConfig        `env:", prefix="`
	Database              DatabaseConfig      `env:", prefix=DB_"`
	Security              SecurityConfig      `env:", prefix="`
	Redis                 RedisConfig         `env:", prefix=REDIS_"`
	RateLimit             RateLimitConfig     `env:", prefix=RATE_LIMIT_"`
	HTTP                  HTTPConfig          `env:", prefix=HTTP_"`
	Observability         ObservabilityConfig `env:", prefix="`
	Debug                 bool                `env:"DEBUG, default=false"`
	AuthServiceURL        string              `env:"AUTH_SERVICE_URL, default=http://localhost:8080/api/v1"`
	CorrelationServiceURL string              `env:"CORRELATION_SERVICE_URL, default=http://localhost:8000"`
}

func LoadConfig() (*Config, error) {
	ctx := context.Background()

	if _, err := os.Stat("compose.env"); err == nil {
		if err := godotenv.Load("compose.env"); err != nil {
			return nil, fmt.Errorf("failed to load compose.env: %w", err)
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
		url.QueryEscape(c.Database.User),
		url.QueryEscape(c.Database.Password),
		url.QueryEscape(c.Database.Host),
		c.Database.Port,
		url.QueryEscape(c.Database.Name),
		url.QueryEscape(c.Database.SSLMode),
	)
}

func EnvKeys() []string {
	return envkeys.FromStruct(Config{})
}

func TestConfig(
	dbHost string,
	dbPort int,
	dbUser, dbPassword, dbName, identityURL, correlationURL string,
) *Config {
	return &Config{
		Server: ServerConfig{Port: 0},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			SSLMode:  "disable",
			MaxConns: 2,
			MinConns: 0,
		},
		Security: SecurityConfig{
			ServiceID:     "classification-test-service",
			ServiceSecret: "classification-test-secret",
		},
		RateLimit: RateLimitConfig{
			Max:               120,
			ExpirationSeconds: 60,
		},
		HTTP: HTTPConfig{
			TrustProxy:          true,
			TrustPrivateProxies: true,
		},
		AuthServiceURL:        identityURL,
		CorrelationServiceURL: correlationURL,
		Debug:                 false,
	}
}
