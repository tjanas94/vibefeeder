package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"strings"

	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	// Password entropy requirement (bits)
	// 50 bits is approximately equivalent to 8 characters with 2 out of 3 character classes
	minPasswordEntropy = 50

	// Event types for logging
	eventUserRegistered    = "user_registered"
	eventUserLogin         = "user_login"
	eventUserPasswordReset = "user_password_reset"
)

// Service handles authentication business logic
type Service struct {
	authClient gotrue.Client
	repo       *Repository
	config     *config.AuthConfig
	logger     *slog.Logger
}

// NewService creates a new auth service
func NewService(supabaseURL, supabaseKey string, repo *Repository, cfg *config.AuthConfig, logger *slog.Logger) (*Service, error) {
	// Create gotrue client with custom URL
	// We use WithCustomGoTrueURL because we expect a full Supabase URL (e.g., http://127.0.0.1:54321 or https://xyz.supabase.co)
	// and need to append /auth/v1 to it
	authClient := gotrue.New("dummy", supabaseKey).WithCustomGoTrueURL(supabaseURL + "/auth/v1")

	if authClient == nil {
		return nil, fmt.Errorf("failed to create gotrue client")
	}

	return &Service{
		authClient: authClient,
		repo:       repo,
		config:     cfg,
		logger:     logger,
	}, nil
}

// ValidatePassword checks if password meets strength requirements
func (s *Service) ValidatePassword(password string) error {
	err := passwordvalidator.Validate(password, minPasswordEntropy)
	if err != nil {
		s.logger.Debug("Password validation failed", "error", err)
		return ErrWeakPassword
	}
	return nil
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req models.RegisterRequest) error {
	// Validate registration code if configured (constant-time comparison to prevent timing attacks)
	if s.config.RegistrationCode != "" {
		if subtle.ConstantTimeCompare([]byte(req.RegistrationCode), []byte(s.config.RegistrationCode)) != 1 {
			s.logger.Debug("Invalid registration code attempt", "email", req.Email)
			return ErrInvalidRegistrationCode
		}
	}

	// Validate password strength
	if err := s.ValidatePassword(req.Password); err != nil {
		return err
	}

	// Build redirect URL for email confirmation
	redirectURL := fmt.Sprintf("%s/auth/confirm", s.config.RedirectURL)

	// Create user with Supabase Auth
	resp, err := s.authClient.Signup(types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Data: map[string]interface{}{
			"redirect_to": redirectURL,
		},
	})

	if err != nil {
		// Check if user already exists
		if strings.Contains(err.Error(), "already registered") || strings.Contains(err.Error(), "email address already") {
			return ErrUserAlreadyExists
		}
		s.logger.Error("Failed to register user", "error", err, "email", req.Email)
		return fmt.Errorf("registration failed: %w", err)
	}

	// Log registration event
	userID := resp.ID.String()
	if err := s.repo.InsertEvent(ctx, database.PublicEventsInsert{
		UserId:    &userID,
		EventType: eventUserRegistered,
		Metadata:  nil,
	}); err != nil {
		s.logger.Error("Failed to log registration event", "error", err, "user_id", userID)
	}

	s.logger.Info("User registered successfully", "user_id", userID, "email", req.Email)
	return nil
}

