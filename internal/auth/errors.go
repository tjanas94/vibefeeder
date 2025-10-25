package auth

import (
	"net/http"

	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
)

// NewInvalidCredentialsError creates a ServiceError for invalid login credentials
// Returns 401 Unauthorized
func NewInvalidCredentialsError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusUnauthorized,
		"Invalid email or password",
	)
}

// NewUserAlreadyExistsError creates a ServiceError for duplicate email registration
// Returns 409 Conflict with field error
func NewUserAlreadyExistsError() *sharederrors.ServiceError {
	return sharederrors.NewServiceErrorWithFields(
		http.StatusConflict,
		"",
		map[string]string{
			"Email": "User with this email already exists",
		},
	)
}

// NewInvalidTokenError creates a ServiceError for invalid/expired tokens
// Returns 400 Bad Request
func NewInvalidTokenError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusBadRequest,
		"Invalid or expired token",
	)
}

// NewSessionExpiredError creates a ServiceError for expired sessions
// Returns 401 Unauthorized
func NewSessionExpiredError() *sharederrors.ServiceError {
	return sharederrors.NewServiceError(
		http.StatusUnauthorized,
		"Session expired",
	)
}

// NewInvalidRegistrationCodeError creates a ServiceError for incorrect registration code
// Returns 422 Unprocessable Entity with field error
func NewInvalidRegistrationCodeError() *sharederrors.ServiceError {
	return sharederrors.NewServiceErrorWithFields(
		http.StatusUnprocessableEntity,
		"",
		map[string]string{
			"RegistrationCode": "Invalid registration code",
		},
	)
}

// NewSamePasswordError creates a ServiceError when attempting to set the same password
// Returns 422 Unprocessable Entity with field error
func NewSamePasswordError() *sharederrors.ServiceError {
	return sharederrors.NewServiceErrorWithFields(
		http.StatusUnprocessableEntity,
		"",
		map[string]string{
			"Password": "New password must be different from your current password",
		},
	)
}
