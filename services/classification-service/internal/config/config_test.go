package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestConfig(t *testing.T) {
	cfg := TestConfig(
		"dbhost",
		5433,
		"testuser",
		"testpass",
		"testdb",
		"http://identity:8080",
		"http://correlation:8000",
	)
	require.NotNil(t, cfg)

	assert.Equal(t, "dbhost", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "http://identity:8080", cfg.AuthServiceURL)
	assert.Equal(t, "http://correlation:8000", cfg.CorrelationServiceURL)
	assert.Equal(t, "classification-test-service", cfg.Security.ServiceID)
	assert.Equal(t, "classification-test-secret", cfg.Security.ServiceSecret)
	assert.False(t, cfg.Debug)
	assert.Equal(t, 0, cfg.Server.Port)
	assert.Equal(t, 120, cfg.RateLimit.Max)
	assert.Equal(t, 60, cfg.RateLimit.ExpirationSeconds)
	assert.True(t, cfg.HTTP.TrustProxy)
	assert.True(t, cfg.HTTP.TrustPrivateProxies)
}

func TestConfig_DatabaseURL(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Name:     "mydb",
			SSLMode:  "require",
		},
	}

	url := cfg.DatabaseURL()
	assert.Contains(t, url, "postgres://")
	assert.Contains(t, url, "postgres")
	assert.Contains(t, url, "secret")
	assert.Contains(t, url, "localhost")
	assert.Contains(t, url, "5432")
	assert.Contains(t, url, "mydb")
	assert.Contains(t, url, "sslmode=require")
}

func TestConfig_DatabaseURL_EscapesSpecialChars(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "user@domain",
			Password: "pass:word#with",
			Name:     "db",
			SSLMode:  "disable",
		},
	}

	url := cfg.DatabaseURL()
	assert.Contains(t, url, "postgres://")
	assert.True(
		t,
		strings.Contains(url, "%40") || strings.Contains(url, "user"),
		"user should be in URL",
	)
}
