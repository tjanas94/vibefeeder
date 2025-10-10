package auth

import (
	"github.com/labstack/echo/v4"
)

const mockUserIDKey = "user_id"

// MockAuthMiddleware injects a mock user ID into the context
// TODO: Replace with real authentication when auth is implemented
func MockAuthMiddleware(userID string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(mockUserIDKey, userID)
			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(mockUserIDKey).(string); ok {
		return userID
	}
	return ""
}
