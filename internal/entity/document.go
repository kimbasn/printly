package entity

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	gorm.Model

	OrderID    string    `gorm:"index;not null" json:"order_id"`
	FileName   string    `gorm:"type:varchar(255)" json:"file_name"`
	MimeType   string    `gorm:"type:varchar(128)" json:"mime_type"`
	URL        string    `gorm:"type:text" json:"url"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`

	Printed    bool       `json:"printed"`                     // Was it printed?
	Deleted    bool       `json:"deleted"`                     // Was it removed from storage?
	PrintedAt  *time.Time `json:"printed_at,omitempty"`        // When it was printed (nullable)
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`        // When it was deleted from GCS

	Order Order `gorm:"foreignKey:OrderID;references:ID" json:"-"`
}
