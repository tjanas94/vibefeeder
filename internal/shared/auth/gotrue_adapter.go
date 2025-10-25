package auth

import (
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
)

// GoTrueAdapter defines the interface for GoTrue client operations.
// This abstraction allows for easier testing and potential future client implementations.
type GoTrueAdapter interface {
	// Signup creates a new user account.
	Signup(req types.SignupRequest) (*types.SignupResponse, error)

	// SignInWithEmailPassword authenticates a user with email and password.
	SignInWithEmailPassword(email, password string) (*types.TokenResponse, error)

	// RefreshToken refreshes an expired access token using a refresh token.
	RefreshToken(refreshToken string) (*types.TokenResponse, error)

	// GetUser retrieves user information using the current access token.
	GetUser() (*types.UserResponse, error)

	// UpdateUser updates user information.
	UpdateUser(req types.UpdateUserRequest) (*types.UpdateUserResponse, error)

	// Logout signs out the user and revokes the session.
	Logout() error

	// Recover sends a password recovery email.
	Recover(req types.RecoverRequest) error

	// Verify verifies an email confirmation or recovery token.
	Verify(req types.VerifyRequest) (*types.VerifyResponse, error)

	// WithToken creates a new client instance with the specified access token.
	WithToken(accessToken string) GoTrueAdapter
}

// GoTrueClientAdapter is a concrete implementation of GoTrueAdapter that wraps gotrue.Client.
type GoTrueClientAdapter struct {
	client gotrue.Client
}

// NewGoTrueClientAdapter creates a new adapter wrapping the provided GoTrue client.
func NewGoTrueClientAdapter(client gotrue.Client) *GoTrueClientAdapter {
	return &GoTrueClientAdapter{
		client: client,
	}
}

// Signup creates a new user account.
func (a *GoTrueClientAdapter) Signup(req types.SignupRequest) (*types.SignupResponse, error) {
	return a.client.Signup(req)
}

// SignInWithEmailPassword authenticates a user with email and password.
func (a *GoTrueClientAdapter) SignInWithEmailPassword(email, password string) (*types.TokenResponse, error) {
	return a.client.SignInWithEmailPassword(email, password)
}

// RefreshToken refreshes an expired access token using a refresh token.
func (a *GoTrueClientAdapter) RefreshToken(refreshToken string) (*types.TokenResponse, error) {
	return a.client.RefreshToken(refreshToken)
}

// GetUser retrieves user information using the current access token.
func (a *GoTrueClientAdapter) GetUser() (*types.UserResponse, error) {
	return a.client.GetUser()
}

// UpdateUser updates user information.
func (a *GoTrueClientAdapter) UpdateUser(req types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	return a.client.UpdateUser(req)
}

// Logout signs out the user and revokes the session.
func (a *GoTrueClientAdapter) Logout() error {
	return a.client.Logout()
}

// Recover sends a password recovery email.
func (a *GoTrueClientAdapter) Recover(req types.RecoverRequest) error {
	return a.client.Recover(req)
}

// Verify verifies an email confirmation or recovery token.
func (a *GoTrueClientAdapter) Verify(req types.VerifyRequest) (*types.VerifyResponse, error) {
	return a.client.Verify(req)
}

// WithToken creates a new adapter instance with the specified access token.
func (a *GoTrueClientAdapter) WithToken(accessToken string) GoTrueAdapter {
	return &GoTrueClientAdapter{
		client: a.client.WithToken(accessToken),
	}
}