// Login authenticates a user and returns session data
func (s *Service) Login(ctx context.Context, req models.LoginRequest) (*models.UserSession, error) {
	// Authenticate with Supabase using SignInWithEmailPassword
	resp, err := s.authClient.SignInWithEmailPassword(req.Email, req.Password)

	if err != nil {
		s.logger.Debug("Login failed", "error", err, "email", req.Email)
		return nil, ErrInvalidCredentials
	}

	// Log login event
	userID := resp.User.ID.String()
	if err := s.repo.InsertEvent(ctx, database.PublicEventsInsert{
		UserId:    &userID,
		EventType: eventUserLogin,
		Metadata:  nil,
	}); err != nil {
		s.logger.Error("Failed to log login event", "error", err, "user_id", userID)
	}

	s.logger.Info("User logged in successfully", "user_id", userID, "email", req.Email)

	return &models.UserSession{
		UserID:       userID,
		Email:        resp.User.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// RefreshSession refreshes an expired access token using refresh token
func (s *Service) RefreshSession(ctx context.Context, refreshToken string) (*models.UserSession, error) {
	resp, err := s.authClient.RefreshToken(refreshToken)
	if err != nil {
		s.logger.Debug("Token refresh failed", "error", err)
		return nil, ErrSessionExpired
	}

	userID := resp.User.ID.String()
	s.logger.Debug("Session refreshed successfully", "user_id", userID)

	return &models.UserSession{
		UserID:       userID,
		Email:        resp.User.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// GetUserByToken retrieves user information from an access token
func (s *Service) GetUserByToken(ctx context.Context, accessToken string) (*models.UserSession, error) {
	// Create a client with the access token
	client := s.authClient.WithToken(accessToken)

	resp, err := client.GetUser()
	if err != nil {
		s.logger.Debug("Failed to get user by token", "error", err)
		return nil, ErrSessionExpired
	}

	return &models.UserSession{
		UserID:      resp.ID.String(),
		Email:       resp.Email,
		AccessToken: accessToken,
	}, nil
}

// Logout signs out the user and revokes the session
func (s *Service) Logout(ctx context.Context, accessToken string) error {
	// Create a client with the access token
	client := s.authClient.WithToken(accessToken)

	err := client.Logout()
	if err != nil {
		s.logger.Error("Logout failed", "error", err)
		return fmt.Errorf("logout failed: %w", err)
	}

	s.logger.Debug("User logged out successfully")
	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(ctx context.Context, email string) error {
	// Note: Redirect URL must be configured in Supabase dashboard under Authentication > URL Configuration
	// Set "Redirect URLs" to include: http://localhost:8080/auth/reset-password (dev) and your production URL
	err := s.authClient.Recover(types.RecoverRequest{
		Email: email,
	})

	if err != nil {
		s.logger.Error("Failed to send password reset email", "error", err, "email", email)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	s.logger.Info("Password reset email sent", "email", email)
	return nil
}

// VerifyRecoveryToken verifies a password recovery token and returns a session
func (s *Service) VerifyRecoveryToken(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	// Build redirect URL
	redirectURL := fmt.Sprintf("%s/auth/reset-password", s.config.RedirectURL)

	// Verify the token using Supabase Auth's verify endpoint
	resp, err := s.authClient.Verify(types.VerifyRequest{
		Type:       "recovery",
		Token:      tokenHash,
		RedirectTo: redirectURL,
	})

	if err != nil {
		s.logger.Error("Failed to verify recovery token", "error", err)
		return nil, ErrInvalidToken
	}

	// Check if verification was successful
	if resp.Error != "" {
		s.logger.Error("Token verification failed", "error", resp.Error, "error_code", resp.ErrorCode)
		return nil, ErrInvalidToken
	}

	// Get user information using the access token
	client := s.authClient.WithToken(resp.AccessToken)
	user, err := client.GetUser()
	if err != nil {
		s.logger.Error("Failed to get user after token verification", "error", err)
		return nil, ErrInvalidToken
	}

	userID := user.ID.String()
	s.logger.Info("Recovery token verified successfully", "user_id", userID)

	return &models.UserSession{
		UserID:       userID,
		Email:        user.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// ResetPassword resets the user's password using an authenticated session
func (s *Service) ResetPassword(ctx context.Context, accessToken, newPassword string) error {
	// Validate password strength
	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Create a client with the access token from the recovery session
	client := s.authClient.WithToken(accessToken)

	// Update password with Supabase
	resp, err := client.UpdateUser(types.UpdateUserRequest{
		Password: &newPassword,
	})

	if err != nil {
		s.logger.Error("Failed to reset password", "error", err)
		return ErrInvalidToken
	}

	// Log password reset event
	userID := resp.ID.String()
	if err := s.repo.InsertEvent(ctx, database.PublicEventsInsert{
		UserId:    &userID,
		EventType: eventUserPasswordReset,
		Metadata:  nil,
	}); err != nil {
		s.logger.Error("Failed to log password reset event", "error", err, "user_id", userID)
	}

	s.logger.Info("Password reset successfully", "user_id", userID)
	return nil
}
