package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

const (
	userIDKey    = "user_id"
	userEmailKey = "user_email"
)

// AuthService defines the interface for authentication operations needed by middleware
type AuthService interface {
	GetUserByToken(ctx context.Context, accessToken string) (*UserSession, error)
	RefreshSession(ctx context.Context, refreshToken string) (*UserSession, error)
}

// SessionManager defines the interface for session cookie operations needed by middleware
type SessionManager interface {
	SetSessionCookies(c echo.Context, session *UserSession)
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
			// Create context from request once at the beginning
			ctx := c.Request().Context()

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
				session, refreshErr := service.RefreshSession(ctx, refreshToken)
				if refreshErr != nil {
					// Refresh failed, clear cookies and redirect to login
					logger.Debug("Token refresh failed, redirecting to login", "error", refreshErr)
					sessionMgr.ClearSessionCookies(c)
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Update access token cookie with refreshed token
				sessionMgr.UpdateAccessToken(c, session.AccessToken)

				// Set user data in context
				c.Set(userIDKey, session.UserID)
				c.Set(userEmailKey, session.Email)

				// Add token to request context for RLS
				ctxWithToken := database.ContextWithToken(ctx, session.AccessToken)
				c.SetRequest(c.Request().WithContext(ctxWithToken))

				logger.Debug("Session refreshed successfully", "user_id", session.UserID)
				return next(c)
			}

			// Validate access token and get user info
			session, err := service.GetUserByToken(ctx, accessToken)
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
				refreshedSession, refreshErr := service.RefreshSession(ctx, refreshToken)
				if refreshErr != nil {
					// Refresh failed, clear cookies and redirect
					logger.Debug("Token validation and refresh both failed, redirecting to login")
					sessionMgr.ClearSessionCookies(c)
					return c.Redirect(http.StatusFound, "/auth/login")
				}

				// Update access token cookie
				sessionMgr.UpdateAccessToken(c, refreshedSession.AccessToken)

				// Set user data in context
				c.Set(userIDKey, refreshedSession.UserID)
				c.Set(userEmailKey, refreshedSession.Email)

				// Add token to request context for RLS
				ctxWithToken := database.ContextWithToken(ctx, refreshedSession.AccessToken)
				c.SetRequest(c.Request().WithContext(ctxWithToken))

				logger.Debug("Session refreshed after access token validation failed", "user_id", refreshedSession.UserID)
				return next(c)
			}

			// Valid access token, set user data in context
			c.Set(userIDKey, session.UserID)
			c.Set(userEmailKey, session.Email)

			// Add token to request context for RLS
			ctxWithToken := database.ContextWithToken(ctx, accessToken)
			c.SetRequest(c.Request().WithContext(ctxWithToken))

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
