package validator

import (
	"errors"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFieldErrors_NilError tests that nil error returns nil map
func TestParseFieldErrors_NilError(t *testing.T) {
	result := ParseFieldErrors(nil)
	assert.Nil(t, result)
}

// TestParseFieldErrors_NotHTTPError tests that non-HTTPError returns nil
func TestParseFieldErrors_NotHTTPError(t *testing.T) {
	err := errors.New("standard error")
	result := ParseFieldErrors(err)
	assert.Nil(t, result)
}

// TestParseFieldErrors_RegularError tests that plain error returns nil
func TestParseFieldErrors_RegularError(t *testing.T) {
	err := errors.New("some error message")
	result := ParseFieldErrors(err)
	assert.Nil(t, result)
}

// TestParseFieldErrors_HTTPErrorWithStringMessage tests HTTPError with string message returns nil
func TestParseFieldErrors_HTTPErrorWithStringMessage(t *testing.T) {
	httpErr := echo.NewHTTPError(400, "bad request")
	result := ParseFieldErrors(httpErr)
	assert.Nil(t, result)
}

// TestParseFieldErrors_HTTPErrorWithIntMessage tests HTTPError with int message returns nil
func TestParseFieldErrors_HTTPErrorWithIntMessage(t *testing.T) {
	httpErr := echo.NewHTTPError(400, 123)
	result := ParseFieldErrors(httpErr)
	assert.Nil(t, result)
}

// TestParseFieldErrors_HTTPErrorWithMapMessage tests HTTPError with map message returns the map
func TestParseFieldErrors_HTTPErrorWithMapMessage(t *testing.T) {
	expectedErrors := map[string]string{
		"email":    "invalid email format",
		"password": "password too short",
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
}

// TestParseFieldErrors_HTTPErrorWithEmptyMap tests HTTPError with empty map
func TestParseFieldErrors_HTTPErrorWithEmptyMap(t *testing.T) {
	expectedErrors := map[string]string{}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
	assert.Len(t, result, 0)
}

// TestParseFieldErrors_SingleFieldError tests single field error
func TestParseFieldErrors_SingleFieldError(t *testing.T) {
	expectedErrors := map[string]string{
		"username": "username is required",
	}
	httpErr := echo.NewHTTPError(422, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "username is required", result["username"])
}

// TestParseFieldErrors_MultipleFieldErrors tests multiple field errors
func TestParseFieldErrors_MultipleFieldErrors(t *testing.T) {
	expectedErrors := map[string]string{
		"email":       "email is invalid",
		"password":    "password must be at least 8 characters",
		"username":    "username is required",
		"confirm_pwd": "passwords do not match",
		"terms":       "you must accept the terms",
	}
	httpErr := echo.NewHTTPError(422, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
	assert.Len(t, result, 5)

	// Verify all fields are present
	for fieldName, fieldError := range expectedErrors {
		assert.Equal(t, fieldError, result[fieldName])
	}
}

// TestParseFieldErrors_SpecialCharactersInErrors tests field errors with special characters
func TestParseFieldErrors_SpecialCharactersInErrors(t *testing.T) {
	expectedErrors := map[string]string{
		"name":    "name must not contain: @#$%^&*()",
		"bio":     "bio contains invalid characters: \"<>\"",
		"url":     "must be a valid URL (e.g., https://example.com)",
		"phone":   "phone format: +1-234-567-8900 or 1234567890",
		"unicode": "must contain: café, naïve, résumé",
	}
	httpErr := echo.NewHTTPError(422, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
}

// TestParseFieldErrors_EmptyStringErrorMessages tests field errors with empty string messages
func TestParseFieldErrors_EmptyStringErrorMessages(t *testing.T) {
	expectedErrors := map[string]string{
		"field1": "",
		"field2": "some error",
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
	assert.Equal(t, "", result["field1"])
}

// TestParseFieldErrors_LongErrorMessages tests field errors with long messages
func TestParseFieldErrors_LongErrorMessages(t *testing.T) {
	longMessage := "This is a very long error message that provides detailed information about " +
		"why the field validation failed. It might include instructions on how to correct the error " +
		"and examples of valid input. The message continues with more details and context."

	expectedErrors := map[string]string{
		"description": longMessage,
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, longMessage, result["description"])
}

// TestParseFieldErrors_FieldNamesWithSpecialCharacters tests field names with underscores and numbers
func TestParseFieldErrors_FieldNamesWithSpecialCharacters(t *testing.T) {
	expectedErrors := map[string]string{
		"field_name_1":      "error 1",
		"field_name_2":      "error 2",
		"user_email_backup": "error 3",
		"field123":          "error 4",
		"_private":          "error 5",
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result := ParseFieldErrors(httpErr)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErrors, result)
}

// TestParseFieldErrors_HTTPErrorVariousStatusCodes tests different HTTP status codes
func TestParseFieldErrors_HTTPErrorVariousStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", 400},
		{"401 Unauthorized", 401},
		{"402 Payment Required", 402},
		{"403 Forbidden", 403},
		{"404 Not Found", 404},
		{"409 Conflict", 409},
		{"422 Unprocessable Entity", 422},
		{"500 Internal Server Error", 500},
		{"503 Service Unavailable", 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedErrors := map[string]string{
				"field": "error message",
			}
			httpErr := echo.NewHTTPError(tt.statusCode, expectedErrors)

			result := ParseFieldErrors(httpErr)

			assert.NotNil(t, result)
			assert.Equal(t, expectedErrors, result)
		})
	}
}

// TestParseFieldErrors_HTTPErrorWithOtherMapTypes tests HTTPError with map but wrong value type
func TestParseFieldErrors_HTTPErrorWithOtherMapTypes(t *testing.T) {
	// Create an HTTPError with a map that has non-string values
	mapWithWrongType := map[string]any{
		"field1": "valid string",
		"field2": 123, // not a string
	}
	httpErr := echo.NewHTTPError(400, mapWithWrongType)

	result := ParseFieldErrors(httpErr)

	// This should return nil because the map type is map[string]any, not map[string]string
	assert.Nil(t, result)
}

// TestParseFieldErrors_HTTPErrorWithNestedMapMessage tests HTTPError with nested map (should fail type assertion)
func TestParseFieldErrors_HTTPErrorWithNestedMapMessage(t *testing.T) {
	nestedMap := map[string]any{
		"errors": map[string]string{
			"email": "invalid",
		},
	}
	httpErr := echo.NewHTTPError(400, nestedMap)

	result := ParseFieldErrors(httpErr)

	// Should return nil because nestedMap is not map[string]string
	assert.Nil(t, result)
}

// TestParseFieldErrors_WrappedHTTPError tests wrapped HTTPError
func TestParseFieldErrors_WrappedHTTPError(t *testing.T) {
	expectedErrors := map[string]string{
		"username": "already taken",
	}
	httpErr := echo.NewHTTPError(409, expectedErrors)
	wrappedErr := errors.Join(errors.New("outer error"), httpErr)

	// ParseFieldErrors only handles direct HTTPError, not wrapped ones
	result := ParseFieldErrors(wrappedErr)

	// Should still return nil because the error is wrapped
	assert.Nil(t, result)
}

// TestParseFieldErrors_PreservesMapReference tests that ParseFieldErrors returns the actual map, not a copy
func TestParseFieldErrors_PreservesMapReference(t *testing.T) {
	originalErrors := map[string]string{
		"field": "error",
	}
	httpErr := echo.NewHTTPError(400, originalErrors)

	result := ParseFieldErrors(httpErr)

	require.NotNil(t, result)

	// The returned map should be the same instance
	assert.Equal(t, &originalErrors, &result)

	// Modifying the returned map should affect the original
	result["field"] = "modified"
	assert.Equal(t, "modified", originalErrors["field"])
}

// TestParseFieldErrors_ConsecutiveCalls tests multiple consecutive calls with same error
func TestParseFieldErrors_ConsecutiveCalls(t *testing.T) {
	expectedErrors := map[string]string{
		"email": "invalid email",
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	result1 := ParseFieldErrors(httpErr)
	result2 := ParseFieldErrors(httpErr)

	assert.Equal(t, result1, result2)
	assert.Equal(t, expectedErrors, result1)
	assert.Equal(t, expectedErrors, result2)
}

// TestParseFieldErrors_IntegrationWithEchoValidationError simulates typical Echo validation error
func TestParseFieldErrors_IntegrationWithEchoValidationError(t *testing.T) {
	// Simulate typical validation error that would come from Echo validator
	validationErrors := map[string]string{
		"email":      "email field is invalid",
		"name":       "name field is required",
		"age":        "age must be a number",
		"accept_tos": "you must accept the terms of service",
	}
	err := echo.NewHTTPError(400, validationErrors)

	result := ParseFieldErrors(err)

	assert.NotNil(t, result)
	assert.Len(t, result, 4)

	// Verify all expected fields are present
	assert.Equal(t, "email field is invalid", result["email"])
	assert.Equal(t, "name field is required", result["name"])
	assert.Equal(t, "age must be a number", result["age"])
	assert.Equal(t, "you must accept the terms of service", result["accept_tos"])
}

// BenchmarkParseFieldErrors_Valid benchmarks parsing valid field errors
func BenchmarkParseFieldErrors_Valid(b *testing.B) {
	expectedErrors := map[string]string{
		"email":    "invalid email",
		"password": "password too short",
		"username": "username taken",
	}
	httpErr := echo.NewHTTPError(400, expectedErrors)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseFieldErrors(httpErr)
	}
}

// BenchmarkParseFieldErrors_NilError benchmarks parsing nil error
func BenchmarkParseFieldErrors_NilError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseFieldErrors(nil)
	}
}

// BenchmarkParseFieldErrors_NonHTTPError benchmarks parsing non-HTTPError
func BenchmarkParseFieldErrors_NonHTTPError(b *testing.B) {
	err := errors.New("standard error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseFieldErrors(err)
	}
}
