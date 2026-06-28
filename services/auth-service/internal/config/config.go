package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/CsortTeam/envkeys"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type DatabaseConfig struct {
	Host     string `env:"HOST, default=localhost"    validate:"required"`
	Port     int    `env:"PORT, default=5432"         validate:"required,gte=1,lte=65535"`
	User     string `env:"USER, default=postgres"`
	Password string `env:"PASSWORD, default=postgres"`
	Name     string `env:"NAME, default=authservice"  validate:"required"`
	SSLMode  string `env:"SSLMODE, default=disable"`
	MaxConns int32  `env:"MAX_CONNS, default=25"      validate:"gte=1,lte=200"`
	MinConns int32  `env:"MIN_CONNS, default=0"       validate:"gte=0,lte=200"`
}

type ServerConfig struct {
	Host              string `env:"HOST, default=0.0.0.0"`
	Port              int    `env:"PORT, default=8080"              validate:"required,gte=1,lte=65535"`
	HealthCacheTTLSec int    `env:"HEALTH_CACHE_TTL_SEC, default=0" validate:"gte=0,lte=60"`
	CORSAllowOrigins  string `env:"CORS_ALLOW_ORIGINS, default=*"`
}

type RedisConfig struct {
	Host     string `env:"HOST, default=localhost" validate:"required"`
	Port     int    `env:"PORT, default=6379"      validate:"required,gte=1,lte=65535"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB, default=0"`
	PoolSize int    `env:"POOL_SIZE, default=0"    validate:"gte=0,lte=500"`
}

type RateLimitConfig struct {
	Max               int `env:"MAX, default=30"`
	ExpirationSeconds int `env:"EXPIRATION_SEC, default=60"`
}

func (r RateLimitConfig) Expiration() time.Duration {
	if r.ExpirationSeconds <= 0 {
		return time.Minute
	}
	return time.Duration(r.ExpirationSeconds) * time.Second
}

type SecurityConfig struct {
	AdminLogin            string `env:"ADMIN_LOGIN, required"                                            validate:"required"`
	AdminPassword         string `env:"ADMIN_PASSWORD, required"                                         validate:"required"`
	AccessExpiryMinutes   int    `env:"JWT_ACCESS_EXPIRY_MINUTES, default=15"`
	RefreshExpiryMinutes  int    `env:"JWT_REFRESH_EXPIRY_MINUTES, default=43200"`
	EncryptionKey         string `env:"ENCRYPTION_KEY, required"                                         validate:"required"`
	RSAPrivateKeyPath     string `env:"RSA_PRIVATE_KEY_PATH, default=/etc/auth-service/keys/private.pem"`
	RSAPublicKeyPath      string `env:"RSA_PUBLIC_KEY_PATH, default=/etc/auth-service/keys/public.pem"`
	DebugBypassSignatures bool   `env:"DEBUG_BYPASS_SIGNATURES, default=false"`
}

type MergeConfig struct {
	ClassificationServiceURL string `env:"MERGE_CLASSIFICATION_SERVICE_URL"`
	AnalysisServiceURL       string `env:"MERGE_ANALYSIS_SERVICE_URL"`
}

type OTELConfig struct {
	ServiceName          string `env:"OTEL_SERVICE_NAME, default=auth-service"`
	ExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ExporterOTLPProtocol string `env:"OTEL_EXPORTER_OTLP_PROTOCOL, default=http/protobuf"`
	ResourceAttributes   string `env:"OTEL_RESOURCE_ATTRIBUTES, default=service.namespace=photobot"`
	SDKDisabled          bool   `env:"OTEL_SDK_DISABLED, default=false"`
}

type Config struct {
	Server             ServerConfig    `env:", prefix="`
	Database           DatabaseConfig  `env:", prefix=DB_"`
	Redis              RedisConfig     `env:", prefix=REDIS_"`
	Security           SecurityConfig  `env:", prefix="`
	Merge              MergeConfig     `env:", prefix="`
	RateLimit          RateLimitConfig `env:", prefix=RATE_LIMIT_"`
	OTEL               OTELConfig      `env:", prefix="`
	Debug              bool            `env:"DEBUG, default=false"`
	DevMode            bool            `env:"DEV_MODE, default=false"`
	ResetOTPTTLSeconds int             `env:"RESET_OTP_TTL_SECONDS, default=600"`
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
		return nil, formatConfigError(err)
	}

	return &cfg, nil
}

func formatConfigError(err error) error {
	if err == nil {
		return nil
	}
	var missing, invalid []string
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			envKey := envkeys.NamespaceToKey(Config{}, "Config", e.Namespace())
			if e.Tag() == "required" {
				missing = append(missing, envKey)
			} else {
				invalid = append(invalid, fmt.Sprintf("%s: %s", envKey, e.Tag()))
			}
		}
	} else {
		return fmt.Errorf("config validation failed: %w", err)
	}
	sort.Strings(missing)
	sort.Strings(invalid)

	var b strings.Builder
	b.WriteString("config validation failed")
	if len(missing) > 0 {
		b.WriteString(": missing required keys ")
		b.WriteString(strings.Join(missing, ", "))
		b.WriteString(
			". Set them in compose.env (local) or COMPOSE_ENVS auth-service block (deploy)",
		)
	}
	if len(invalid) > 0 {
		if len(missing) > 0 {
			b.WriteString("; ")
		}
		b.WriteString(strings.Join(invalid, "; "))
	}
	return fmt.Errorf("%s", b.String())
}

func EnvKeys() []string {
	return envkeys.FromStruct(Config{})
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
