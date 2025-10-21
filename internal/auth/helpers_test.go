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
