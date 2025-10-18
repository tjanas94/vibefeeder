package auth

import (
	"github.com/labstack/echo/v4"
)

// MiddlewareAdapter adapts the auth Service to implement the middleware interfaces
type MiddlewareAdapter struct {
	service *Service
}

// NewMiddlewareAdapter creates a new middleware adapter
func NewMiddlewareAdapter(service *Service) *MiddlewareAdapter {
	return &MiddlewareAdapter{service: service}
}

// GetUserByToken implements the AuthService interface for middleware
func (a *MiddlewareAdapter) GetUserByToken(ctx echo.Context, accessToken string) (string, string, error) {
	session, err := a.service.GetUserByToken(ctx.Request().Context(), accessToken)
	if err != nil {
		return "", "", err
	}
	return session.UserID, session.Email, nil
}

// RefreshSession implements the AuthService interface for middleware
func (a *MiddlewareAdapter) RefreshSession(ctx echo.Context, refreshToken string) (string, string, string, error) {
	session, err := a.service.RefreshSession(ctx.Request().Context(), refreshToken)
	if err != nil {
		return "", "", "", err
	}
	return session.AccessToken, session.UserID, session.Email, nil
}
