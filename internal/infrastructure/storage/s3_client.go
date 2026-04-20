package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// S3Service implements StorageService using an S3-compatible backend (e.g. Garage).
type S3Service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicBaseURL string
}

// NewS3Service creates a new S3-backed StorageService.
func NewS3Service(cfg *config.StorageConfig) (*S3Service, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	// SeaweedFS, Garage, and older MinIO/Ceph reject the STREAMING-UNSIGNED-PAYLOAD-TRAILER
	// payload mode that aws-sdk-go-v2 v1.74+ adopted by default for PutObject.
	// Force UNSIGNED-PAYLOAD instead: the SDK signs headers only, the body is
	// transferred with normal Content-Length, and integrity is left to TLS. Every
	// S3-compatible backend supports this mode.
	useUnsignedPayload := func(stack *middleware.Stack) error {
		return v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware(stack)
	}
	sdkLogger := logging.NewStandardLogger(os.Stderr)
	client := s3.New(s3.Options{
		BaseEndpoint:               aws.String(cfg.Endpoint),
		Region:                     cfg.Region,
		Credentials:                creds,
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
		ClientLogMode:              aws.LogSigning | aws.LogRequest | aws.LogResponse | aws.LogRetries,
		Logger:                     sdkLogger,
		APIOptions:                 []func(*middleware.Stack) error{useUnsignedPayload},
	})

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	// Separate client for presigning URLs using the public endpoint.
	// PresignGetObject is a local computation (no network call), so this client
	// doesn't need to reach the public endpoint from inside Docker.
	// This ensures the S3v4 signature matches the Host header clients will send.
	presignClient := s3.NewPresignClient(client)
	if publicEndpoint != cfg.Endpoint {
		publicClient := s3.New(s3.Options{
			BaseEndpoint:               aws.String(publicEndpoint),
			Region:                     cfg.Region,
			Credentials:                creds,
			UsePathStyle:               true,
			RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
			ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
			APIOptions:                 []func(*middleware.Stack) error{useUnsignedPayload},
		})
		presignClient = s3.NewPresignClient(publicClient)
	}

	return &S3Service{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(publicEndpoint, "/"),
	}, nil
}

// EnsureBucket creates the bucket if it does not exist.
//
// A 403 Forbidden from HeadBucket is treated as "bucket exists but our
// credentials lack s3:ListBucket": production keys are often scoped to
// object-level operations only, so we can't verify existence but must
// assume the bucket is already provisioned.
func (s *S3Service) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &s.bucket,
	})
	if err == nil {
		return nil
	}

	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		if _, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &s.bucket,
		}); err != nil {
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
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("uploading object: %w", err)
	}
	return s.PublicURL(key), nil
}

func (s *S3Service) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}

func (s *S3Service) Move(ctx context.Context, srcKey, dstKey string) error {
	copySource := s.bucket + "/" + srcKey
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &s.bucket,
		Key:        &dstKey,
		CopySource: &copySource,
	})
	if err != nil {
		return fmt.Errorf("copying object from %s to %s: %w", srcKey, dstKey, err)
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &srcKey,
	})
	if err != nil {
		return fmt.Errorf("removing source object %s after copy: %w", srcKey, err)
	}

	return nil
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
