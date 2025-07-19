package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/kimbasn/printly/internal/config"
	"go.uber.org/zap"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	UploadFile(file multipart.File, filename, userUID string) (string, error)
	UploadFromReader(reader io.Reader, filename, userUID string) (string, error)
	DeleteFile(storagePath string) error
	GetFileURL(storagePath string) (string, error)
	GetSignedURL(storagePath string, expiration time.Duration) (string, error)
}

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeGCS   StorageType = "gcs"
)

// LocalStorageConfig holds configuration for local storage
type LocalStorageConfig struct {
	BasePath string // Base directory for file storage (e.g., "./uploads")
	BaseURL  string // Base URL for serving files (e.g., "http://localhost:8080/files")
}

// GCSStorageConfig holds configuration for Google Cloud Storage
type GCSStorageConfig struct {
	BucketName            string // GCS bucket name
	ProjectID             string // GCP project ID
	CredentialsPath       string // Path to service account JSON file (optional)
	CredentialsJSON       string // Service account JSON content (optional)
	UseApplicationDefault bool   // Use application default credentials
}

// StorageConfig holds configuration for storage services
type StorageConfig struct {
	Type  StorageType
	Local LocalStorageConfig
	GCS   GCSStorageConfig
}

// GetStorageService creates and returns a StorageService instance based on the provided config
func GetStorageService(cfg *config.Config, logger *zap.Logger) (StorageService, error) {
	storageConfig := cfg.GetStorageConfig()

	switch storageConfig.Type {
	case config.StorageTypeLocal:
		return NewLocalStorageService(storageConfig.Local, logger)
	case config.StorageTypeGCS:
		return NewGCSStorageService(storageConfig.GCS, logger)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageConfig.Type)
	}
}
