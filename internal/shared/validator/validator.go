package validator

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

// CustomValidator wraps the go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

// New creates a new CustomValidator instance
func New() *CustomValidator {
	v := validator.New()

	// Register custom validation for strong passwords
	_ = v.RegisterValidation("strongpassword", validateStrongPassword)

	return &CustomValidator{
		validator: v,
	}
}

// validateStrongPassword validates password strength using entropy
// Tag format: strongpassword=50 (where 50 is minimum entropy in bits)
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Get entropy parameter from tag (e.g., "strongpassword=50")
	entropyParam := fl.Param()
	if entropyParam == "" {
		// Default to 50 bits if no parameter provided
		entropyParam = "50"
	}

	// Parse entropy value
	var minEntropy float64
	if _, err := fmt.Sscanf(entropyParam, "%f", &minEntropy); err != nil {
		return false
	}

	// Validate password entropy
	err := passwordvalidator.Validate(password, minEntropy)
	return err == nil
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
	case "http_url":
		return "Must be a valid HTTP or HTTPS URL"
	case "strongpassword":
		return "Make password longer or add numbers and symbols"
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
	case "eqfield":
		return fmt.Sprintf("Must match %s", param)
	default:
		return fmt.Sprintf("Failed validation: %s", tag)
	}
}
