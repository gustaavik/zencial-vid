package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
	CDN      CDNConfig
	Stripe   StripeConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5m"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10m"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"15s"`
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
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type JWTConfig struct {
	AccessSecret    string        `env:"JWT_ACCESS_SECRET,required"`
	RefreshSecret   string        `env:"JWT_REFRESH_SECRET,required"`
	AccessDuration  time.Duration `env:"JWT_ACCESS_DURATION" envDefault:"15m"`
	RefreshDuration time.Duration `env:"JWT_REFRESH_DURATION" envDefault:"168h"`
	Issuer          string        `env:"JWT_ISSUER" envDefault:"zencial"`
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
	BaseURL        string        `env:"CDN_BASE_URL"`
	SigningKeyID   string        `env:"CDN_SIGNING_KEY_ID"`
	SigningKeyPath string        `env:"CDN_SIGNING_KEY_PATH"`
	URLExpiry      time.Duration `env:"CDN_URL_EXPIRY" envDefault:"4h"`
}

type StripeConfig struct {
	SecretKey      string `env:"STRIPE_SECRET_KEY"`
	WebhookSecret  string `env:"STRIPE_WEBHOOK_SECRET"`
	PublishableKey string `env:"STRIPE_PUBLISHABLE_KEY"`
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
