package errors

// ServiceError represents a structured business logic error that handlers can use to
// generate appropriate HTTP responses with specific status codes and field-level errors.
type ServiceError struct {
	// Code is the HTTP status code for this error
	Code int

	// Message is the user-facing error message
	Message string

	// FieldErrors contains field-level validation or business logic errors (e.g., "email": "already taken")
	FieldErrors map[string]string

	// Err is the underlying error for logging purposes (can be nil)
	Err error
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error for error wrapping/chain inspection
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new ServiceError with the given HTTP code and user message
func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{
		Code:        code,
		Message:     message,
		FieldErrors: make(map[string]string),
		Err:         nil,
	}
}

// NewServiceErrorWithFields creates a new ServiceError with field-level errors
func NewServiceErrorWithFields(code int, message string, fieldErrors map[string]string) *ServiceError {
	if fieldErrors == nil {
		fieldErrors = make(map[string]string)
	}
	return &ServiceError{
		Code:        code,
		Message:     message,
		FieldErrors: fieldErrors,
		Err:         nil,
	}
}

// NewServiceErrorWithCause creates a new ServiceError with an underlying error for logging
func NewServiceErrorWithCause(code int, message string, err error) *ServiceError {
	return &ServiceError{
		Code:        code,
		Message:     message,
		FieldErrors: make(map[string]string),
		Err:         err,
	}
}

// AddFieldError adds or updates a field-level error
func (e *ServiceError) AddFieldError(field, message string) *ServiceError {
	e.FieldErrors[field] = message
	return e
}

// HasFieldErrors returns true if there are any field-level errors
func (e *ServiceError) HasFieldErrors() bool {
	return len(e.FieldErrors) > 0
}

// IsServiceError checks if an error is a ServiceError
func IsServiceError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

// AsServiceError attempts to type-assert an error to *ServiceError
// Returns the ServiceError and true if successful, nil and false otherwise
func AsServiceError(err error) (*ServiceError, bool) {
	se, ok := err.(*ServiceError)
	return se, ok
}
