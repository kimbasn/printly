package dto

import "github.com/kimbasn/printly/internal/entity"

// CreateUserRequest defines the structure for creating a new user.
// It includes validation tags to ensure data integrity and only exposes fields
// that should be provided by the client during registration.
type CreateUserRequest struct {
	UID         string      `json:"uid" validate:"required"`
	Email       string      `json:"email" validate:"omitempty,email"`
	PhoneNumber string      `json:"phone_number" validate:"required"`
	Role        entity.Role `json:"role" validate:"required"`
}

// UpdateUserRequest defines the structure for updating a user's profile.
// Only a subset of fields are exposed for modification to enhance security.
type UpdateUserRequest struct {
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}
