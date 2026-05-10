package storage

import (
	"context"
	"io"
	"time"
)

// DefaultThumbnailURLExpiry is the presigned URL lifetime for thumbnails.
const DefaultThumbnailURLExpiry = 7 * 24 * time.Hour

// ObjectInfo describes a stored object's metadata.
type ObjectInfo struct {
	Size        int64
	ContentType string
}

// StorageService defines the interface for file storage.
type StorageService interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
	Move(ctx context.Context, srcKey, dstKey string) error
	PublicURL(key string) string
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	// PresignedPutURL returns a URL the client can PUT to directly. The client
	// must send the same Content-Type header in the PUT request.
	PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	// Stat returns object metadata, or (nil, nil) if the object does not exist.
	Stat(ctx context.Context, key string) (*ObjectInfo, error)
}
