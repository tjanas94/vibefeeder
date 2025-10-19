package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when login credentials are incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUserAlreadyExists is returned when trying to register with an existing email
	ErrUserAlreadyExists = errors.New("user with this email already exists")

	// ErrInvalidToken is returned when email confirmation or password reset token is invalid
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrSessionExpired is returned when the session has expired
	ErrSessionExpired = errors.New("session expired")

	// ErrInvalidRegistrationCode is returned when registration code is incorrect
	ErrInvalidRegistrationCode = errors.New("invalid registration code")

	// ErrSamePassword is returned when trying to change password to the same value
	ErrSamePassword = errors.New("new password must be different from the current password")
)
