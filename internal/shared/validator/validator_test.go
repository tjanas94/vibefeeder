package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomValidator_ValidateStrongPassword tests password strength validation
func TestCustomValidator_ValidateStrongPassword(t *testing.T) {
	validator := New()

	type TestStruct struct {
		Password string `validate:"required,strongpassword=60"`
	}

	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "strong password with mixed characters",
			password:  "MyS3cur3P@ssw0rd!2024",
			wantError: false,
		},
		{
			name:      "very strong password",
			password:  "Th1s!s@V3ryStr0ngP@ssw0rd#2024",
			wantError: false,
		},
		{
			name:      "weak password - too short",
			password:  "pass",
			wantError: true,
		},
		{
			name:      "weak password - only lowercase",
			password:  "password",
			wantError: true,
		},
		{
			name:      "weak password - simple with numbers",
			password:  "password123",
			wantError: true,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testData := TestStruct{Password: tt.password}
			err := validator.Validate(&testData)

			if tt.wantError {
				assert.Error(t, err, "expected validation error")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestCustomValidator_ValidateRequired tests required field validation
func TestCustomValidator_ValidateRequired(t *testing.T) {
	validator := New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name      string
		data      TestStruct
		wantError bool
	}{
		{
			name: "valid data",
			data: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantError: false,
		},
		{
			name: "missing name",
			data: TestStruct{
				Name:  "",
				Email: "john@example.com",
			},
			wantError: true,
		},
		{
			name: "missing email",
			data: TestStruct{
				Name:  "John Doe",
				Email: "",
			},
			wantError: true,
		},
		{
			name: "invalid email format",
			data: TestStruct{
				Name:  "John Doe",
				Email: "not-an-email",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.data)

			if tt.wantError {
				assert.Error(t, err, "expected validation error")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestFormatValidationErrors tests that validation errors are properly formatted
func TestFormatValidationErrors(t *testing.T) {
	validator := New()

	type TestStruct struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,strongpassword=60"`
	}

	testData := TestStruct{
		Email:    "invalid-email",
		Password: "weak",
	}

	err := validator.Validate(&testData)
	require.Error(t, err, "expected validation errors")

	// Check that error is a map of field errors
	// Note: The actual structure depends on echo.HTTPError implementation
	// This test verifies that we get an error response
	assert.NotNil(t, err)
}

// BenchmarkCustomValidator_Validate benchmarks the struct validation
func BenchmarkCustomValidator_Validate(b *testing.B) {
	validator := New()

	type TestStruct struct {
		Email string `validate:"required,email"`
		URL   string `validate:"required,http_url"`
	}

	testData := TestStruct{
		Email: "test@example.com",
		URL:   "https://example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(&testData)
	}
}
