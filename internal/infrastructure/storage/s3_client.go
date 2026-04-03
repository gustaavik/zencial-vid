package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.Endpoint),
		Region:       cfg.Region,
		Credentials:  creds,
		UsePathStyle: true,
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
			BaseEndpoint: aws.String(publicEndpoint),
			Region:       cfg.Region,
			Credentials:  creds,
			UsePathStyle: true,
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
func (s *S3Service) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &s.bucket,
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: &s.bucket,
			})
			if err != nil {
				return fmt.Errorf("creating bucket: %w", err)
			}
		} else {
			return fmt.Errorf("checking bucket existence: %w", err)
		}
	}

	// Allow anonymous GET on thumbnail objects so PublicURL works without signing.
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"PublicReadThumbnails","Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::%s/videos/*/thumbnail*"}]}`, s.bucket)
	_, err = s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: &s.bucket,
		Policy: &policy,
	})
	if err != nil {
		return fmt.Errorf("setting bucket policy: %w", err)
	}

	return nil
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
