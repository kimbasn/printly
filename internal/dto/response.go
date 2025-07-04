package dto

// ErrorResponse represents a standard error response format for API calls.
// It's used to provide a consistent structure for error messages.
type ErrorResponse struct {
	Error string `json:"error" example:"A description of the error"`
}

// SuccessResponse represents a standard success message format for API calls.
// It's used for operations that return a simple confirmation message.
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

