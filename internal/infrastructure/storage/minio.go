package storage

import (
	"context"
	"io"
	"time"
)

// StorageService defines the interface for file storage.
type StorageService interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
