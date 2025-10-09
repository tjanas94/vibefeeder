package database

import (
	"github.com/labstack/echo/v4"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// dbContextKey is the key used to store the database client in the Echo context
	dbContextKey contextKey = "db"
)

// Middleware returns an Echo middleware that injects the database client into the request context
func Middleware(db *Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Store the database client in the Echo context
			c.Set(string(dbContextKey), db)
			return next(c)
		}
	}
}

// FromContext retrieves the database client from the Echo context
// Returns nil if the client is not found in the context
func FromContext(c echo.Context) *Client {
	db, ok := c.Get(string(dbContextKey)).(*Client)
	if !ok {
		return nil
	}
	return db
}

// MustFromContext retrieves the database client from the Echo context
// Panics if the client is not found - use this only when you're certain
// the middleware has been applied
func MustFromContext(c echo.Context) *Client {
	db := FromContext(c)
	if db == nil {
		panic("database client not found in context - ensure database middleware is applied")
	}
	return db
}
