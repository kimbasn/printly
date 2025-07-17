package dto

import "github.com/kimbasn/printly/internal/entity"

// CreateUserRequest defines the structure for creating a new user.
// It includes validation tags to ensure data integrity and only exposes fields
// that should be provided by the client during registration.
type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	// PhoneNumber string      `json:"phone_number" validate:"required"`
	// Role entity.Role `json:"role" validate:"required"`
}

// UpdateUserRequest defines the structure for updating a user's profile.
// Only a subset of fields are exposed for modification to enhance security.
type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty"`
	LastName  string `json:"last_name" validate:"omitempty,"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	Disabled  bool   `json:"disabled,omitempty"`
	// PhoneNumber string `json:"phone_number,omitempty"`
}

// UpdateUserRoleRequest defines the structure for changing a user's role.
type UpdateUserRoleRequest struct {
	// The new role for the user. Must be 'user', 'manager', or 'admin'.
	Role entity.Role `json:"role" validate:"required,is-valid-role"`
}

// CreatePrintCenterRequest defines the structure for creating a new print center.
type CreatePrintCenterRequest struct {
	Name         string               `json:"name" validate:"required,min=3"`
	Email        string               `json:"email" validate:"required,email"`
	PhoneNumber  string               `json:"phone_number" validate:"required"`
	Location     entity.Location      `json:"location" validate:"required"`
	Services     []entity.Service     `json:"services" validate:"dive"`
	WorkingHours []entity.WorkingHour `json:"working_hours" validate:"required,min=1,dive"`
}

// UpdatePrintCenterRequest defines the structure for partially updating a print center.
// Pointers are used to distinguish between a field not being provided and a field being provided with its zero value.
type UpdatePrintCenterRequest struct {
	Name         *string               `json:"name,omitempty" validate:"omitempty,min=3"`
	PhoneNumber  *string               `json:"phone_number,omitempty"`
	Location     *entity.Location      `json:"location,omitempty" validate:"omitempty"`
	Services     *[]entity.Service     `json:"services,omitempty" validate:"omitempty,min=1,dive"`
	WorkingHours *[]entity.WorkingHour `json:"working_hours,omitempty" validate:"omitempty,min=1,dive"`
}

// UpdatePrintCenterStatusRequest defines the structure for updating a print center's status.
type UpdatePrintCenterStatusRequest struct {
	Status entity.PrintCenterStatus `json:"status" validate:"required,oneof=pending approved rejected suspended"`
}

// DocumentRequest defines the metadata for a single document within an order creation request.
type DocumentRequest struct {
	FileName     string              `json:"file_name" validate:"required"`
	MimeType     string              `json:"mime_type" validate:"required"`
	Size         int64               `json:"size" validate:"required,gt=0"`
	PrintOptions entity.PrintOptions `json:"print_options" validate:"required,dive"`
}

// CreateOrderRequest defines the structure for creating a new order with multiple documents.
type CreateOrderRequest struct {
	PrintMode entity.PrintMode  `json:"print_mode" validate:"required,oneof=PRE_PRINT,PRINT_UPON_ARRIVAL"`
	Documents []DocumentRequest `json:"documents" validate:"required,min=1,dive"`
}

// UpdateOrderStatusRequest defines the structure for updating an order's status.
type UpdateOrderStatusRequest struct {
	Status entity.OrderStatus `json:"status" validate:"required"`
}
