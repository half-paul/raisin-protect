// Package services provides business logic and external service integrations.
package services

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// MinIOService wraps the MinIO client for evidence file storage.
type MinIOService struct {
	client      *minio.Client
	bucket      string
	uploadTTL   time.Duration
	downloadTTL time.Duration
}

// MinIOConfig holds MinIO connection configuration.
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewMinIOService creates a new MinIO service and ensures the bucket exists.
func NewMinIOService(cfg MinIOConfig) (*MinIOService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	svc := &MinIOService{
		client:      client,
		bucket:      cfg.Bucket,
		uploadTTL:   15 * time.Minute,
		downloadTTL: 1 * time.Hour,
	}

	return svc, nil
}

// EnsureBucket creates the evidence bucket if it doesn't exist.
func (s *MinIOService) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", s.bucket).Msg("Created MinIO bucket")
	}
	return nil
}

// GenerateUploadURL creates a presigned PUT URL for uploading a file.
// Note: Content-Type enforcement happens at upload confirmation time via VerifyObjectExists,
// as MinIO's PresignedPutObject doesn't support Content-Type query param enforcement for PUT.
// The client must set the correct Content-Type header when uploading.
func (s *MinIOService) GenerateUploadURL(objectKey, contentType string) (string, error) {
	presignedURL, err := s.client.PresignedPutObject(context.Background(), s.bucket, objectKey, s.uploadTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GenerateDownloadURL creates a presigned GET URL for downloading a file.
func (s *MinIOService) GenerateDownloadURL(objectKey, fileName string) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	presignedURL, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectKey, s.downloadTTL, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return presignedURL.String(), nil
}

// VerifyObjectExists checks if a file exists in MinIO and returns its actual size.
func (s *MinIOService) VerifyObjectExists(objectKey string) (int64, error) {
	info, err := s.client.StatObject(context.Background(), s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("object not found: %w", err)
	}
	return info.Size, nil
}

// UploadTTLSeconds returns the upload URL TTL in seconds.
func (s *MinIOService) UploadTTLSeconds() int {
	return int(s.uploadTTL.Seconds())
}

// DownloadTTLSeconds returns the download URL TTL in seconds.
func (s *MinIOService) DownloadTTLSeconds() int {
	return int(s.downloadTTL.Seconds())
}
