package entity

import (
	"time"

)

// Role defines the access level of a user
type Role string

const (
	RoleUser    Role = "user"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
)

type User struct {
	UID         string    `gorm:"primaryKey" json:"uid"`         // Firebase UID (unique)
	Role        Role      `json:"role"`                          // "user", "manager", "admin"
	Email       string    `json:"email,omitempty"`               // Optional for anonymous
	PhoneNumber string    `json:"phone_number"`                  // Required for Mobile Money and contact
	CenterID    *uint     `json:"center_id,omitempty"`           // Nullable: only for managers
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
