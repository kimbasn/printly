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

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin, RoleManager:
		return true
	default:
		return false
	}
}

type User struct {
	UID       string    `gorm:"unique" json:"uid"` // Firebase UID (unique)
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      Role      `json:"role"` // "user", "manager", "admin"
	Email     string    `json:"email"`
	Disabled  bool      `json:"disabled" gorm:"default:false"`
	CenterID  *uint     `json:"center_id,omitempty"` // Nullable: only for managers
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
