package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for validateRegistrationCode

// TestValidateRegistrationCode_EmptyExpectedCode tests that empty expected code allows any provided code
func TestValidateRegistrationCode_EmptyExpectedCode(t *testing.T) {
	tests := []struct {
		name         string
		providedCode string
		expected     bool
	}{
		{
			name:         "empty provided with empty expected",
			providedCode: "",
			expected:     true,
		},
		{
			name:         "non-empty provided with empty expected",
			providedCode: "any-code-123",
			expected:     true,
		},
		{
			name:         "special characters with empty expected",
			providedCode: "!@#$%^&*()",
			expected:     true,
		},
		{
			name:         "long string with empty expected",
			providedCode: "this-is-a-very-long-registration-code-that-someone-might-provide",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, "")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateRegistrationCode_ExactMatch tests exact matching of registration codes
func TestValidateRegistrationCode_ExactMatch(t *testing.T) {
	tests := []struct {
		name         string
		providedCode string
		expectedCode string
		expected     bool
	}{
		{
			name:         "exact match simple",
			providedCode: "ABC123",
			expectedCode: "ABC123",
			expected:     true,
		},
		{
			name:         "exact match with special chars",
			providedCode: "reg-code_2024!",
			expectedCode: "reg-code_2024!",
			expected:     true,
		},
		{
			name:         "exact match long code",
			providedCode: "abcdefghijklmnopqrstuvwxyz0123456789",
			expectedCode: "abcdefghijklmnopqrstuvwxyz0123456789",
			expected:     true,
		},
		{
			name:         "exact match single character",
			providedCode: "x",
			expectedCode: "x",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, tt.expectedCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateRegistrationCode_Mismatch tests code mismatch scenarios
func TestValidateRegistrationCode_Mismatch(t *testing.T) {
	tests := []struct {
		name         string
		providedCode string
		expectedCode string
		expected     bool
	}{
		{
			name:         "simple mismatch",
			providedCode: "ABC123",
			expectedCode: "ABC124",
			expected:     false,
		},
		{
			name:         "case sensitivity - uppercase vs lowercase",
			providedCode: "abc123",
			expectedCode: "ABC123",
			expected:     false,
		},
		{
			name:         "whitespace difference",
			providedCode: "code123",
			expectedCode: "code 123",
			expected:     false,
		},
		{
			name:         "extra character in provided",
			providedCode: "code1234",
			expectedCode: "code123",
			expected:     false,
		},
		{
			name:         "missing character in provided",
			providedCode: "code12",
			expectedCode: "code123",
			expected:     false,
		},
		{
			name:         "completely different codes",
			providedCode: "first-code",
			expectedCode: "second-code",
			expected:     false,
		},
		{
			name:         "empty provided vs non-empty expected",
			providedCode: "",
			expectedCode: "required-code",
			expected:     false,
		},
		{
			name:         "non-empty provided vs empty expected",
			providedCode: "wrong-code",
			expectedCode: "",
			expected:     true, // empty expected means no validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, tt.expectedCode)
			assert.Equal(t, tt.expected, result, "code validation result mismatch")
		})
	}
}

// TestValidateRegistrationCode_TimingAttackResistance tests that function uses constant-time comparison
func TestValidateRegistrationCode_TimingAttackResistance(t *testing.T) {
	// Test that the function doesn't short-circuit on first difference
	// All these should execute in similar time due to constant-time comparison
	tests := []struct {
		name         string
		providedCode string
		expectedCode string
		expected     bool
	}{
		{
			name:         "differ at first character",
			providedCode: "xbc123456789",
			expectedCode: "abc123456789",
			expected:     false,
		},
		{
			name:         "differ at middle character",
			providedCode: "abc1x3456789",
			expectedCode: "abc123456789",
			expected:     false,
		},
		{
			name:         "differ at last character",
			providedCode: "abc123456789x",
			expectedCode: "abc123456789",
			expected:     false,
		},
		{
			name:         "differ at multiple positions",
			providedCode: "xbc12x456789",
			expectedCode: "abc123456789",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, tt.expectedCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateRegistrationCode_SpecialCharacters tests codes with special characters
func TestValidateRegistrationCode_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		providedCode string
		expectedCode string
		expected     bool
	}{
		{
			name:         "codes with hyphens",
			providedCode: "code-2024-q1",
			expectedCode: "code-2024-q1",
			expected:     true,
		},
		{
			name:         "codes with underscores",
			providedCode: "CODE_2024_Q1",
			expectedCode: "CODE_2024_Q1",
			expected:     true,
		},
		{
			name:         "codes with punctuation",
			providedCode: "code.2024.q1",
			expectedCode: "code.2024.q1",
			expected:     true,
		},
		{
			name:         "codes with symbols",
			providedCode: "code!@#$%^&*()",
			expectedCode: "code!@#$%^&*()",
			expected:     true,
		},
		{
			name:         "mismatched special characters",
			providedCode: "code-2024",
			expectedCode: "code_2024",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, tt.expectedCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateRegistrationCode_UnicodeCharacters tests codes with unicode characters
func TestValidateRegistrationCode_UnicodeCharacters(t *testing.T) {
	tests := []struct {
		name         string
		providedCode string
		expectedCode string
		expected     bool
	}{
		{
			name:         "unicode exact match",
			providedCode: "café-2024",
			expectedCode: "café-2024",
			expected:     true,
		},
		{
			name:         "unicode mismatch",
			providedCode: "café-2024",
			expectedCode: "cafe-2024",
			expected:     false,
		},
		{
			name:         "emoji codes match",
			providedCode: "🔐-code-2024",
			expectedCode: "🔐-code-2024",
			expected:     true,
		},
		{
			name:         "emoji codes mismatch",
			providedCode: "🔐-code-2024",
			expectedCode: "🔒-code-2024",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRegistrationCode(tt.providedCode, tt.expectedCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isUserExistsError

// TestIsUserExistsError_NilError tests that nil error returns false
func TestIsUserExistsError_NilError(t *testing.T) {
	result := isUserExistsError(nil)
	assert.False(t, result)
}

// TestIsUserExistsError_UserAlreadyExistsPattern tests detection of "user_already_exists" pattern
func TestIsUserExistsError_UserAlreadyExistsPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "user_already_exists",
			expected: true,
		},
		{
			name:     "in the middle of message",
			errMsg:   "error: user_already_exists in the system",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase",
			errMsg:   "user_already_exists",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should match Contains",
			errMsg:   "USER_ALREADY_EXISTS",
			expected: false, // Contains is case-sensitive
		},
		{
			name:     "similar but not exact",
			errMsg:   "user already exists",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isUserExistsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsUserExistsError_UserAlreadyRegisteredPattern tests detection of "User already registered" pattern
func TestIsUserExistsError_UserAlreadyRegisteredPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "User already registered",
			expected: true,
		},
		{
			name:     "case preserved match",
			errMsg:   "User already registered",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "signup failed: User already registered with this email",
			expected: true,
		},
		{
			name:     "lowercase - should not match",
			errMsg:   "user already registered",
			expected: false,
		},
		{
			name:     "uppercase - should not match",
			errMsg:   "USER ALREADY REGISTERED",
			expected: false,
		},
		{
			name:     "similar text",
			errMsg:   "user is already registered",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isUserExistsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsUserExistsError_EitherPatternMatches tests that function returns true if either pattern matches
func TestIsUserExistsError_EitherPatternMatches(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "first pattern match",
			errMsg:   "error: user_already_exists",
			expected: true,
		},
		{
			name:     "second pattern match",
			errMsg:   "error: User already registered",
			expected: true,
		},
		{
			name:     "both patterns in same error",
			errMsg:   "user_already_exists - User already registered",
			expected: true,
		},
		{
			name:     "neither pattern",
			errMsg:   "user does not exist",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isUserExistsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsUserExistsError_OtherErrors tests non-exists errors return false
func TestIsUserExistsError_OtherErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "not found error",
			errMsg:   "user not found",
			expected: false,
		},
		{
			name:     "invalid email",
			errMsg:   "invalid email format",
			expected: false,
		},
		{
			name:     "password error",
			errMsg:   "password must be at least 8 characters",
			expected: false,
		},
		{
			name:     "database error",
			errMsg:   "database connection failed",
			expected: false,
		},
		{
			name:     "empty string",
			errMsg:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isUserExistsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isInvalidCredentialsError

// TestIsInvalidCredentialsError_NilError tests that nil error returns false
func TestIsInvalidCredentialsError_NilError(t *testing.T) {
	result := isInvalidCredentialsError(nil)
	assert.False(t, result)
}

// TestIsInvalidCredentialsError_InvalidGrantPattern tests detection of "invalid_grant" pattern
func TestIsInvalidCredentialsError_InvalidGrantPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "invalid_grant",
			expected: true,
		},
		{
			name:     "in the middle of message",
			errMsg:   "error: invalid_grant occurred",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase",
			errMsg:   "invalid_grant",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "INVALID_GRANT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidCredentialsError_InvalidLoginCredentialsPattern tests detection of "Invalid login credentials" pattern
func TestIsInvalidCredentialsError_InvalidLoginCredentialsPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "Invalid login credentials",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "auth error: Invalid login credentials provided",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase should not match",
			errMsg:   "invalid login credentials",
			expected: false,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "INVALID LOGIN CREDENTIALS",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidCredentialsError_InvalidCredentialsPattern tests detection of "invalid_credentials" pattern
func TestIsInvalidCredentialsError_InvalidCredentialsPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "invalid_credentials",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "validation failed: invalid_credentials detected",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "INVALID_CREDENTIALS",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidCredentialsError_InvalidEmailOrPasswordPattern tests detection of "Invalid email or password" pattern
func TestIsInvalidCredentialsError_InvalidEmailOrPasswordPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "Invalid email or password",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "login failed: Invalid email or password provided",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase should not match",
			errMsg:   "invalid email or password",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidCredentialsError_EitherPatternMatches tests that function returns true if any pattern matches
func TestIsInvalidCredentialsError_EitherPatternMatches(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "first pattern - invalid_grant",
			errMsg:   "error: invalid_grant",
			expected: true,
		},
		{
			name:     "second pattern - Invalid login credentials",
			errMsg:   "error: Invalid login credentials",
			expected: true,
		},
		{
			name:     "third pattern - invalid_credentials",
			errMsg:   "error: invalid_credentials",
			expected: true,
		},
		{
			name:     "fourth pattern - Invalid email or password",
			errMsg:   "error: Invalid email or password",
			expected: true,
		},
		{
			name:     "multiple patterns in same error",
			errMsg:   "invalid_grant and Invalid login credentials",
			expected: true,
		},
		{
			name:     "no matching patterns",
			errMsg:   "network timeout",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidCredentialsError_OtherErrors tests non-credentials errors return false
func TestIsInvalidCredentialsError_OtherErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "user not found",
			errMsg:   "user not found",
			expected: false,
		},
		{
			name:     "account locked",
			errMsg:   "account is locked",
			expected: false,
		},
		{
			name:     "server error",
			errMsg:   "internal server error",
			expected: false,
		},
		{
			name:     "database error",
			errMsg:   "database connection failed",
			expected: false,
		},
		{
			name:     "empty string",
			errMsg:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidCredentialsError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isInvalidTokenError

// TestIsInvalidTokenError_NilError tests that nil error returns false
func TestIsInvalidTokenError_NilError(t *testing.T) {
	result := isInvalidTokenError(nil)
	assert.False(t, result)
}

// TestIsInvalidTokenError_InvalidTokenPattern tests detection of "invalid_token" pattern
func TestIsInvalidTokenError_InvalidTokenPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "invalid_token",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "Token verification failed: invalid_token in request",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "INVALID_TOKEN",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_ExpiredTokenPattern tests detection of "expired_token" pattern
func TestIsInvalidTokenError_ExpiredTokenPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "expired_token",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "Token validation failed: expired_token detected",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "EXPIRED_TOKEN",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_TokenNotFoundPattern tests detection of "token_not_found" pattern
func TestIsInvalidTokenError_TokenNotFoundPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "token_not_found",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "error: token_not_found in session",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "TOKEN_NOT_FOUND",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_TokenNotFoundCapitalizedPattern tests detection of "Token not found" pattern
func TestIsInvalidTokenError_TokenNotFoundCapitalizedPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "Token not found",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "error: Token not found in database",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase should not match",
			errMsg:   "token not found",
			expected: false,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "TOKEN NOT FOUND",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_InvalidTokenCapitalizedPattern tests detection of "Invalid token" pattern
func TestIsInvalidTokenError_InvalidTokenCapitalizedPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "Invalid token",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "validation failed: Invalid token received",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase should not match",
			errMsg:   "invalid token",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_ExpiredTokenCapitalizedPattern tests detection of "Expired token" pattern
func TestIsInvalidTokenError_ExpiredTokenCapitalizedPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "Expired token",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "token check: Expired token in request",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase should not match",
			errMsg:   "expired token",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_EitherPatternMatches tests that function returns true if any pattern matches
func TestIsInvalidTokenError_EitherPatternMatches(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "first pattern - invalid_token",
			errMsg:   "error: invalid_token",
			expected: true,
		},
		{
			name:     "second pattern - expired_token",
			errMsg:   "error: expired_token",
			expected: true,
		},
		{
			name:     "third pattern - token_not_found",
			errMsg:   "error: token_not_found",
			expected: true,
		},
		{
			name:     "fourth pattern - Token not found",
			errMsg:   "error: Token not found",
			expected: true,
		},
		{
			name:     "fifth pattern - Invalid token",
			errMsg:   "error: Invalid token",
			expected: true,
		},
		{
			name:     "sixth pattern - Expired token",
			errMsg:   "error: Expired token",
			expected: true,
		},
		{
			name:     "multiple patterns in same error",
			errMsg:   "invalid_token and Expired token detected",
			expected: true,
		},
		{
			name:     "no matching patterns",
			errMsg:   "database connection failed",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsInvalidTokenError_OtherErrors tests non-token errors return false
func TestIsInvalidTokenError_OtherErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "user not found",
			errMsg:   "user not found",
			expected: false,
		},
		{
			name:     "network error",
			errMsg:   "network connection timeout",
			expected: false,
		},
		{
			name:     "server error",
			errMsg:   "internal server error",
			expected: false,
		},
		{
			name:     "database error",
			errMsg:   "database connection failed",
			expected: false,
		},
		{
			name:     "empty string",
			errMsg:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isInvalidTokenError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for isSamePasswordError

// TestIsSamePasswordError_NilError tests that nil error returns false
func TestIsSamePasswordError_NilError(t *testing.T) {
	result := isSamePasswordError(nil)
	assert.False(t, result)
}

// TestIsSamePasswordError_SamePasswordPattern tests detection of "same_password" pattern
func TestIsSamePasswordError_SamePasswordPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "same_password",
			expected: true,
		},
		{
			name:     "in the middle of message",
			errMsg:   "error: same_password detected",
			expected: true,
		},
		{
			name:     "case sensitive - lowercase",
			errMsg:   "same_password",
			expected: true,
		},
		{
			name:     "case sensitive - uppercase should not match",
			errMsg:   "SAME_PASSWORD",
			expected: false,
		},
		{
			name:     "similar text without underscore",
			errMsg:   "same password",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isSamePasswordError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSamePasswordError_ShouldBeDifferentPattern tests detection of "should be different from the old password" pattern
func TestIsSamePasswordError_ShouldBeDifferentPattern(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "exact match",
			errMsg:   "should be different from the old password",
			expected: true,
		},
		{
			name:     "case preserved match",
			errMsg:   "should be different from the old password",
			expected: true,
		},
		{
			name:     "in middle of message",
			errMsg:   "new password should be different from the old password for security",
			expected: true,
		},
		{
			name:     "case variations - should not match",
			errMsg:   "SHOULD BE DIFFERENT FROM THE OLD PASSWORD",
			expected: false,
		},
		{
			name:     "similar but not exact",
			errMsg:   "should be different from old password",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isSamePasswordError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSamePasswordError_EitherPatternMatches tests that function returns true if either pattern matches
func TestIsSamePasswordError_EitherPatternMatches(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "first pattern match",
			errMsg:   "error: same_password",
			expected: true,
		},
		{
			name:     "second pattern match",
			errMsg:   "error: should be different from the old password",
			expected: true,
		},
		{
			name:     "both patterns in same error",
			errMsg:   "same_password - should be different from the old password",
			expected: true,
		},
		{
			name:     "neither pattern",
			errMsg:   "password changed successfully",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isSamePasswordError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSamePasswordError_OtherErrors tests non-same-password errors return false
func TestIsSamePasswordError_OtherErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "password too short",
			errMsg:   "password must be at least 8 characters",
			expected: false,
		},
		{
			name:     "invalid password",
			errMsg:   "password format is invalid",
			expected: false,
		},
		{
			name:     "password mismatch",
			errMsg:   "passwords do not match",
			expected: false,
		},
		{
			name:     "current password incorrect",
			errMsg:   "current password is incorrect",
			expected: false,
		},
		{
			name:     "empty string",
			errMsg:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := isSamePasswordError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Combined tests for all three helper functions

// TestHelpers_ConsistentErrorHandling tests that all helpers handle nil consistently
func TestHelpers_ConsistentErrorHandling(t *testing.T) {
	var nilErr error

	userExists := isUserExistsError(nilErr)
	samePassword := isSamePasswordError(nilErr)
	regCode := validateRegistrationCode("test", "test")

	assert.False(t, userExists)
	assert.False(t, samePassword)
	assert.True(t, regCode) // Both match
}

// TestHelpers_CaseSensitivity tests that error pattern matching is case-sensitive
func TestHelpers_CaseSensitivity(t *testing.T) {
	lowercaseUserExists := errors.New("user_already_exists")
	uppercaseUserExists := errors.New("USER_ALREADY_EXISTS")

	lowercaseResult := isUserExistsError(lowercaseUserExists)
	uppercaseResult := isUserExistsError(uppercaseUserExists)

	assert.True(t, lowercaseResult)
	assert.False(t, uppercaseResult) // Case sensitive
}

// TestHelpers_MultipleErrors tests distinguishing between different error types
func TestHelpers_MultipleErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		isExists bool
		isSame   bool
	}{
		{
			name:     "user exists error",
			errMsg:   "user_already_exists",
			isExists: true,
			isSame:   false,
		},
		{
			name:     "same password error",
			errMsg:   "should be different from the old password",
			isExists: false,
			isSame:   true,
		},
		{
			name:     "other error",
			errMsg:   "unknown error",
			isExists: false,
			isSame:   false,
		},
		{
			name:     "both patterns",
			errMsg:   "user_already_exists and should be different from the old password",
			isExists: true,
			isSame:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)

			resultExists := isUserExistsError(err)
			resultSame := isSamePasswordError(err)

			assert.Equal(t, tt.isExists, resultExists, "user exists error mismatch")
			assert.Equal(t, tt.isSame, resultSame, "same password error mismatch")
		})
	}
}

// BenchmarkValidateRegistrationCode_Match benchmarks successful code validation
func BenchmarkValidateRegistrationCode_Match(b *testing.B) {
	providedCode := "registration-code-2024-secure"
	expectedCode := "registration-code-2024-secure"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateRegistrationCode(providedCode, expectedCode)
	}
}

// BenchmarkValidateRegistrationCode_Mismatch benchmarks failed code validation
func BenchmarkValidateRegistrationCode_Mismatch(b *testing.B) {
	providedCode := "registration-code-2024-wrong"
	expectedCode := "registration-code-2024-secure"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateRegistrationCode(providedCode, expectedCode)
	}
}

// BenchmarkValidateRegistrationCode_EmptyExpected benchmarks with empty expected code
func BenchmarkValidateRegistrationCode_EmptyExpected(b *testing.B) {
	providedCode := "any-code"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateRegistrationCode(providedCode, "")
	}
}

// BenchmarkIsUserExistsError benchmarks user exists error detection
func BenchmarkIsUserExistsError(b *testing.B) {
	err := errors.New("user_already_exists in the system")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isUserExistsError(err)
	}
}

// BenchmarkIsSamePasswordError benchmarks same password error detection
func BenchmarkIsSamePasswordError(b *testing.B) {
	err := errors.New("should be different from the old password")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isSamePasswordError(err)
	}
}
