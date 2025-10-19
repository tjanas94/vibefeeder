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
	return strings.Contains(errMsg, "already registered") ||
		strings.Contains(errMsg, "email address already")
}
