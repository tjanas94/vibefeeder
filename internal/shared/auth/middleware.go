package auth

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	userIDKey    = "user_id"
	userEmailKey = "user_email"
)

// AuthService defines the interface for authentication operations needed by middleware
type AuthService interface {
	GetUserByToken(ctx echo.Context, accessToken string) (userID string, email string, err error)
	RefreshSession(ctx echo.Context, refreshToken string) (accessToken string, userID string, email string, err error)
}

// SessionManager defines the interface for session cookie operations needed by middleware
type SessionManager interface {
	GetAccessToken(c echo.Context) (string, error)
	GetRefreshToken(c echo.Context) (string, error)
	UpdateAccessToken(c echo.Context, accessToken string)
	ClearSessionCookies(c echo.Context)
}

// AuthMiddleware creates middleware that requires authentication
// It checks for valid session cookies and automatically refreshes expired tokens
func AuthMiddleware(service AuthService, sessionMgr SessionManager, logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to get access token from cookie
			accessToken, err := sessionMgr.GetAccessToken(c)

			// If no access token, try to refresh using refresh token
			if err != nil || accessToken == "" {
				refreshToken, refreshErr := sessionMgr.GetRefreshToken(c)
				if refreshErr != nil || refreshToken == "" {
					// No valid session, redirect to login
					logger.Debug("No valid session found, redirecting to login")
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Attempt to refresh the session
				newAccessToken, userID, email, refreshErr := service.RefreshSession(c, refreshToken)
				if refreshErr != nil {
					// Refresh failed, clear cookies and redirect to login
					logger.Debug("Token refresh failed, redirecting to login", "error", refreshErr)
					sessionMgr.ClearSessionCookies(c)
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Update access token cookie with refreshed token
				sessionMgr.UpdateAccessToken(c, newAccessToken)

				// Set user data in context
				c.Set(userIDKey, userID)
				c.Set(userEmailKey, email)

				logger.Debug("Session refreshed successfully", "user_id", userID)
				return next(c)
			}

			// Validate access token and get user info
			userID, email, err := service.GetUserByToken(c, accessToken)
			if err != nil {
				// Try to refresh using refresh token
				refreshToken, refreshErr := sessionMgr.GetRefreshToken(c)
				if refreshErr != nil || refreshToken == "" {
					// No refresh token, clear cookies and redirect
					logger.Debug("Access token invalid and no refresh token, redirecting to login")
					sessionMgr.ClearSessionCookies(c)
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Attempt to refresh
				newAccessToken, userID, email, refreshErr := service.RefreshSession(c, refreshToken)
				if refreshErr != nil {
					// Refresh failed, clear cookies and redirect
					logger.Debug("Token validation and refresh both failed, redirecting to login")
					sessionMgr.ClearSessionCookies(c)
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Update access token cookie
				sessionMgr.UpdateAccessToken(c, newAccessToken)

				// Set user data in context
				c.Set(userIDKey, userID)
				c.Set(userEmailKey, email)

				logger.Debug("Session refreshed after access token validation failed", "user_id", userID)
				return next(c)
			}

			// Valid access token, set user data in context
			c.Set(userIDKey, userID)
			c.Set(userEmailKey, email)

			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserEmail retrieves the user email from the context
func GetUserEmail(c echo.Context) string {
	if email, ok := c.Get(userEmailKey).(string); ok {
		return email
	}
	return ""
}
