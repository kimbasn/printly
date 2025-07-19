package entity

import (
	"fmt"
	"strings"
	"time"
)

type ColorMode string

const (
	Color         ColorMode = "COLOR"
	BlackAndWhite ColorMode = "BLACK_AND_WHITE"
)

type PrintOptions struct {
	Copies      int       `json:"copies" validate:"min=1,max=100"`
	Pages       string    `json:"pages" validate:"required"` // e.g., "1-3,5" - add custom validation
	Color       ColorMode `json:"color" gorm:"type:varchar(16)" validate:"required"`
	PaperSize   PaperSize `json:"paper_size" gorm:"type:varchar(8)" validate:"required"`
	DoubleSided bool      `json:"double_sided" gorm:"default:true"`
}

type Document struct {
	ID uint `gorm:"primaryKey" json:"id"`

	OrderID  uint   `gorm:"index;not null" json:"order_id"`
	FileName string `gorm:"type:varchar(255)" json:"file_name" validate:"required,max=255"`
	MimeType string `gorm:"type:varchar(128)" json:"mime_type" validate:"required"`
	//StoragePath string     `gorm:"type:text" json:"-"`                 // Internal storage path
	Size       int64      `json:"size" validate:"min=1,max=52428800"` // 50MB limit
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`

	PrintOptions PrintOptions `gorm:"embedded;embeddedPrefix:print_" json:"print_options"`

	PrintedAt        *time.Time `json:"printed_at,omitempty"`
	StorageDeletedAt *time.Time `json:"storage_deleted_at,omitempty"`

	Order Order `gorm:"foreignKey:OrderID;references:ID" json:"-"`
}

// Helper methods for Document

func (d *Document) GetStoragePath() string {
	fileType := strings.Split(d.MimeType, "/")[1]
	return fmt.Sprintf("documents/%d/%s.%s", d.OrderID, d.FileName, fileType)
}

func (d *Document) GetPublicURL() string {
	// Generate signed URL or public URL from storage path
	// This should be implemented based on your storage solution
	return fmt.Sprintf("https://storage.example.com%s", d.GetStoragePath())
}

func (d *Document) GetSizeInMB() float64 {
	return float64(d.Size) / (1024 * 1024)
}

func (d *Document) IsImage() bool {
	return d.MimeType == "image/jpeg" || d.MimeType == "image/png" || d.MimeType == "image/gif"
}

func (d *Document) IsPDF() bool {
	return d.MimeType == "application/pdf"
}

// Calculate total cost for print options
func (po *PrintOptions) CalculateCost(pricePerPage int64) int64 {
	// This is a simplified calculation
	// You'd need to implement proper page range parsing
	return int64(po.Copies) * pricePerPage
}
