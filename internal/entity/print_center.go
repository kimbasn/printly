package entity

import (
	"gorm.io/gorm"
)

type GeoPoint struct {
	ID  uint    `json:"-"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Location struct {
    gorm.Model 

	// Postal address
	Number                  uint8       `json:"number"`
	Type                    string      `json:"type"`                   // e.g., "Rue", "Avenue"
	Street                  string      `json:"street"`
	City                    string      `json:"city"`

    // Geographical Coordinates
    GeoPointID              uint        `json:"-"`
	GeoPoint                GeoPoint    `json:"geo_point"` 
}

type WorkingHour struct {
	ID                      uint        `gorm:"primaryKey" json:"-"`
	Day                     string      `json:"day"`                    // e.g., "Monday"
	Start                   string      `json:"start"`                  // Format: "08:00"
	End                     string      `json:"end"`                    // Format: "08:00"
    PrintCenterID           uint        `json:"-"`
}

type Service struct {
	gorm.Model

	PrintCenterID           string      `json:"-"`
	Name                    string      `json:"name"`                   // e.g., "printing"
	PaperSize               string      `json:"paper_size"`             // e.g., "A4"
	Price                   float64     `json:"price"`
	Description             string      `json:"description"`
}

type PrintCenter struct {
	gorm.Model

	Name                    string      `json:"name"`
	Email                   string      `json:"email"`
	PhoneNumber             string      `json:"phone_number"`

    Location                Location    `json:"location" gorm:"embedded"`
    LocationID              int         `json:"location_id"`            // Internal FK

	WorkingHours            []WorkingHour `json:"working_hours" gorm:"foreignKey:PrintCenterID"`
	Services                []Service     `json:"services" gorm:"foreignKey:PrintCenterID"`

	Approved     bool          `json:"approved"`
	OwnerUID     string        `json:"owner_uid"`                       // Firebase UID of the manager
}
