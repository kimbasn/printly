package entity

import (
	"time"

	"gorm.io/gorm"
)

type GeoPoint struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Location struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Postal address
	Number                  uint8       `json:"number"`
	Type                    string      `json:"type"`                   // e.g., "Rue", "Avenue"
	Street                  string      `json:"street"`
	City                    string      `json:"city"`

    // Geographical Coordinates
    GeoPointID              *uint        `json:"-"`
	GeoPoint                *GeoPoint    `json:"geo_point"` 
}

type WorkingHour struct {
	ID                      uint        `gorm:"primaryKey" json:"-"`
	Day                     string      `json:"day"`                    // e.g., "Monday"
	Start                   string      `json:"start"`                  // Format: "08:00"
	End                     string      `json:"end"`                    // Format: "08:00"
    PrintCenterID           uint        `json:"-"`
}

type Service struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PrintCenterID           uint      `json:"-"`
	Name                    string      `json:"name"`                   // e.g., "printing"
	PaperSize               string      `json:"paper_size"`             // e.g., "A4"
	Price                   float64     `json:"price"`
	Description             string      `json:"description"`
}

type PrintCenterStatus string

const (
	StatusPending   PrintCenterStatus = "pending"
	StatusApproved  PrintCenterStatus = "approved"
	StatusRejected  PrintCenterStatus = "rejected"
	StatusSuspended PrintCenterStatus = "suspended"
)

type PrintCenter struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name                    string      		`json:"name"`
	Email                   string      		`json:"email" gorm:"uniqueIndex"`
	PhoneNumber             string      		`json:"phone_number"`

    Location                Location    		`json:"location"`
    LocationID              uint         		`json:"location_id"`            // Internal FK

	WorkingHours            []WorkingHour 		`json:"working_hours" gorm:"foreignKey:PrintCenterID"`
	Services                []Service     		`json:"services" gorm:"foreignKey:PrintCenterID"`

	Status     				PrintCenterStatus   `json:"status" gorm:"type:varchar(32);default:'pending'"`
	OwnerUID     			string        		`json:"owner_uid" gorm:"index"`                       // Firebase UID of the manager
}
