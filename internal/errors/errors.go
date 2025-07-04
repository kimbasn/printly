package errors

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with given UID already exists")
	ErrInvalidUserData   = errors.New("invalid user data")
	ErrEmailAlreadyTaken = errors.New("email already in use")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInternalServer    = errors.New("internal server error")
)