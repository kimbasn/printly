package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"cloud.google.com/go/storage"
)

// Example implementation for Google Cloud Storage
type gcsStorageService struct {
	bucketName string
	client     *storage.Client
}

func NewGCSStorageService(bucketName string, client *storage.Client) StorageService {
	return &gcsStorageService{
		bucketName: bucketName,
		client:     client,
	}
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
