package errors

// ErrorCode represents the platform-wide error codes
type ErrorCode string

const (
	// InvalidArgument is a OnePlatform error code.
	InvalidArgument ErrorCode = "INVALID_ARGUMENT"

	// FailedPrecondition is a OnePlatform error code.
	FailedPrecondition ErrorCode = "FAILED_PRECONDITION"

	// OutOfRange is a OnePlatform error code.
	OutOfRange ErrorCode = "OUT_OF_RANGE"

	// Unauthenticated is a OnePlatform error code.
	Unauthenticated ErrorCode = "UNAUTHENTICATED"

	// PermissionDenied is a OnePlatform error code.
	PermissionDenied ErrorCode = "PERMISSION_DENIED"

	// NotFound is a OnePlatform error code.
	NotFound ErrorCode = "NOT_FOUND"

	Disabled ErrorCode = "DISABLED"

	Conflict ErrorCode = "CONFLICT"

	// Aborted is a OnePlatform error code.
	Aborted ErrorCode = "ABORTED"

	// AlreadyExists is a OnePlatform error code.
	AlreadyExists ErrorCode = "ALREADY_EXISTS"

	// ResourceExhausted is a OnePlatform error code.
	ResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"

	// Cancelled is a OnePlatform error code.
	Cancelled ErrorCode = "CANCELLED"

	// DataLoss is a OnePlatform error code.
	DataLoss ErrorCode = "DATA_LOSS"

	// Unknown is a OnePlatform error code.
	Unknown ErrorCode = "UNKNOWN"

	// Internal is a OnePlatform error code.
	Internal ErrorCode = "INTERNAL"

	// Unavailable is a OnePlatform error code.
	Unavailable ErrorCode = "UNAVAILABLE"

	// DeadlineExceeded is a OnePlatform error code.
	DeadlineExceeded ErrorCode = "DEADLINE_EXCEEDED"
)

type AppError struct {
	Code   ErrorCode
	String string
	Cause  error
}

func (e *AppError) Error() string {
	return e.String
}

func New(code ErrorCode, msg string) *AppError {
	return &AppError{
		Code:   code,
		String: msg,
	}
}

func NewWithCause(code ErrorCode, msg string, cause error) *AppError {
	return &AppError{
		Code:   code,
		String: msg,
		Cause:  cause,
	}

}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	return ok && e.Code == t.Code
}

func (e *AppError) UnWrap() error {
	return e.Cause
}

var (
	ErrUserNotFound       = New(NotFound, "user not found")
	ErrUserDisabled       = New(Disabled, "user disabled")
	ErrUserAlreadyExists  = New(AlreadyExists, "user with given UID already exists")
	ErrInvalidUserData    = New(InvalidArgument, "invalid user data")
	ErrEmailAlreadyExists = New(AlreadyExists, "email already in use")

	ErrUnauthorized   = New(Unauthenticated, "unauthorized access")
	ErrInternalServer = New(Internal, "internal server error")
	ErrDatabase       = New(Internal, "database error")

	ErrPrintCenterNotFound        = New(NotFound, "center not found")
	ErrPrintCenterAlreadyApproved = New(FailedPrecondition, "center already approved")

	ErrOrderNotFound = New(NotFound, "order not found")
)
