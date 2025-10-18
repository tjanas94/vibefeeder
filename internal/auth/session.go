package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

// SessionManager handles session cookie operations
type SessionManager struct {
	config *config.AuthConfig
}

// NewSessionManager creates a new session manager
func NewSessionManager(cfg *config.AuthConfig) *SessionManager {
	return &SessionManager{config: cfg}
}

// SetSessionCookies sets both access and refresh token cookies
func (sm *SessionManager) SetSessionCookies(c echo.Context, session *models.UserSession) {
	// Set access token cookie
	c.SetCookie(&http.Cookie{
		Name:     sm.config.SessionCookieName,
		Value:    session.AccessToken,
		Path:     "/",
		MaxAge:   int(sm.config.AccessTokenMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   sm.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     sm.config.RefreshCookieName,
		Value:    session.RefreshToken,
		Path:     "/",
		MaxAge:   int(sm.config.RefreshTokenMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   sm.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookies removes session cookies
func (sm *SessionManager) ClearSessionCookies(c echo.Context) {
	// Delete access token cookie
	c.SetCookie(&http.Cookie{
		Name:     sm.config.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sm.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Delete refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     sm.config.RefreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sm.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetAccessToken retrieves the access token from cookie
func (sm *SessionManager) GetAccessToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie(sm.config.SessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRefreshToken retrieves the refresh token from cookie
func (sm *SessionManager) GetRefreshToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie(sm.config.RefreshCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// UpdateAccessToken updates only the access token cookie (used after refresh)
func (sm *SessionManager) UpdateAccessToken(c echo.Context, accessToken string) {
	c.SetCookie(&http.Cookie{
		Name:     sm.config.SessionCookieName,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(sm.config.AccessTokenMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   sm.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
