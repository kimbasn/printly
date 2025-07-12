package entity

import (
	"time"
)

type PrintOptions struct {
	Copies    int       `json:"copies"`
	Pages     string    `json:"pages"`                             // e.g., "1-3,5"
	Color     ColorMode `json:"color" gorm:"type:varchar(16)"`     // "color" | "black_white"
	PaperSize PaperSize `json:"paper_size" gorm:"type:varchar(8)"` // "A4", "A3", etc.
}

type Document struct {
	ID uint `gorm:"primaryKey" json:"id"`

	OrderID    uint       `gorm:"index;not null" json:"order_id"`
	FileName   string     `gorm:"type:varchar(255)" json:"file_name"`
	MimeType   string     `gorm:"type:varchar(128)" json:"mime_type"`
	URL        string     `gorm:"type:text" json:"url"`
	Size       int64      `json:"size"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`

	PrintOptions PrintOptions `gorm:"embedded;embeddedPrefix:print_" json:"print_options"`

	PrintedAt        *time.Time `json:"printed_at,omitempty"`         // When it was printed (nullable)
	StorageDeletedAt *time.Time `json:"storage_deleted_at,omitempty"` // When it was deleted from GCS

	Order Order `gorm:"foreignKey:OrderID;references:ID" json:"-"`
}
