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

type DBConfig struct {
	Host            string `env:"HOST, default=localhost"         validate:"required"`
	Port            int    `env:"PORT, default=5432"              validate:"required,gte=1,lte=65535"`
	User            string `env:"USER, default=user"`
	Password        string `env:"PASSWORD, default=password"`
	Name            string `env:"NAME, default=db"                validate:"required"`
	SSLMode         string `env:"SSL_MODE, default=disable"`
	MaxConns        int32  `env:"MAX_CONNS, default=10"           validate:"gte=1"`
	MinConns        int32  `env:"MIN_CONNS, default=2"            validate:"gte=0"`
	MaxConnLifetime int64  `env:"MAX_CONN_LIFETIME, default=3600"`
	MaxConnIdleTime int64  `env:"MAX_CONN_IDLE_TIME, default=300"`
}

type RabbitMQConfig struct {
	URL               string `env:"URL"` // when set, used directly (amqp:// or amqps://)
	Host              string `env:"HOST, default=localhost"`
	Port              int    `env:"PORT, default=5672"                         validate:"required,gte=1,lte=65535"`
	User              string `env:"USER, default=admin"`
	Password          string `env:"PASSWORD"`
	VHost             string `env:"VHOST, default=/"`
	RequestExchange   string `env:"REQUEST_EXCHANGE, default=request_exchange"`
	RequestQueue      string `env:"REQUEST_QUEUE, default=request_queue"`
	RequestRoutingKey string `env:"REQUEST_ROUTING_KEY, default=request"`
}

type RedisConfig struct {
	Host     string `env:"HOST, default=localhost" validate:"required"`
	Port     int    `env:"PORT, default=6379"      validate:"required,gte=1,lte=65535"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB, default=0"`
}

type ImageStorageConfig struct {
	Host           string `env:"HOST, default=minio-analysis" validate:"required"`
	Port           int    `env:"PORT, default=9000"           validate:"required,gte=1,lte=65535"`
	RootUser       string `env:"ROOT_USER"`
	RootPassword   string `env:"ROOT_PASSWORD"`
	TempBucket     string `env:"TEMP_BUCKET"                  validate:"required"`
	AnalysisBucket string `env:"ANALYSIS_BUCKET"              validate:"required"`
	UseSSL         bool   `env:"USE_SSL, default=false"`
	ExternalHost   string `env:"EXTERNAL_HOST"`
}

type AppConfig struct {
	Port              int `env:"PORT, default=6040"              validate:"required,gte=1,lte=65535"`
	MaxQueuedRequests int `env:"MAX_QUEUED_REQUESTS, default=10" validate:"required,gte=1"`
}

type OutboxRelayConfig struct {
	BatchSize            int `env:"BATCH_SIZE, default=10"            validate:"gte=1,lte=100"`
	PollIntervalSec      int `env:"POLL_INTERVAL_SECONDS, default=1"  validate:"gte=1"`
	CleanupIntervalHr    int `env:"CLEANUP_INTERVAL_HOURS, default=1" validate:"gte=1"`
	MessageTTLHours      int `env:"MESSAGE_TTL_HOURS, default=24"     validate:"gte=1"`
	MarkPublishedRetries int `env:"MARK_PUBLISHED_RETRIES, default=3" validate:"gte=1,lte=10"`
}

type SecurityConfig struct {
	ServiceUrl          string        `env:"AUTH_SERVICE_URL, required"        validate:"required,url"`
	ServiceId           string        `env:"SERVICE_ID, required"              validate:"required"`
	ServiceSecret       string        `env:"SERVICE_SECRET, required"          validate:"required"`
	JWKSRefreshInterval time.Duration `env:"JWKS_REFRESH_INTERVAL, default=1h"`
}

type ShareLinkConfig struct {
	HMACSecret      string `env:"HMAC_SECRET, required"          validate:"required"`
	MaxClockSkewSec int    `env:"MAX_CLOCK_SKEW_SEC, default=60" validate:"gte=0,lte=3600"`
}

type OTELConfig struct {
	ServiceName          string `env:"OTEL_SERVICE_NAME"`
	ExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ExporterOTLPProtocol string `env:"OTEL_EXPORTER_OTLP_PROTOCOL, default=http/protobuf"`
	ResourceAttributes   string `env:"OTEL_RESOURCE_ATTRIBUTES"`
	SDKDisabled          bool   `env:"OTEL_SDK_DISABLED, default=false"`
}

type Config struct {
	KalibrDB          DBConfig           `env:", prefix=DB_"`
	App               AppConfig          `env:", prefix="`
	RabbitMQ          RabbitMQConfig     `env:", prefix=RABBITMQ_"`
	OutboxRelay       OutboxRelayConfig  `env:", prefix=OUTBOX_RELAY_"`
	Redis             RedisConfig        `env:", prefix=REDIS_"`
	RequestsDB        DBConfig           `env:", prefix=REQUESTS_DB_"`
	ImageStorage      ImageStorageConfig `env:", prefix=MINIO_IMAGES_"`
	AnalysisAPI       string             `env:"ANALYSIS_API_URL, required"           validate:"required,url"`
	ReportsAPI        string             `env:"REPORTS_API_URL, required"            validate:"required,url"`
	ClassificationAPI string             `env:"CLASSIFICATION_SERVICE_URL, required" validate:"required,url"`
	Security          SecurityConfig     `env:", prefix="`
	ShareLink         ShareLinkConfig    `env:", prefix=SHARE_LINK_"`
	OTEL              OTELConfig         `env:", prefix="`
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

	v := validator.New()
	if err := v.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		u, err := url.Parse(fl.Field().String())
		return err == nil && u.Scheme != "" && u.Host != ""
	}); err != nil {
		return nil, fmt.Errorf("register url validator: %w", err)
	}
	if err := v.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func EnvKeys() []string {
	keys := envkeys.FromStruct(Config{})
	filtered := make([]string, 0, len(keys))
	for _, k := range keys {
		if k == "-" {
			continue
		}
		filtered = append(filtered, k)
	}
	return filtered
}
