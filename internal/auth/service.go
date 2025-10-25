package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/supabase-community/gotrue-go/types"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	sharedAuth "github.com/tjanas94/vibefeeder/internal/shared/auth"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
	"github.com/tjanas94/vibefeeder/internal/shared/events"
)

// Service handles authentication business logic
type Service struct {
	authClient sharedAuth.GoTrueAdapter
	eventRepo  events.EventRepository
	config     *config.AuthConfig
	logger     *slog.Logger
}

// NewService creates a new auth service
func NewService(authClient sharedAuth.GoTrueAdapter, eventRepo events.EventRepository, cfg *config.AuthConfig, logger *slog.Logger) *Service {
	return &Service{
		authClient: authClient,
		eventRepo:  eventRepo,
		config:     cfg,
		logger:     logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req models.RegisterRequest) error {
	// Validate registration code if configured (constant-time comparison to prevent timing attacks)
	if !validateRegistrationCode(req.RegistrationCode, s.config.RegistrationCode) {
		s.logger.Debug("Invalid registration code attempt", "email", req.Email)
		return NewInvalidRegistrationCodeError()
	}

	// Build redirect URL for email confirmation
	redirectURL := fmt.Sprintf("%s/auth/confirm", s.config.RedirectURL)

	// Create user with Supabase Auth
	resp, err := s.authClient.Signup(types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Data: map[string]any{
			"redirect_to": redirectURL,
		},
	})

	if err != nil {
		// Check if user already exists
		if isUserExistsError(err) {
			s.logger.Debug("User registration attempt with existing email", "email", req.Email)
			return NewUserAlreadyExistsError()
		}
		return fmt.Errorf("registration failed: %w", err)
	}

	// Log registration event
	userID := resp.ID.String()
	s.recordUserEvent(ctx, userID, events.EventUserRegistered)

	s.logger.Info("User registered successfully", "user_id", userID, "email", req.Email)
	return nil
}

