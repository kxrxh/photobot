package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig
	AnalysisService AnalysisServiceConfig
	Auth            AuthConfig
	Redis           RedisConfig
	RateLimit       RateLimitConfig
	MinIO           MinIOConfig
	Templates       TemplatesConfig
	ReportPack      ReportPackConfig
	ReportFormat    ReportFormatConfig
}

type ServerConfig struct {
	Port             int
	CORSAllowOrigins string
	TrustProxy       bool
	PublicBaseURL    string
}

type AnalysisServiceConfig struct {
	Host string
}

type AuthConfig struct {
	ServiceURL          string
	ServiceID           string
	ServiceSecret       string
	JWKSRefreshInterval time.Duration
}

// RedisConfig holds optional Redis settings for distributed rate limiting.
// When URL is empty, the rate limiter uses in-process memory (not shared across replicas).
type RedisConfig struct {
	URL string
}

type RateLimitConfig struct {
	// Global applies to every authenticated /api/reports request (PUT and GET).
	GlobalMax    int
	GlobalWindow time.Duration

	// Generate covers PUT /api/reports/:id only (expensive report generation),
	// in addition to Global.
	GenerateMax    int
	GenerateWindow time.Duration

	// Download covers GET .../download-url only, in addition to Global.
	DownloadMax    int
	DownloadWindow time.Duration
}

type MinIOConfig struct {
	Host          string
	Port          int
	RootUser      string
	RootPassword  string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string
}

type TemplatesConfig struct {
	Dir string
}

type ReportPackConfig struct {
	HMACSecret string
	TTLSeconds int
}

type ReportFormatConfig struct {
	CSV string
	PDF string
}

