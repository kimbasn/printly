package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeGCS   StorageType = "gcs"
)

// StorageConfig holds configuration for storage services
type StorageConfig struct {
	Type  StorageType
	Local LocalStorageConfig
	GCS   GCSStorageConfig
}

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

type Config struct {
	AppEnv                  string
	DBDriver                string // "sqlite", "postgres", etc.
	DBSource                string // DSN or file path
	Host                    string
	Port                    string
	FirebaseCredentialsFile string
	Storage                 StorageConfig
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("⚠️ Invalid boolean value for %s: %s, using fallback: %t", key, val, fallback)
		return fallback
	}
	return boolVal
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, loading from system ENV only")
	}

	cfg := &Config{
		AppEnv:                  getEnv("APP_ENV", "development"),
		DBDriver:                getEnv("DB_DRIVER", "sqlite"),
		DBSource:                getEnv("DB_SOURCE", "printly.db"),
		Host:                    getEnv("SERVER_ADDRESS", "localhost"),
		Port:                    getEnv("PORT", "8080"),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", "FIREBASE_CREDENTIALS_FILE_NOT_FOUND"),
		Storage:                 loadStorageConfig(),
	}

	return cfg
}

func loadStorageConfig() StorageConfig {
	storageType := getEnv("STORAGE_TYPE", "local")

	config := StorageConfig{
		Type: StorageType(storageType),
	}

	switch config.Type {
	case StorageTypeLocal:
		config.Local = LocalStorageConfig{
			BasePath: getEnv("STORAGE_LOCAL_BASE_PATH", "./uploads"),
			BaseURL:  getEnv("STORAGE_LOCAL_BASE_URL", "http://localhost:8080/files"),
		}
	case StorageTypeGCS:
		config.GCS = GCSStorageConfig{
			BucketName:            getEnv("STORAGE_GCS_BUCKET_NAME", ""),
			ProjectID:             getEnv("STORAGE_GCS_PROJECT_ID", ""),
			CredentialsPath:       getEnv("STORAGE_GCS_CREDENTIALS_PATH", ""),
			CredentialsJSON:       getEnv("STORAGE_GCS_CREDENTIALS_JSON", ""),
			UseApplicationDefault: getEnvBool("STORAGE_GCS_USE_APPLICATION_DEFAULT", false),
		}
	default:
		log.Printf("⚠️ Unsupported storage type: %s, defaulting to local", storageType)
		config.Type = StorageTypeLocal
		config.Local = LocalStorageConfig{
			BasePath: getEnv("STORAGE_LOCAL_BASE_PATH", "./uploads"),
			BaseURL:  getEnv("STORAGE_LOCAL_BASE_URL", "http://localhost:8080/files"),
		}
	}

	return config
}

// ValidateConfig validates the loaded configuration
func (c *Config) ValidateConfig() error {
	// Validate storage configuration
	switch c.Storage.Type {
	case StorageTypeLocal:
		if c.Storage.Local.BasePath == "" {
			return fmt.Errorf("local storage base path is required")
		}
		if c.Storage.Local.BaseURL == "" {
			return fmt.Errorf("local storage base URL is required")
		}
	case StorageTypeGCS:
		if c.Storage.GCS.BucketName == "" {
			return fmt.Errorf("GCS bucket name is required")
		}
		// At least one authentication method should be specified
		if !c.Storage.GCS.UseApplicationDefault &&
			c.Storage.GCS.CredentialsPath == "" &&
			c.Storage.GCS.CredentialsJSON == "" {
			return fmt.Errorf("GCS authentication method is required")
		}
	}

	// Validate other configuration fields
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.DBDriver == "" {
		return fmt.Errorf("database driver is required")
	}
	if c.DBSource == "" {
		return fmt.Errorf("database source is required")
	}

	return nil
}

// GetStorageConfig returns the storage configuration
func (c *Config) GetStorageConfig() StorageConfig {
	return c.Storage
}

// IsProduction returns true if the app is running in production
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment returns true if the app is running in development
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Host + ":" + c.Port
}

// PrintConfig prints the configuration (without sensitive data)
func (c *Config) PrintConfig() {
	log.Printf("📋 Configuration loaded:")
	log.Printf("  App Environment: %s", c.AppEnv)
	log.Printf("  Database Driver: %s", c.DBDriver)
	log.Printf("  Database Source: %s", c.DBSource)
	log.Printf("  Server Address: %s", c.GetServerAddress())
	log.Printf("  Storage Type: %s", c.Storage.Type)

	switch c.Storage.Type {
	case StorageTypeLocal:
		log.Printf("  Local Storage Path: %s", c.Storage.Local.BasePath)
		log.Printf("  Local Storage URL: %s", c.Storage.Local.BaseURL)
	case StorageTypeGCS:
		log.Printf("  GCS Bucket: %s", c.Storage.GCS.BucketName)
		log.Printf("  GCS Project ID: %s", c.Storage.GCS.ProjectID)
		log.Printf("  GCS Use App Default: %t", c.Storage.GCS.UseApplicationDefault)
		if c.Storage.GCS.CredentialsPath != "" {
			log.Printf("  GCS Credentials Path: %s", c.Storage.GCS.CredentialsPath)
		}
		if c.Storage.GCS.CredentialsJSON != "" {
			log.Printf("  GCS Credentials JSON: [PROVIDED]")
		}
	}
}