// Login authenticates a user and returns session data
func (s *Service) Login(ctx context.Context, req models.LoginRequest) (*sharedAuth.UserSession, error) {
	// Authenticate with Supabase using SignInWithEmailPassword
	resp, err := s.authClient.SignInWithEmailPassword(req.Email, req.Password)

	if err != nil {
		if isInvalidCredentialsError(err) {
			s.logger.Debug("Login failed with invalid credentials", "email", req.Email)
			return nil, NewInvalidCredentialsError()
		}
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Log login event
	userID := resp.User.ID.String()
	s.recordUserEvent(ctx, userID, events.EventUserLogin)

	s.logger.Info("User logged in successfully", "user_id", userID, "email", req.Email)

	return &sharedAuth.UserSession{
		UserID:       userID,
		Email:        resp.User.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// RefreshSession refreshes an expired access token using refresh token
func (s *Service) RefreshSession(ctx context.Context, refreshToken string) (*sharedAuth.UserSession, error) {
	resp, err := s.authClient.RefreshToken(refreshToken)
	if err != nil {
		s.logger.Debug("Token refresh failed", "error", err)
		return nil, NewSessionExpiredError()
	}

	userID := resp.User.ID.String()
	s.logger.Debug("Session refreshed successfully", "user_id", userID)

	return &sharedAuth.UserSession{
		UserID:       userID,
		Email:        resp.User.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// GetUserByToken retrieves user information from an access token
func (s *Service) GetUserByToken(ctx context.Context, accessToken string) (*sharedAuth.UserSession, error) {
	// Create a client with the access token
	client := s.authClient.WithToken(accessToken)

	resp, err := client.GetUser()
	if err != nil {
		s.logger.Debug("Failed to get user by token", "error", err)
		return nil, NewSessionExpiredError()
	}

	return &sharedAuth.UserSession{
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

// VerifyEmailConfirmation verifies an email confirmation token from signup
func (s *Service) VerifyEmailConfirmation(ctx context.Context, tokenHash string) error {
	// Build redirect URL
	redirectURL := fmt.Sprintf("%s/auth/confirm", s.config.RedirectURL)

	// Verify the token using Supabase Auth's verify endpoint
	resp, err := s.authClient.Verify(types.VerifyRequest{
		Type:       "signup",
		Token:      tokenHash,
		RedirectTo: redirectURL,
	})

	if err != nil {
		s.logger.Error("Failed to verify email confirmation token", "error", err)
		return NewInvalidTokenError()
	}

	// Check if verification was successful
	if resp.Error != "" {
		s.logger.Error("Email confirmation failed", "error", resp.Error, "error_code", resp.ErrorCode)
		return NewInvalidTokenError()
	}

	// Get user information using the access token
	client := s.authClient.WithToken(resp.AccessToken)
	user, err := client.GetUser()
	if err != nil {
		s.logger.Error("Failed to get user after email confirmation", "error", err)
		return NewInvalidTokenError()
	}

	userID := user.ID.String()
	s.logger.Info("Email confirmed successfully", "user_id", userID, "email", user.Email)

	// Log email confirmation event
	s.recordUserEvent(ctx, userID, events.EventUserEmailConfirmed)

	return nil
}

// ResetPassword resets the user's password using a recovery token
func (s *Service) ResetPassword(ctx context.Context, tokenHash, newPassword string) error {
	// Verify recovery token and get temporary access token
	session, err := s.verifyRecoveryToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Create a client with the access token from the recovery session
	client := s.authClient.WithToken(session.AccessToken)

	// Update password with Supabase
	resp, err := client.UpdateUser(types.UpdateUserRequest{
		Password: &newPassword,
	})

	if err != nil {
		if isSamePasswordError(err) {
			s.logger.Debug("User attempted to reset password to the same value", "user_id", session.UserID)
			return NewSamePasswordError()
		}
		if isInvalidTokenError(err) {
			s.logger.Debug("Token error during password reset", "user_id", session.UserID)
			return NewInvalidTokenError()
		}
		return fmt.Errorf("reset password failed: %w", err)
	}

	// Log password reset event
	userID := resp.ID.String()
	s.recordUserEvent(ctx, userID, events.EventUserPasswordReset)

	s.logger.Info("Password reset successfully", "user_id", userID)
	return nil
}

// verifyRecoveryToken verifies a password recovery token and returns a session
func (s *Service) verifyRecoveryToken(ctx context.Context, tokenHash string) (*sharedAuth.UserSession, error) {
	// Build redirect URL
	redirectURL := fmt.Sprintf("%s/auth/reset-password", s.config.RedirectURL)

	// Verify the token using Supabase Auth's verify endpoint
	resp, err := s.authClient.Verify(types.VerifyRequest{
		Type:       "recovery",
		Token:      tokenHash,
		RedirectTo: redirectURL,
	})

	if err != nil {
		if isInvalidTokenError(err) {
			s.logger.Debug("Invalid recovery token", "error", err)
			return nil, NewInvalidTokenError()
		}
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// Check if verification was successful
	if resp.Error != "" {
		s.logger.Debug("Token verification failed", "error", resp.Error, "error_code", resp.ErrorCode)
		return nil, NewInvalidTokenError()
	}

	// Get user information using the access token
	client := s.authClient.WithToken(resp.AccessToken)
	user, err := client.GetUser()
	if err != nil {
		if isInvalidTokenError(err) {
			s.logger.Debug("Invalid token during user retrieval", "error", err)
			return nil, NewInvalidTokenError()
		}
		return nil, fmt.Errorf("failed to get user after token verification: %w", err)
	}

	userID := user.ID.String()
	s.logger.Info("Recovery token verified successfully", "user_id", userID)

	return &sharedAuth.UserSession{
		UserID:       userID,
		Email:        user.Email,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// recordUserEvent records an event for a user with the specified event type
func (s *Service) recordUserEvent(ctx context.Context, userID string, eventType string) {
	if err := s.eventRepo.RecordEvent(ctx, database.PublicEventsInsert{
		UserId:    &userID,
		EventType: eventType,
		Metadata:  nil,
	}); err != nil {
		s.logger.Error("Failed to log event", "event_type", eventType, "error", err, "user_id", userID)
	}
}
