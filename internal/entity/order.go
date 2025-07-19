package entity

import (
	"fmt"
	"slices"
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
	A5 PaperSize = "A5"
	A6 PaperSize = "A6"
)

type Order struct {
	// gorm.Model is replaced to be explicit for swagger
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Code          string      `gorm:"uniqueIndex;type:varchar(32)" json:"code" validate:"required,len=8"`
	UserUID       string      `gorm:"index" json:"user_uid" validate:"required"`
	PrintCenterID uint        `gorm:"index" json:"print_center_id" validate:"required"`
	Status        OrderStatus `gorm:"index;type:varchar(32)" json:"status" validate:"required"`

	// Pricing
	TotalCost int64  `json:"total_cost" validate:"min=0"`                   // in cents
	Currency  string `json:"currency" gorm:"type:varchar(3);default:'EUR'"` // ISO currency code

	// Timestamps
	PickupTime  *time.Time `json:"pickup_time,omitempty"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Audit fields
	CreatedBy string `gorm:"index" json:"created_by"`
	UpdatedBy string `gorm:"index" json:"updated_by"`

	// Relationships
	Documents []Document  `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"documents"`
	User      User        `gorm:"foreignKey:UserUID;references:UID" json:"-"`
	Center    PrintCenter `gorm:"foreignKey:PrintCenterID" json:"-"`
}

// Helper methods for Order
func (o *Order) IsActive() bool {
	return o.Status != StatusCompleted && o.Status != StatusCancelled && o.Status != StatusFailed
}

func (o *Order) CanCancel() bool {
	return o.Status == StatusPendingPayment || o.Status == StatusAwaitingDocument
}

func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		StatusCreated:          {StatusAwaitingDocument, StatusCancelled},
		StatusAwaitingDocument: {StatusPendingPayment, StatusCancelled},
		StatusPendingPayment:   {StatusPaid, StatusCancelled, StatusFailed},
		StatusPaid:             {StatusAwaitingUser, StatusReadyToPrint, StatusCancelled},
		StatusAwaitingUser:     {StatusReadyToPrint, StatusCancelled},
		StatusReadyToPrint:     {StatusPrinting, StatusCancelled},
		StatusPrinting:         {StatusPrinted, StatusFailed},
		StatusPrinted:          {StatusReadyForPickup},
		StatusReadyForPickup:   {StatusCompleted},
		// Terminal states
		StatusCompleted: {},
		StatusCancelled: {},
		StatusFailed:    {},
	}

	allowed, exists := validTransitions[o.Status]
	if !exists {
		return false
	}

	return slices.Contains(allowed, newStatus)
}

func (o *Order) GetTotalCostInCurrency() float64 {
	return float64(o.TotalCost) / 100.0
}

func (o *Order) GenerateCode() string {
	// This should be implemented based on your business logic
	// Example: ORD-20250718-1234
	return fmt.Sprintf("ORD-%s-%04d",
		time.Now().Format("20060102"),
		o.ID%10000)
}
