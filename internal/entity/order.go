package entity

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	StatusCreated          OrderStatus = "CREATED"
	StatusAwaitingDocument OrderStatus = "AWAITING_DOCUMENT"
	StatusPendingPayment   OrderStatus = "PENDING_PAYMENT"
	StatusPaid             OrderStatus = "PAID"
	StatusAwaitingUser     OrderStatus = "AWAITING_USER"
	StatusReadyToPrint     OrderStatus = "READY_TO_PRINT"
	StatusPrinting         OrderStatus = "PRINTING"
	StatusPrinted          OrderStatus = "PRINTED"
	StatusReadyForPickup   OrderStatus = "READY_FOR_PICKUP"
	StatusCompleted        OrderStatus = "COMPLETED"
	StatusCancelled        OrderStatus = "CANCELLED"
	StatusFailed           OrderStatus = "FAILED"
)

type PrintMode string

const (
	PrePrint         PrintMode = "PRE_PRINT"
	PrintUponArrival PrintMode = "PRINT_UPON_ARRIVAL"
)

type PaperSize string

const (
	A4 PaperSize = "A4"
	A3 PaperSize = "A3"
)

type ColorMode string

const (
	Color			ColorMode = "COLOR"
	BlackAndWhite	ColorMode = "BLACK_AND_WHITE"
)

type Order struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Code         	string       	`gorm:"uniqueIndex;type:varchar(32)" json:"code"`
	UserUID      	string      	`gorm:"index" json:"user_uid"`
	PrintCenterID   uint      		`gorm:"index" json:"print_center_id"`
	PrintMode    	PrintMode   	`gorm:"type:varchar(32)" json:"print_mode"`
	Status       	OrderStatus 	`gorm:"index;type:varchar(32)" json:"status"`

	
	PickupTime   	*time.Time  	`json:"pickup_time,omitempty"`
	PaidAt         	*time.Time      `json:"paid_at,omitempty"`
	CancelledAt  	*time.Time  	`json:"cancelled_at,omitempty"`
	
	Documents		[]Document		`gorm:"foreignKey:OrderID"`

	// Relationships
	User   		User        `gorm:"foreignKey:UserUID;references:UID" json:"-"`
	Center 		PrintCenter `gorm:"foreignKey:PrintCenterID" json:"-"`
}
