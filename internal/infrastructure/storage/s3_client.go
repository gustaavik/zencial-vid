package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// S3Service implements StorageService against any S3-compatible backend
// (AWS S3, MinIO, SeaweedFS, Garage).
type S3Service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicBaseURL string
}

// NewS3Service creates a new S3-backed StorageService.
func NewS3Service(cfg *config.StorageConfig) (*S3Service, error) {
	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	client := newS3Client(cfg, cfg.Endpoint)
	presignSource := client
	if publicEndpoint != cfg.Endpoint {
		presignSource = newS3Client(cfg, publicEndpoint)
	}

	return &S3Service{
		client:        client,
		presignClient: s3.NewPresignClient(presignSource),
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(publicEndpoint, "/"),
	}, nil
}

// newS3Client builds an s3.Client compatible with AWS S3, MinIO, and SeaweedFS.
//
// SwapComputePayloadSHA256ForUnsignedPayloadMiddleware replaces the default
// STREAMING-UNSIGNED-PAYLOAD-TRAILER mode (aws-sdk-go-v2 v1.74+) which
// MinIO/SeaweedFS/Garage reject; checksum calculation is likewise gated to
// "when required" for the same compatibility reason.
func newS3Client(cfg *config.StorageConfig, endpoint string) *s3.Client {
	return s3.New(s3.Options{
		BaseEndpoint:               aws.String(endpoint),
		Region:                     cfg.Region,
		Credentials:                credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
		APIOptions: []func(*middleware.Stack) error{
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		},
	})
}

// EnsureBucket creates the bucket if it does not exist. A 403 on HeadBucket is
// treated as "exists but caller lacks s3:ListBucket" (common on AWS with
// object-scoped credentials).
func (s *S3Service) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket})
	if err == nil {
		return nil
	}

	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		if _, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &s.bucket}); err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
		return nil
	}

	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusForbidden {
		return nil
	}

	return fmt.Errorf("checking bucket existence: %w", err)
}

func (s *S3Service) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	}); err != nil {
		return "", fmt.Errorf("uploading object: %w", err)
	}
	return s.PublicURL(key), nil
}

func (s *S3Service) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}

func (s *S3Service) Move(ctx context.Context, srcKey, dstKey string) error {
	copySource := s.bucket + "/" + srcKey
	if _, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &s.bucket,
		Key:        &dstKey,
		CopySource: &copySource,
	}); err != nil {
		return fmt.Errorf("copying object %s to %s: %w", srcKey, dstKey, err)
	}
	return s.Delete(ctx, srcKey)
}

func (s *S3Service) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, s.bucket, key)
}

func (s *S3Service) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}
	return req.URL, nil
}
