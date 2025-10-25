package auth

import (
	"crypto/subtle"
	"strings"
)

// validateRegistrationCode performs constant-time comparison of registration codes
// to prevent timing attacks. Returns true if codes match.
func validateRegistrationCode(providedCode, expectedCode string) bool {
	// If no registration code is configured, always allow
	if expectedCode == "" {
		return true
	}

	return subtle.ConstantTimeCompare([]byte(providedCode), []byte(expectedCode)) == 1
}

// isUserExistsError checks if the error indicates a user already exists
func isUserExistsError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "user_already_exists") ||
		strings.Contains(errMsg, "User already registered")
}

// isSamePasswordError checks if the error indicates user tried to change password to the same value
func isSamePasswordError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "same_password") ||
		strings.Contains(errMsg, "should be different from the old password")
}

// isInvalidCredentialsError checks if the error indicates invalid email or password
func isInvalidCredentialsError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "invalid_grant") ||
		strings.Contains(errMsg, "Invalid login credentials") ||
		strings.Contains(errMsg, "invalid_credentials") ||
		strings.Contains(errMsg, "Invalid email or password")
}

// isInvalidTokenError checks if the error indicates an invalid, expired, or malformed token
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "invalid_token") ||
		strings.Contains(errMsg, "expired_token") ||
		strings.Contains(errMsg, "token_not_found") ||
		strings.Contains(errMsg, "Token not found") ||
		strings.Contains(errMsg, "Invalid token") ||
		strings.Contains(errMsg, "Expired token")
}
