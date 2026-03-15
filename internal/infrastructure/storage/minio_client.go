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
	client        *minio.Client
	presignClient *minio.Client // uses public endpoint for correct S3v4 signatures
	bucket        string
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

	// Create a separate client for presigning URLs using the public endpoint.
	// PresignedGetObject is a local computation (no network call), so this client
	// doesn't need to reach the public endpoint from inside Docker.
	// This ensures the S3v4 signature matches the Host header clients will send.
	presignClient := client
	if publicEndpoint != cfg.Endpoint {
		pubU, err := url.Parse(publicEndpoint)
		if err != nil {
			return nil, fmt.Errorf("parsing public endpoint: %w", err)
		}
		pc, err := minio.New(pubU.Host, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: pubU.Scheme == "https",
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("creating presign client: %w", err)
		}
		presignClient = pc
	}

	return &MinIOService{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.Bucket,
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

func (s *MinIOService) Move(ctx context.Context, srcKey, dstKey string) error {
	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: srcKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: dstKey,
	}

	if _, err := s.client.CopyObject(ctx, dst, src); err != nil {
		return fmt.Errorf("copying object from %s to %s: %w", srcKey, dstKey, err)
	}

	if err := s.client.RemoveObject(ctx, s.bucket, srcKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("removing source object %s after copy: %w", srcKey, err)
	}

	return nil
}

func (s *MinIOService) PublicURL(key string) string {
	ep := s.presignClient.EndpointURL()
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(ep.String(), "/"), s.bucket, key)
}

func (s *MinIOService) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}
