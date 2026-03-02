package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// PresignedUpload holds the result of a presigned upload URL generation.
type PresignedUpload struct {
	URL       string
	Key       string
	PublicURL string
	ExpiresIn time.Duration
}

// StorageService defines the interface for file storage.
type StorageService interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, size int64) (*PresignedUpload, error)
}

// S3Client implements StorageService using S3-compatible storage (AWS S3 / MinIO).
type S3Client struct {
	client         *s3.Client
	presigner      *s3.PresignClient
	bucket         string
	endpoint       string
	publicEndpoint string
	cdnBase        string
}

// NewS3Client creates a new S3Client from the storage and CDN configuration.
func NewS3Client(storageCfg config.StorageConfig, cdnBaseURL string) (*S3Client, error) {
	if storageCfg.Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required for storage")
	}

	// Internal client for server-side operations.
	client := s3.New(s3.Options{
		Region:       storageCfg.Region,
		BaseEndpoint: aws.String(storageCfg.Endpoint),
		Credentials: credentials.NewStaticCredentialsProvider(
			storageCfg.AccessKey,
			storageCfg.SecretKey,
			"",
		),
		UsePathStyle: true, // Required for MinIO
	})

	// Presign client uses the public endpoint so URLs are reachable from the browser.
	publicEndpoint := storageCfg.Endpoint
	if storageCfg.PublicEndpoint != "" {
		publicEndpoint = storageCfg.PublicEndpoint
	}
	presignClient := s3.New(s3.Options{
		Region:       storageCfg.Region,
		BaseEndpoint: aws.String(publicEndpoint),
		Credentials: credentials.NewStaticCredentialsProvider(
			storageCfg.AccessKey,
			storageCfg.SecretKey,
			"",
		),
		UsePathStyle: true,
	})

	return &S3Client{
		client:         client,
		presigner:      s3.NewPresignClient(presignClient),
		bucket:         storageCfg.Bucket,
		endpoint:       strings.TrimRight(storageCfg.Endpoint, "/"),
		publicEndpoint: strings.TrimRight(publicEndpoint, "/"),
		cdnBase:        strings.TrimRight(cdnBaseURL, "/"),
	}, nil
}

// EnsureBucket creates the bucket if it does not already exist.
func (s *S3Client) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil // bucket exists
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}

	// Make bucket publicly readable (for MinIO dev).
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Sid": "PublicRead",
			"Effect": "Allow",
			"Principal": "*",
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/*"]
		}]
	}`, s.bucket)

	_, err = s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		// Non-fatal - some S3-compatible stores may not support this.
		return nil
	}

	return nil
}

// Upload stores a file in S3 and returns the public URL.
func (s *S3Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload %q: %w", key, err)
	}

	return s.PublicURL(key), nil
}

// Delete removes a file from S3.
func (s *S3Client) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %q: %w", key, err)
	}

	return nil
}

// PublicURL returns the publicly accessible URL for a given key.
func (s *S3Client) PublicURL(key string) string {
	if s.cdnBase != "" {
		return s.cdnBase + "/" + key
	}
	return s.publicEndpoint + "/" + s.bucket + "/" + key
}

const presignExpiry = 30 * time.Minute

// GeneratePresignedUploadURL creates a presigned PUT URL for direct client-to-S3 uploads.
func (s *S3Client) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, size int64) (*PresignedUpload, error) {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	resp, err := s.presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return nil, fmt.Errorf("presign upload %q: %w", key, err)
	}

	return &PresignedUpload{
		URL:       resp.URL,
		Key:       key,
		PublicURL: s.PublicURL(key),
		ExpiresIn: presignExpiry,
	}, nil
}
