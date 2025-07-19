package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kimbasn/printly/internal/config"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// Example implementation for Google Cloud Storage
type gcsStorageService struct {
	bucketName string
	client     *storage.Client
	logger     *zap.Logger
}

// NewGCSStorageService creates a new Google Cloud Storage service
func NewGCSStorageService(config config.GCSStorageConfig, logger *zap.Logger) (StorageService, error) {
	ctx := context.Background()
	var client *storage.Client
	var err error

	// Create client based on authentication method
	switch {
	case config.UseApplicationDefault:
		client, err = storage.NewClient(ctx)
	case config.CredentialsJSON != "":
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(config.CredentialsJSON)))
	case config.CredentialsPath != "":
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(config.CredentialsPath))
	default:
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &gcsStorageService{
		bucketName: config.BucketName,
		client:     client,
		logger:     logger,
	}, nil
}

func (s *gcsStorageService) UploadFile(file multipart.File, filename, userUID string) (string, error) {
	// Generate unique storage path
	storagePath := fmt.Sprintf("orders/%s/%s/%s", userUID, time.Now().Format("2006-01-02"), filename)

	// Upload to GCS
	ctx := context.Background()
	wc := s.client.Bucket(s.bucketName).Object(storagePath).NewWriter(ctx)

	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	return storagePath, nil
}

func (s *gcsStorageService) DeleteFile(storagePath string) error {
	ctx := context.Background()
	return s.client.Bucket(s.bucketName).Object(storagePath).Delete(ctx)
}

func (s *gcsStorageService) GetFileURL(storagePath string) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, storagePath), nil
}

func (s *gcsStorageService) GetSignedURL(storagePath string, expiration time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	return s.client.Bucket(s.bucketName).SignedURL(storagePath, opts)
}

func (s *gcsStorageService) UploadFromReader(reader io.Reader, filename, userUID string) (string, error) {
	return "", nil
}
