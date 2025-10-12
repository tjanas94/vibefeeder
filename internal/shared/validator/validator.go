package validator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
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
			return echo.NewHTTPError(
				http.StatusBadRequest,
				formatValidationErrors(validationErrors),
			)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable message
func formatValidationErrors(errs validator.ValidationErrors) string {
	var messages []string
	for _, err := range errs {
		messages = append(messages, formatFieldError(err))
	}
	return strings.Join(messages, "; ")
}

// formatFieldError formats a single field error
func formatFieldError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters long", field, param)
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("Field '%s' must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of [%s]", field, param)
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field)
	case "datetime":
		return fmt.Sprintf("Field '%s' must be a valid datetime in format %s", field, param)
	default:
		return fmt.Sprintf("Field '%s' failed validation on tag '%s'", field, tag)
	}
}
