# Validator

Custom validator wrapper for Echo using `go-playground/validator/v10` with tag-based validation support.

## Usage in Handlers

### Basic Example

```go
package myfeature

import (
    "net/http"
    "github.com/labstack/echo/v4"
)

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"gte=18,lte=100"`
}

func (h *Handler) CreateUser(c echo.Context) error {
    req := new(CreateUserRequest)

    // Bind request data
    if err := c.Bind(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
    }

    // Validate using tags
    if err := c.Validate(req); err != nil {
        return err // Returns formatted validation errors
    }

    // Process valid request...
    return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}
```

## Available Validation Tags

### String Validations

- `required` - Field must be present and non-empty
- `email` - Must be a valid email address
- `url` - Must be a valid URL
- `min=N` - Minimum string length
- `max=N` - Maximum string length
- `len=N` - Exact string length
- `uuid` - Must be a valid UUID
- `oneof=val1 val2` - Must be one of the specified values

### Numeric Validations

- `gte=N` - Greater than or equal to N
- `lte=N` - Less than or equal to N
- `gt=N` - Greater than N
- `lt=N` - Less than N

### Date/Time Validations

- `datetime=2006-01-02` - Must match the specified datetime format

### Example Struct

```go
type FeedRequest struct {
    URL         string `json:"url" validate:"required,url"`
    Title       string `json:"title" validate:"required,min=3,max=100"`
    Category    string `json:"category" validate:"oneof=tech news sports"`
    RefreshRate int    `json:"refresh_rate" validate:"gte=5,lte=1440"` // minutes
}
```

## Error Messages

The validator automatically formats error messages for better UX:

```json
{
  "message": "Field 'Email' must be a valid email address; Field 'Password' must be at least 8 characters long"
}
```

## Adding Custom Validations

Edit `internal/shared/validator/validator.go` to register custom validators:

```go
func New() *CustomValidator {
    v := validator.New()

    // Register custom validation
    v.RegisterValidation("custom_tag", func(fl validator.FieldLevel) bool {
        // Your validation logic
        return true
    })

    return &CustomValidator{validator: v}
}
```

Then update `formatFieldError()` to handle the custom tag message formatting.

## References

- [go-playground/validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [All available validation tags](https://github.com/go-playground/validator#baked-in-validations)
