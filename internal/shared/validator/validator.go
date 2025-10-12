package validator

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

// New creates a new CustomValidator instance
func New() *CustomValidator {
	v := validator.New()

	// Register custom validation tags here if needed
	// Example: v.RegisterValidation("custom_tag", customValidationFunc)

	return &CustomValidator{
		validator: v,
	}
}

// Validate validates a struct based on validation tags
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// Type assert to validator.ValidationErrors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Return structured error map for easy parsing
			return echo.NewHTTPError(
				http.StatusBadRequest,
				formatValidationErrors(validationErrors),
			)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// formatValidationErrors formats validation errors into a map
// Format: {"FieldName": "error message"}
func formatValidationErrors(errs validator.ValidationErrors) map[string]string {
	errorMap := make(map[string]string)
	for _, err := range errs {
		errorMap[err.Field()] = formatFieldError(err)
	}
	return errorMap
}

// formatFieldError formats a single field error
func formatFieldError(err validator.FieldError) string {
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "url":
		return "Must be a valid URL"
	case "min":
		return fmt.Sprintf("Must be at least %s characters long", param)
	case "max":
		return fmt.Sprintf("Must be at most %s characters long", param)
	case "len":
		return fmt.Sprintf("Must be exactly %s characters long", param)
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", param)
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", param)
	case "gt":
		return fmt.Sprintf("Must be greater than %s", param)
	case "lt":
		return fmt.Sprintf("Must be less than %s", param)
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", param)
	case "uuid":
		return "Must be a valid UUID"
	case "datetime":
		return fmt.Sprintf("Must be a valid datetime in format %s", param)
	default:
		return fmt.Sprintf("Failed validation: %s", tag)
	}
}

// IsValidUUID checks if a string is a valid UUID format.
// Returns true if valid, false otherwise.
// Use this for validating path/query parameters that should be UUIDs.
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
