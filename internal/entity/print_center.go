package entity

import (
	"time"

	"gorm.io/gorm"
)

type GeoPoint struct {
	Lat float64 `json:"lat" validate:"min=-90,max=90"`
	Lng float64 `json:"lng" validate:"min=-180,max=180"`
}

type Address struct {
	Number string `json:"number" validate:"required"`          
	Type   string `json:"type" validate:"required"`
	Street string `json:"street" validate:"required,min=2,max=100"`
	City   string `json:"city" validate:"required,min=2,max=50"`
}

type Weekday string
const (
	Monday    Weekday = "Monday"
	Tuesday   Weekday = "Tuesday"
	Wednesday Weekday = "Wednesday"
	Thursday  Weekday = "Thursday"
	Friday    Weekday = "Friday"
	Saturday  Weekday = "Saturday"
	Sunday    Weekday = "Sunday"
)

type WorkingHour struct {
	ID            uint    `gorm:"primaryKey" json:"-"`
	Day           Weekday `json:"day" validate:"required"`    
	Start         string  `json:"start" validate:"required"`     // Format: "08:00"
	End           string  `json:"end" validate:"required"`       // Format: "18:00"
	PrintCenterID uint    `json:"-"`
}

type Service struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PrintCenterID uint    `json:"-"`
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	PaperSize     string  `json:"paper_size" validate:"required"`
	Price         int64   `json:"price" validate:"min=0"`      
	Description   string  `json:"description" validate:"max=500"`
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
	CreatedAt time.Time      `json:"created_at"`                   // Expose creation time
	UpdatedAt time.Time      `json:"updated_at"`                   // Expose update time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email" gorm:"uniqueIndex"`
	PhoneNumber string `json:"phone_number" validate:"required"`

	GeoCoordinates GeoPoint `json:"geo_coordinates" gorm:"embedded"`
	Address        Address  `json:"address" gorm:"embedded"`

	WorkingHours []WorkingHour `json:"working_hours" gorm:"foreignKey:PrintCenterID;constraint:OnDelete:CASCADE"`
	Services     []Service     `json:"services" gorm:"foreignKey:PrintCenterID;constraint:OnDelete:CASCADE"`

	Status   PrintCenterStatus `json:"status" gorm:"type:varchar(32);default:'pending';index"`
	OwnerUID string            `json:"owner_uid" gorm:"index"`
}
