package service

import (
	"io"
	"mime/multipart"
	"time"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	UploadFile(file multipart.File, filename, userUID string) (string, error)
	UploadFromReader(reader io.Reader, filename, userUID string) (string, error)
	DeleteFile(storagePath string) error
	GetFileURL(storagePath string) (string, error)
	GetSignedURL(storagePath string, expiration time.Duration) (string, error)
}


type 