func Load() (Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("PORT", 6050)
	viper.SetDefault("TRUST_PROXY", false)
	viper.SetDefault("MINIO_HOST", "localhost")
	viper.SetDefault("MINIO_PORT", 9000)
	viper.SetDefault("MINIO_ROOT_USER", "minioadmin")
	viper.SetDefault("MINIO_ROOT_PASSWORD", "minioadmin")
	viper.SetDefault("MINIO_BUCKET", "reports")
	viper.SetDefault("MINIO_USE_SSL", false)
	viper.SetDefault("TEMPLATES_DIR", "./templates")
	viper.SetDefault("REPORT_PACK_TTL_SECONDS", 604800)
	viper.SetDefault("REPORT_CSV_FORMAT_VERSION", "1.0-go")
	viper.SetDefault("REPORT_PDF_FORMAT_VERSION", "1.0-go")
	viper.SetDefault("JWKS_REFRESH_INTERVAL", "1h")
	viper.SetDefault("RATE_LIMIT_GLOBAL_MAX", 200)
	viper.SetDefault("RATE_LIMIT_GLOBAL_WINDOW", "1m")
	viper.SetDefault("RATE_LIMIT_GENERATE_MAX", 15)
	viper.SetDefault("RATE_LIMIT_GENERATE_WINDOW", "1h")
	viper.SetDefault("RATE_LIMIT_DOWNLOAD_MAX", 120)
	viper.SetDefault("RATE_LIMIT_DOWNLOAD_WINDOW", "1m")

	csvFmt := strings.TrimSpace(viper.GetString("REPORT_CSV_FORMAT_VERSION"))
	if csvFmt == "" {
		csvFmt = "1.0-go"
	}
	pdfFmt := strings.TrimSpace(viper.GetString("REPORT_PDF_FORMAT_VERSION"))
	if pdfFmt == "" {
		pdfFmt = "1.0-go"
	}

	cfg := Config{
		Server: ServerConfig{
			Port:             viper.GetInt("PORT"),
			CORSAllowOrigins: viper.GetString("CORS_ALLOW_ORIGINS"),
			TrustProxy:       viper.GetBool("TRUST_PROXY"),
			PublicBaseURL: strings.TrimSuffix(
				strings.TrimSpace(viper.GetString("REPORTS_PUBLIC_BASE_URL")),
				"/",
			),
		},
		AnalysisService: AnalysisServiceConfig{
			Host: strings.TrimSuffix(viper.GetString("ANALYSIS_SERVICE_HOST"), "/"),
		},
		Auth: AuthConfig{
			ServiceURL:          strings.TrimSuffix(viper.GetString("AUTH_SERVICE_URL"), "/"),
			ServiceID:           strings.TrimSpace(viper.GetString("REPORTS_SERVICE_ID")),
			ServiceSecret:       strings.TrimSpace(viper.GetString("REPORTS_SERVICE_SECRET")),
			JWKSRefreshInterval: viper.GetDuration("JWKS_REFRESH_INTERVAL"),
		},
		Redis: RedisConfig{
			URL: strings.TrimSpace(viper.GetString("REDIS_URL")),
		},
		RateLimit: RateLimitConfig{
			GlobalMax:      viper.GetInt("RATE_LIMIT_GLOBAL_MAX"),
			GlobalWindow:   viper.GetDuration("RATE_LIMIT_GLOBAL_WINDOW"),
			GenerateMax:    viper.GetInt("RATE_LIMIT_GENERATE_MAX"),
			GenerateWindow: viper.GetDuration("RATE_LIMIT_GENERATE_WINDOW"),
			DownloadMax:    viper.GetInt("RATE_LIMIT_DOWNLOAD_MAX"),
			DownloadWindow: viper.GetDuration("RATE_LIMIT_DOWNLOAD_WINDOW"),
		},
		MinIO: MinIOConfig{
			Host:          viper.GetString("MINIO_HOST"),
			Port:          viper.GetInt("MINIO_PORT"),
			RootUser:      viper.GetString("MINIO_ROOT_USER"),
			RootPassword:  viper.GetString("MINIO_ROOT_PASSWORD"),
			Bucket:        viper.GetString("MINIO_BUCKET"),
			UseSSL:        viper.GetBool("MINIO_USE_SSL"),
			PublicBaseURL: strings.TrimSpace(viper.GetString("MINIO_PUBLIC_BASE_URL")),
		},
		Templates: TemplatesConfig{
			Dir: viper.GetString("TEMPLATES_DIR"),
		},
		ReportPack: ReportPackConfig{
			HMACSecret: strings.TrimSpace(viper.GetString("REPORT_PACK_HMAC_SECRET")),
			TTLSeconds: viper.GetInt("REPORT_PACK_TTL_SECONDS"),
		},
		ReportFormat: ReportFormatConfig{
			CSV: csvFmt,
			PDF: pdfFmt,
		},
	}

	if cfg.AnalysisService.Host == "" {
		return Config{}, errors.New("ANALYSIS_SERVICE_HOST is required")
	}
	if cfg.Auth.ServiceURL == "" {
		return Config{}, errors.New("AUTH_SERVICE_URL is required")
	}
	if cfg.Auth.ServiceID == "" {
		return Config{}, errors.New("REPORTS_SERVICE_ID is required")
	}
	if cfg.Auth.ServiceSecret == "" {
		return Config{}, errors.New("REPORTS_SERVICE_SECRET is required")
	}
	if cfg.Auth.JWKSRefreshInterval <= 0 {
		cfg.Auth.JWKSRefreshInterval = time.Hour
	}

	if cfg.RateLimit.GlobalWindow <= 0 {
		cfg.RateLimit.GlobalWindow = time.Minute
	}
	if cfg.RateLimit.GenerateWindow <= 0 {
		cfg.RateLimit.GenerateWindow = time.Hour
	}
	if cfg.RateLimit.DownloadWindow <= 0 {
		cfg.RateLimit.DownloadWindow = time.Minute
	}
	if cfg.RateLimit.GlobalMax < 0 {
		cfg.RateLimit.GlobalMax = 0
	}
	if cfg.RateLimit.GenerateMax < 0 {
		cfg.RateLimit.GenerateMax = 0
	}
	if cfg.RateLimit.DownloadMax < 0 {
		cfg.RateLimit.DownloadMax = 0
	}

	return cfg, nil
}
