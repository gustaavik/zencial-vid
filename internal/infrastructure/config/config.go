package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Session     SessionConfig
	Storage     StorageConfig
	CDN         CDNConfig
	InternalAPI InternalAPIConfig
	Stripe      StripeConfig
	Log         LogConfig
}

type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5m"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10m"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"15s"`
	AppEnv          string        `env:"APP_ENV" envDefault:"development"`
}

// Addr returns the server address in host:port format.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host         string `env:"DB_HOST" envDefault:"localhost"`
	Port         int    `env:"DB_PORT" envDefault:"5432"`
	User         string `env:"DB_USER" envDefault:"zencial"`
	Password     string `env:"DB_PASSWORD,required"`
	Name         string `env:"DB_NAME" envDefault:"zencial"`
	SSLMode      string `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// SessionConfig configures sliding-session authentication. Sessions are
// validated by hashing the bearer token and looking it up in user_sessions
// on every request; activity slides idle expiry forward, debounced by
// SlideDebounce so a single session generates at most ~1 write per debounce
// window.
type SessionConfig struct {
	IdleTimeout     time.Duration `env:"SESSION_IDLE_TIMEOUT"     envDefault:"720h"`  // 30d
	AbsoluteTimeout time.Duration `env:"SESSION_ABSOLUTE_TIMEOUT" envDefault:"2160h"` // 90d
	SlideDebounce   time.Duration `env:"SESSION_SLIDE_DEBOUNCE"   envDefault:"5m"`
	CleanupInterval time.Duration `env:"SESSION_CLEANUP_INTERVAL" envDefault:"1h"`
}

type StorageConfig struct {
	Endpoint       string `env:"S3_ENDPOINT"`
	PublicEndpoint string `env:"S3_PUBLIC_ENDPOINT"` // Externally reachable endpoint for presigned URLs (defaults to Endpoint)
	Bucket         string `env:"S3_BUCKET" envDefault:"zencial-videos"`
	Region         string `env:"S3_REGION" envDefault:"eu-west-1"`
	AccessKey      string `env:"S3_ACCESS_KEY"`
	SecretKey      string `env:"S3_SECRET_KEY"`
}

type CDNConfig struct {
	BaseURL        string        `env:"CDN_BASE_URL"`     // Public/browser-facing URL (e.g. http://localhost:8090)
	InternalURL    string        `env:"CDN_INTERNAL_URL"` // Docker-internal URL for backend→CDN calls (e.g. http://cdn:8090)
	SigningKeyID   string        `env:"CDN_SIGNING_KEY_ID"`
	SigningKeyPath string        `env:"CDN_SIGNING_KEY_PATH"`
	URLExpiry      time.Duration `env:"CDN_URL_EXPIRY" envDefault:"4h"`
	// UploadSigningKey is the shared HMAC secret used to sign CDN upload URLs.
	// Must match CDN_UPLOAD_SIGNING_KEY on the zencial-cdn service.
	UploadSigningKey string `env:"CDN_UPLOAD_SIGNING_KEY"`
	// UploadKeyID identifies which signing key version is in use, embedded in
	// the URL so future key rotation is non-breaking.
	UploadKeyID string `env:"CDN_UPLOAD_SIGNING_KEY_ID" envDefault:"v1"`
}

// InternalAPIConfig holds settings for the internal service-to-service API surface
// (e.g. CDN → API transcode callbacks).
type InternalAPIConfig struct {
	SharedSecret string `env:"INTERNAL_API_SHARED_SECRET"` // Required when CDN integration is enabled.
}

type StripeConfig struct {
	SecretKey      string `env:"STRIPE_SECRET_KEY"`
	WebhookSecret  string `env:"STRIPE_WEBHOOK_SECRET"`
	PublishableKey string `env:"STRIPE_PUBLISHABLE_KEY"`
	Currency       string `env:"STRIPE_CURRENCY" envDefault:"usd"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}
