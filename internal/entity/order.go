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

type PrintOptions struct {
	Copies    int       `json:"copies"`
	Pages     string    `json:"pages"`      // e.g., "1-3,5"
	Color     string    `json:"color"`      // "color" | "black_white"
	PaperSize PaperSize `json:"paper_size" gorm:"type:varchar(8)"` // "A4", "A3", etc.
}

type Order struct {
	gorm.Model

	Code         	string       	`gorm:"column:pickup_code;uniqueIndex;type:varchar(32)" json:"code"`
	UserUID      	string      	`gorm:"index" json:"user_uid"`
	PrintCenterID   string  		`gorm:"index" json:"center_id"`
	Status       	OrderStatus 	`gorm:"index;type:varchar(32)" json:"status"`
	PrintMode    	PrintMode   	`gorm:"type:varchar(32)" json:"print_mode"`
	PickupTime   	*time.Time  	`json:"pickup_time,omitempty"`
	Paid         	bool        	`json:"paid"`
	Printed      	bool        	`json:"printed"`
	PrintedAt    	*time.Time  	`json:"printed_at,omitempty"`
	CancelledAt  	*time.Time  	`json:"cancelled_at,omitempty"`
	PrintOptions 	PrintOptions	`gorm:"embedded;embeddedPrefix:print_" json:"print_options"`

	// Optional relationships
	User   		User        `gorm:"foreignKey:UserUID;references:UID" json:"-"`
	Center 		PrintCenter `gorm:"foreignKey:PrintCenterID" json:"-"`
}

