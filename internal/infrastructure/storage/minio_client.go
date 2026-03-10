package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// MinIOService implements StorageService using MinIO.
type MinIOService struct {
	client         *minio.Client
	bucket         string
	publicEndpoint string
}

// NewMinIOService creates a new MinIO-backed StorageService.
func NewMinIOService(cfg config.StorageConfig) (*MinIOService, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing S3 endpoint: %w", err)
	}

	endpoint := u.Host
	useSSL := u.Scheme == "https"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	return &MinIOService{
		client:         client,
		bucket:         cfg.Bucket,
		publicEndpoint: publicEndpoint,
	}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *MinIOService) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("checking bucket existence: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
	}
	return nil
}

func (s *MinIOService) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("uploading object: %w", err)
	}
	return s.PublicURL(key), nil
}

func (s *MinIOService) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}

func (s *MinIOService) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.publicEndpoint, "/"), s.bucket, key)
}

func (s *MinIOService) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}

	// Replace internal endpoint with public endpoint for external access.
	if s.publicEndpoint != "" {
		internalHost := s.client.EndpointURL().Host
		publicU, _ := url.Parse(s.publicEndpoint)
		if publicU != nil && publicU.Host != internalHost {
			result := presignedURL.String()
			result = strings.Replace(result, internalHost, publicU.Host, 1)
			internalScheme := s.client.EndpointURL().Scheme
			if internalScheme != publicU.Scheme {
				result = strings.Replace(result, internalScheme+"://", publicU.Scheme+"://", 1)
			}
			return result, nil
		}
	}

	return presignedURL.String(), nil
}
