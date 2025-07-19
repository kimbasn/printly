package dto

import (
	"mime/multipart"

	"github.com/kimbasn/printly/internal/entity"
)

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
	Name            string               `json:"name" validate:"required,min=3"`
	Email           string               `json:"email" validate:"required,email"`
	PhoneNumber     string               `json:"phone_number" validate:"required"`
	Address         entity.Address       `json:"address" validate:"required"`
	Geo_Coordinates entity.GeoPoint      `json:"geo_coordinates" validate:"required"`
	Services        []entity.Service     `json:"services" validate:"dive"`
	WorkingHours    []entity.WorkingHour `json:"working_hours" validate:"required,min=1,dive"`
}

// UpdatePrintCenterRequest defines the structure for partially updating a print center.
// Pointers are used to distinguish between a field not being provided and a field being provided with its zero value.
type UpdatePrintCenterRequest struct {
	Name            *string               `json:"name,omitempty" validate:"omitempty,min=3"`
	PhoneNumber     *string               `json:"phone_number,omitempty"`
	Address         entity.Address        `json:"address" validate:"omitempty"`
	Geo_Coordinates entity.GeoPoint       `json:"geo_coordinates" validate:"omitempty"`
	Services        *[]entity.Service     `json:"services,omitempty" validate:"omitempty,min=1,dive"`
	WorkingHours    *[]entity.WorkingHour `json:"working_hours,omitempty" validate:"omitempty,min=1,dive"`
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
	Documents []CreateDocumentRequest `json:"documents" validate:"required,min=1,dive"`
}

// UpdateOrderStatusRequest defines the structure for updating an order's status.
type UpdateOrderStatusRequest struct {
	Status entity.OrderStatus `json:"status" validate:"required"`
}

// DocumentPrintRequest represents the print configuration for a single document
type DocumentPrintRequest struct {
	PrintMode    string              `json:"print_mode" validate:"required"`
	PrintOptions entity.PrintOptions `json:"print_options" validate:"required"`
}

// CreateDocumentRequest represents a document in the order creation request
type CreateDocumentRequest struct {
	FileName     string              `json:"file_name" validate:"required,max=255"`
	MimeType     string              `json:"mime_type" validate:"required"`
	Size         int64               `json:"size" validate:"required,min=1,max=52428800"` // 50MB
	StoragePath  string              `json:"storage_path,omitempty"`                      // Internal storage path
	URL          string              `json:"url,omitempty"`                               // For JSON uploads
	PrintMode    entity.PrintMode    `json:"print_mode" validate:"required,oneof=PRE_PRINT,PRINT_UPON_ARRIVAL"`
	PrintOptions entity.PrintOptions `json:"print_options" validate:"required"`
}

// MultipartOrderRequest represents the multipart form data structure
type MultipartOrderRequest struct {
	PrintMode    string                 `form:"print_mode" validate:"required"`
	Files        []multipart.FileHeader `form:"files" validate:"required,min=1"`
	PrintOptions string                 `form:"print_options" validate:"required"` // JSON string
}

// FileUploadResponse represents the response after uploading files
type FileUploadResponse struct {
	Files []UploadedFile `json:"files"`
}

// UploadedFile represents an uploaded file's metadata
type UploadedFile struct {
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	MimeType    string `json:"mime_type"`
	StoragePath string `json:"storage_path"`
	URL         string `json:"url"`
}
