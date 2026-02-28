package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// NewRedis creates a Redis client and verifies the connection.
func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}
