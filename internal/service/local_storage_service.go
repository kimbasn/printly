package service

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kimbasn/printly/internal/config"
	"go.uber.org/zap"
)

type localStorageService struct {
	basePath string
	baseURL  string
	logger   *zap.Logger
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService(localConfig config.LocalStorageConfig, logger *zap.Logger) (StorageService, error) {
	// Ensure base path exists
	if err := os.MkdirAll(localConfig.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &localStorageService{
		basePath: localConfig.BasePath,
		baseURL:  localConfig.BaseURL,
		logger:   logger,
	}, nil
}

var fileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// sanitizeFileName cleans a filename for safe storage
func sanitizeFileName(filename string) string {
	// Replace unsafe characters with underscores
	sanitized := fileNameSanitizer.ReplaceAllString(filename, "_")
	// Remove any double underscores
	sanitized = strings.ReplaceAll(sanitized, "__", "_")
	// Trim leading/trailing underscores and dots
	sanitized = strings.Trim(sanitized, "_.")

	// If filename becomes empty after sanitization, use a default
	if sanitized == "" {
		sanitized = "file"
	}

	return sanitized
}

// generateUniqueFileName creates a unique filename to prevent conflicts
func (s *localStorageService) generateUniqueFileName(originalFilename, userUID string) (string, error) {
	// Sanitize the original filename
	sanitized := sanitizeFileName(originalFilename)

	// Extract extension
	ext := filepath.Ext(sanitized)
	baseName := strings.TrimSuffix(sanitized, ext)

	// Generate a random suffix
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %w", err)
	}

	// Create unique filename: baseName_userUID_timestamp_randomSuffix.ext
	timestamp := time.Now().Unix()
	uniqueName := fmt.Sprintf("%s_%s_%d_%x%s", baseName, userUID, timestamp, suffix, ext)

	return uniqueName, nil
}

// getUserStoragePath creates the storage path for a user
func (s *localStorageService) getUserStoragePath(userUID string) string {
	return filepath.Join(s.basePath, userUID)
}

// UploadFile uploads a file from multipart.File to local storage
func (s *localStorageService) UploadFile(file multipart.File, filename, userUID string) (string, error) {
	s.logger.Info("Uploading file", zap.String("filename", filename), zap.String("userUID", userUID))

	// Generate unique filename
	uniqueFilename, err := s.generateUniqueFileName(filename, userUID)
	if err != nil {
		return "", fmt.Errorf("failed to generate unique filename: %w", err)
	}

	// Create user directory if it doesn't exist
	userDir := s.getUserStoragePath(userUID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Create full file path
	filePath := filepath.Join(userDir, uniqueFilename)

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy file content
	if _, err := io.Copy(destFile, file); err != nil {
		// Clean up on error
		os.Remove(filePath)
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Return relative storage path
	storagePath := filepath.Join(userUID, uniqueFilename)

	s.logger.Info("File uploaded successfully",
		zap.String("originalFilename", filename),
		zap.String("storagePath", storagePath),
		zap.String("userUID", userUID))

	return storagePath, nil
}

// UploadFromReader uploads a file from an io.Reader to local storage
func (s *localStorageService) UploadFromReader(reader io.Reader, filename, userUID string) (string, error) {
	s.logger.Info("Uploading file from reader", zap.String("filename", filename), zap.String("userUID", userUID))

	// Generate unique filename
	uniqueFilename, err := s.generateUniqueFileName(filename, userUID)
	if err != nil {
		return "", fmt.Errorf("failed to generate unique filename: %w", err)
	}

	// Create user directory if it doesn't exist
	userDir := s.getUserStoragePath(userUID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Create full file path
	filePath := filepath.Join(userDir, uniqueFilename)

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content from reader
	if _, err := io.Copy(destFile, reader); err != nil {
		// Clean up on error
		os.Remove(filePath)
		return "", fmt.Errorf("failed to copy content from reader: %w", err)
	}

	// Return relative storage path
	storagePath := filepath.Join(userUID, uniqueFilename)

	s.logger.Info("File uploaded successfully from reader",
		zap.String("originalFilename", filename),
		zap.String("storagePath", storagePath),
		zap.String("userUID", userUID))

	return storagePath, nil
}

// DeleteFile removes a file from local storage
func (s *localStorageService) DeleteFile(storagePath string) error {
	s.logger.Info("Deleting file", zap.String("storagePath", storagePath))

	if storagePath == "" {
		return fmt.Errorf("storage path cannot be empty")
	}

	// Construct full file path
	fullPath := filepath.Join(s.basePath, storagePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", storagePath)
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Info("File deleted successfully", zap.String("storagePath", storagePath))
	return nil
}

// GetFileURL returns the public URL for accessing a file
func (s *localStorageService) GetFileURL(storagePath string) (string, error) {
	if storagePath == "" {
		return "", fmt.Errorf("storage path cannot be empty")
	}

	// Check if file exists
	fullPath := filepath.Join(s.basePath, storagePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", storagePath)
	}

	// Construct public URL
	// Replace backslashes with forward slashes for URL compatibility
	urlPath := strings.ReplaceAll(storagePath, "\\", "/")
	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(s.baseURL, "/"), urlPath)

	return fileURL, nil
}

// GetSignedURL returns a signed URL for temporary access (for local storage, this is just a regular URL with query params)
func (s *localStorageService) GetSignedURL(storagePath string, expiration time.Duration) (string, error) {
	s.logger.Info("Generating signed URL", zap.String("storagePath", storagePath), zap.Duration("expiration", expiration))

	// Get the base URL
	baseURL, err := s.GetFileURL(storagePath)
	if err != nil {
		return "", err
	}

	// For local storage, we'll simulate signed URLs with query parameters
	// In a real implementation, you might use JWT tokens or similar
	expires := time.Now().Add(expiration).Unix()

	// Generate a simple signature (in production, use proper cryptographic signing)
	signature := s.generateSimpleSignature(storagePath, expires)

	signedURL := fmt.Sprintf("%s?expires=%d&signature=%s", baseURL, expires, signature)

	s.logger.Info("Signed URL generated", zap.String("signedURL", signedURL))
	return signedURL, nil
}

// generateSimpleSignature creates a simple signature for the signed URL
// Note: This is a simplified implementation. In production, use proper HMAC or similar
func (s *localStorageService) generateSimpleSignature(storagePath string, expires int64) string {
	// Simple hash-based signature (not cryptographically secure, for demo purposes)
	data := fmt.Sprintf("%s:%d", storagePath, expires)
	hash := make([]byte, 8)
	for i, b := range []byte(data) {
		hash[i%len(hash)] ^= b
	}
	return fmt.Sprintf("%x", hash)
}

// GetFileInfo returns information about a stored file
func (s *localStorageService) GetFileInfo(storagePath string) (os.FileInfo, error) {
	fullPath := filepath.Join(s.basePath, storagePath)
	return os.Stat(fullPath)
}

// ListUserFiles returns a list of files for a specific user
func (s *localStorageService) ListUserFiles(userUID string) ([]string, error) {
	userDir := s.getUserStoragePath(userUID)

	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []string{}, nil // Return empty list if user directory doesn't exist
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %w", err)
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, filepath.Join(userUID, file.Name()))
		}
	}

	return fileList, nil
}
