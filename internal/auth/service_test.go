package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// MockAuth embeds a real gotrue.Client and overrides methods for testing
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) Signup(req types.SignupRequest) (*types.SignupResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.SignupResponse), args.Error(1)
}

func (m *MockAuth) SignInWithEmailPassword(email, password string) (*types.TokenResponse, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TokenResponse), args.Error(1)
}

func (m *MockAuth) RefreshToken(token string) (*types.TokenResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TokenResponse), args.Error(1)
}

func (m *MockAuth) Verify(req types.VerifyRequest) (*types.VerifyResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.VerifyResponse), args.Error(1)
}

func (m *MockAuth) Recover(req types.RecoverRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockAuth) WithToken(token string) gotrue.Client {
	args := m.Called(token)
	return args.Get(0).(gotrue.Client)
}

func (m *MockAuth) GetUser() (*types.UserResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockAuth) UpdateUser(req types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.UpdateUserResponse), args.Error(1)
}

func (m *MockAuth) Logout() error {
	args := m.Called()
	return args.Error(0)
}

// Implement all remaining gotrue.Client methods
func (m *MockAuth) WithCustomGoTrueURL(url string) gotrue.Client { return m }
func (m *MockAuth) WithClient(client http.Client) gotrue.Client  { return m }
func (m *MockAuth) AdminAudit(req types.AdminAuditRequest) (*types.AdminAuditResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminGenerateLink(req types.AdminGenerateLinkRequest) (*types.AdminGenerateLinkResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminListSSOProviders() (*types.AdminListSSOProvidersResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminCreateSSOProvider(req types.AdminCreateSSOProviderRequest) (*types.AdminCreateSSOProviderResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminGetSSOProvider(req types.AdminGetSSOProviderRequest) (*types.AdminGetSSOProviderResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminUpdateSSOProvider(req types.AdminUpdateSSOProviderRequest) (*types.AdminUpdateSSOProviderResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminDeleteSSOProvider(req types.AdminDeleteSSOProviderRequest) (*types.AdminDeleteSSOProviderResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminCreateUser(req types.AdminCreateUserRequest) (*types.AdminCreateUserResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminListUsers() (*types.AdminListUsersResponse, error) { return nil, nil }
func (m *MockAuth) AdminGetUser(req types.AdminGetUserRequest) (*types.AdminGetUserResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminUpdateUser(req types.AdminUpdateUserRequest) (*types.AdminUpdateUserResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminDeleteUser(req types.AdminDeleteUserRequest) error { return nil }
func (m *MockAuth) AdminListUserFactors(req types.AdminListUserFactorsRequest) (*types.AdminListUserFactorsResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminUpdateUserFactor(req types.AdminUpdateUserFactorRequest) (*types.AdminUpdateUserFactorResponse, error) {
	return nil, nil
}
func (m *MockAuth) AdminDeleteUserFactor(req types.AdminDeleteUserFactorRequest) error { return nil }
func (m *MockAuth) Authorize(req types.AuthorizeRequest) (*types.AuthorizeResponse, error) {
	return nil, nil
}
func (m *MockAuth) EnrollFactor(req types.EnrollFactorRequest) (*types.EnrollFactorResponse, error) {
	return nil, nil
}
func (m *MockAuth) ChallengeFactor(req types.ChallengeFactorRequest) (*types.ChallengeFactorResponse, error) {
	return nil, nil
}
func (m *MockAuth) VerifyFactor(req types.VerifyFactorRequest) (*types.VerifyFactorResponse, error) {
	return nil, nil
}
func (m *MockAuth) UnenrollFactor(req types.UnenrollFactorRequest) (*types.UnenrollFactorResponse, error) {
	return nil, nil
}
func (m *MockAuth) HealthCheck() (*types.HealthCheckResponse, error)              { return nil, nil }
func (m *MockAuth) Invite(req types.InviteRequest) (*types.InviteResponse, error) { return nil, nil }
func (m *MockAuth) Magiclink(req types.MagiclinkRequest) error                    { return nil }
func (m *MockAuth) OTP(req types.OTPRequest) error                                { return nil }
func (m *MockAuth) Reauthenticate() error                                         { return nil }
func (m *MockAuth) GetSettings() (*types.SettingsResponse, error)                 { return nil, nil }
func (m *MockAuth) SignInWithPhonePassword(phone, password string) (*types.TokenResponse, error) {
	return nil, nil
}
func (m *MockAuth) Token(req types.TokenRequest) (*types.TokenResponse, error) { return nil, nil }
func (m *MockAuth) VerifyForUser(req types.VerifyForUserRequest) (*types.VerifyForUserResponse, error) {
	return nil, nil
}
func (m *MockAuth) SAMLACS(req *http.Request) (*http.Response, error)    { return nil, nil }
func (m *MockAuth) SAMLMetadata() ([]byte, error)                        { return nil, nil }
func (m *MockAuth) SSO(req types.SSORequest) (*types.SSOResponse, error) { return nil, nil }

type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) RecordEvent(ctx context.Context, event database.PublicEventsInsert) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func newTestLogger() *slog.Logger {
	// Use io.Discard to avoid nil handler issues
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestConfig() *config.AuthConfig {
	return &config.AuthConfig{
		RedirectURL:      "http://localhost:8080",
		RegistrationCode: "test-code-123",
	}
}

// TestRegister_Success tests successful user registration
func TestRegister_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testEmail := "user@example.com"

	mockAuth.On("Signup", mock.MatchedBy(func(req types.SignupRequest) bool {
		return req.Email == testEmail
	})).Return(&types.SignupResponse{
		User: types.User{Email: testEmail},
		Session: types.Session{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
	}, nil)

	mockEvents.On("RecordEvent", ctx, mock.Anything).Return(nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.Register(ctx, models.RegisterRequest{
		Email:            testEmail,
		Password:         "SecurePass123!@#",
		PasswordConfirm:  "SecurePass123!@#",
		RegistrationCode: "test-code-123",
	})

	assert.NoError(t, err)
	mockAuth.AssertCalled(t, "Signup", mock.Anything)
}

// TestRegister_InvalidCode tests registration with invalid code
func TestRegister_InvalidCode(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.Register(ctx, models.RegisterRequest{
		Email:            "user@example.com",
		Password:         "SecurePass123!@#",
		PasswordConfirm:  "SecurePass123!@#",
		RegistrationCode: "wrong-code",
	})

	assert.Equal(t, ErrInvalidRegistrationCode, err)
	mockAuth.AssertNotCalled(t, "Signup")
}

// TestRegister_UserExists tests registration with existing email
func TestRegister_UserExists(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	mockAuth.On("Signup", mock.Anything).Return(nil, errors.New("user_already_exists"))

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.Register(ctx, models.RegisterRequest{
		Email:            "user@example.com",
		Password:         "SecurePass123!@#",
		PasswordConfirm:  "SecurePass123!@#",
		RegistrationCode: "test-code-123",
	})

	assert.Equal(t, ErrUserAlreadyExists, err)
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testEmail := "user@example.com"
	testAccessToken := "access-token-123"
	testRefreshToken := "refresh-token-456"

	mockAuth.On("SignInWithEmailPassword", testEmail, "password123").Return(&types.TokenResponse{
		Session: types.Session{
			AccessToken:  testAccessToken,
			RefreshToken: testRefreshToken,
			User: types.User{
				Email: testEmail,
			},
		},
	}, nil)

	mockEvents.On("RecordEvent", ctx, mock.Anything).Return(nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.Login(ctx, models.LoginRequest{
		Email:    testEmail,
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, testEmail, session.Email)
	assert.Equal(t, testAccessToken, session.AccessToken)
}

// TestLogin_InvalidCredentials tests login with wrong password
func TestLogin_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	mockAuth.On("SignInWithEmailPassword", "user@example.com", "wrong").
		Return(nil, errors.New("Invalid login credentials"))

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.Login(ctx, models.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, session)
}

// TestRefreshSession_Success tests token refresh
func TestRefreshSession_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testRefreshToken := "refresh-token-456"
	testNewAccessToken := "new-access-token"

	mockAuth.On("RefreshToken", testRefreshToken).Return(&types.TokenResponse{
		Session: types.Session{
			AccessToken:  testNewAccessToken,
			RefreshToken: testRefreshToken,
			User: types.User{
				Email: "user@example.com",
			},
		},
	}, nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.RefreshSession(ctx, testRefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, testNewAccessToken, session.AccessToken)
}

// TestRefreshSession_Expired tests refresh with expired token
func TestRefreshSession_Expired(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	mockAuth.On("RefreshToken", "expired-token").Return(nil, errors.New("token_expired"))

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.RefreshSession(ctx, "expired-token")

	assert.Equal(t, ErrSessionExpired, err)
	assert.Nil(t, session)
}

// TestGetUserByToken_Success tests retrieving user from token
func TestGetUserByToken_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testAccessToken := "access-token-123"
	testEmail := "user@example.com"

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("GetUser").Return(&types.UserResponse{
		User: types.User{
			Email: testEmail,
		},
	}, nil)

	mockAuth.On("WithToken", testAccessToken).Return(mockClientWithToken)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.GetUserByToken(ctx, testAccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, testEmail, session.Email)
}

// TestGetUserByToken_Invalid tests invalid token handling
func TestGetUserByToken_Invalid(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("GetUser").Return(nil, errors.New("invalid token"))

	mockAuth.On("WithToken", "invalid-token").Return(mockClientWithToken)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	session, err := svc.GetUserByToken(ctx, "invalid-token")

	assert.Equal(t, ErrSessionExpired, err)
	assert.Nil(t, session)
}

// TestLogout_Success tests successful logout
func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testAccessToken := "access-token-123"

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("Logout").Return(nil)

	mockAuth.On("WithToken", testAccessToken).Return(mockClientWithToken)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.Logout(ctx, testAccessToken)

	assert.NoError(t, err)
}

// TestSendPasswordResetEmail_Success tests sending reset email
func TestSendPasswordResetEmail_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testEmail := "user@example.com"

	mockAuth.On("Recover", mock.MatchedBy(func(req types.RecoverRequest) bool {
		return req.Email == testEmail
	})).Return(nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.SendPasswordResetEmail(ctx, testEmail)

	assert.NoError(t, err)
}

// TestVerifyEmailConfirmation_Success tests email confirmation
func TestVerifyEmailConfirmation_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testToken := "email-confirmation-token"
	testAccessToken := "access-token-123"
	testEmail := "user@example.com"

	mockAuth.On("Verify", mock.MatchedBy(func(req types.VerifyRequest) bool {
		return req.Type == "signup" && req.Token == testToken
	})).Return(&types.VerifyResponse{
		AccessToken:  testAccessToken,
		RefreshToken: "refresh-token",
	}, nil)

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("GetUser").Return(&types.UserResponse{
		User: types.User{
			Email: testEmail,
		},
	}, nil)

	mockAuth.On("WithToken", testAccessToken).Return(mockClientWithToken)
	mockEvents.On("RecordEvent", ctx, mock.Anything).Return(nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.VerifyEmailConfirmation(ctx, testToken)

	assert.NoError(t, err)
	mockEvents.AssertCalled(t, "RecordEvent", ctx, mock.Anything)
}

// TestVerifyEmailConfirmation_InvalidToken tests invalid email token
func TestVerifyEmailConfirmation_InvalidToken(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	mockAuth.On("Verify", mock.Anything).Return(nil, errors.New("invalid token"))

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.VerifyEmailConfirmation(ctx, "invalid-token")

	assert.Equal(t, ErrInvalidToken, err)
}

// TestResetPassword_Success tests successful password reset
func TestResetPassword_Success(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testToken := "recovery-token"
	testNewPassword := "NewSecurePass123!@#"
	testAccessToken := "access-token-from-recovery"
	testEmail := "user@example.com"

	mockAuth.On("Verify", mock.MatchedBy(func(req types.VerifyRequest) bool {
		return req.Type == "recovery" && req.Token == testToken
	})).Return(&types.VerifyResponse{
		AccessToken:  testAccessToken,
		RefreshToken: "refresh-token",
	}, nil)

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("GetUser").Return(&types.UserResponse{
		User: types.User{
			Email: testEmail,
		},
	}, nil)

	mockClientWithToken.On("UpdateUser", mock.MatchedBy(func(req types.UpdateUserRequest) bool {
		return req.Password != nil && *req.Password == testNewPassword
	})).Return(&types.UpdateUserResponse{
		User: types.User{
			Email: testEmail,
		},
	}, nil)

	mockAuth.On("WithToken", testAccessToken).Return(mockClientWithToken)
	mockEvents.On("RecordEvent", ctx, mock.Anything).Return(nil)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.ResetPassword(ctx, testToken, testNewPassword)

	assert.NoError(t, err)
	mockEvents.AssertCalled(t, "RecordEvent", ctx, mock.Anything)
}

// TestResetPassword_SamePassword tests same password rejection
func TestResetPassword_SamePassword(t *testing.T) {
	ctx := context.Background()
	mockAuth := new(MockAuth)
	mockEvents := new(MockEventRepo)
	cfg := newTestConfig()

	testToken := "recovery-token"
	testPassword := "SamePassword123!@#"
	testAccessToken := "access-token-from-recovery"
	testEmail := "user@example.com"

	mockAuth.On("Verify", mock.Anything).Return(&types.VerifyResponse{
		AccessToken:  testAccessToken,
		RefreshToken: "refresh-token",
	}, nil)

	mockClientWithToken := new(MockAuth)
	mockClientWithToken.On("GetUser").Return(&types.UserResponse{
		User: types.User{
			Email: testEmail,
		},
	}, nil)

	mockClientWithToken.On("UpdateUser", mock.Anything).Return(nil, errors.New("same_password"))

	mockAuth.On("WithToken", testAccessToken).Return(mockClientWithToken)

	svc := NewService(mockAuth, mockEvents, cfg, newTestLogger())
	err := svc.ResetPassword(ctx, testToken, testPassword)

	assert.Equal(t, ErrSamePassword, err)
}
