package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// StorageService defines the interface for file storage.
type StorageService interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

// S3Client implements StorageService using S3-compatible storage (AWS S3 / MinIO).
type S3Client struct {
	client   *s3.Client
	bucket   string
	endpoint string
	cdnBase  string
}

// NewS3Client creates a new S3Client from the storage and CDN configuration.
func NewS3Client(storageCfg config.StorageConfig, cdnBaseURL string) (*S3Client, error) {
	if storageCfg.Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required for storage")
	}

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

	return &S3Client{
		client:   client,
		bucket:   storageCfg.Bucket,
		endpoint: strings.TrimRight(storageCfg.Endpoint, "/"),
		cdnBase:  strings.TrimRight(cdnBaseURL, "/"),
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
	return s.endpoint + "/" + s.bucket + "/" + key
}
