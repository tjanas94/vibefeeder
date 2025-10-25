package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService mocks the AuthService interface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetUserByToken(ctx context.Context, accessToken string) (*UserSession, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserSession), args.Error(1)
}

func (m *MockAuthService) RefreshSession(ctx context.Context, refreshToken string) (*UserSession, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserSession), args.Error(1)
}

// MockSessionManager mocks the SessionManager interface
type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) SetSessionCookies(c echo.Context, session *UserSession) {
	m.Called(c, session)
}

func (m *MockSessionManager) GetAccessToken(c echo.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

func (m *MockSessionManager) GetRefreshToken(c echo.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

func (m *MockSessionManager) UpdateAccessToken(c echo.Context, accessToken string) {
	m.Called(c, accessToken)
}

func (m *MockSessionManager) ClearSessionCookies(c echo.Context) {
	m.Called(c)
}

// Helper to create a test logger that discards output
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create Echo context for testing
func newTestEchoContext(req *http.Request) echo.Context {
	e := echo.New()
	return e.NewContext(req, httptest.NewRecorder())
}

// TestAuthMiddleware_ValidAccessToken tests successful authentication with valid access token
func TestAuthMiddleware_ValidAccessToken(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testUserID := "user-123"
	testEmail := "user@example.com"
	testAccessToken := "valid-access-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return(testAccessToken, nil)
	mockService.On("GetUserByToken", mock.Anything, testAccessToken).Return(&UserSession{
		UserID:       testUserID,
		Email:        testEmail,
		AccessToken:  testAccessToken,
		RefreshToken: "refresh-token",
	}, nil)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	c := newTestEchoContext(req)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, testUserID, c.Get(userIDKey), "user ID should be set in context")
	assert.Equal(t, testEmail, c.Get(userEmailKey), "email should be set in context")
	mockService.AssertCalled(t, "GetUserByToken", mock.Anything, testAccessToken)
	mockSessionMgr.AssertNotCalled(t, "GetRefreshToken")
	mockSessionMgr.AssertNotCalled(t, "ClearSessionCookies")
}

// TestAuthMiddleware_NoAccessTokenNoRefreshToken tests redirect when no tokens available
func TestAuthMiddleware_NoAccessTokenNoRefreshToken(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return("", errors.New("no access token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return("", errors.New("no refresh token"))

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusFound, rec.Code, "should redirect with 302 status")
	assert.Contains(t, rec.Header().Get("Location"), "/auth/login", "should redirect to login")
	mockService.AssertNotCalled(t, "GetUserByToken")
	mockService.AssertNotCalled(t, "RefreshSession")
}

// TestAuthMiddleware_NoAccessTokenValidRefreshToken tests token refresh when access token is missing
func TestAuthMiddleware_NoAccessTokenValidRefreshToken(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testUserID := "user-456"
	testEmail := "user@example.com"
	testRefreshToken := "valid-refresh-token"
	testNewAccessToken := "new-access-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return("", errors.New("no access token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return(testRefreshToken, nil)
	mockService.On("RefreshSession", mock.Anything, testRefreshToken).Return(&UserSession{
		UserID:       testUserID,
		Email:        testEmail,
		AccessToken:  testNewAccessToken,
		RefreshToken: testRefreshToken,
	}, nil)
	mockSessionMgr.On("UpdateAccessToken", mock.Anything, testNewAccessToken)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	c := newTestEchoContext(req)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, testUserID, c.Get(userIDKey), "user ID should be set in context")
	assert.Equal(t, testEmail, c.Get(userEmailKey), "email should be set in context")
	mockService.AssertCalled(t, "RefreshSession", mock.Anything, testRefreshToken)
	mockSessionMgr.AssertCalled(t, "UpdateAccessToken", mock.Anything, testNewAccessToken)
	mockSessionMgr.AssertNotCalled(t, "ClearSessionCookies")
}

// TestAuthMiddleware_NoAccessTokenRefreshFails tests redirect when refresh token fails
func TestAuthMiddleware_NoAccessTokenRefreshFails(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testRefreshToken := "expired-refresh-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return("", errors.New("no access token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return(testRefreshToken, nil)
	mockService.On("RefreshSession", mock.Anything, testRefreshToken).Return(nil, errors.New("refresh failed"))
	mockSessionMgr.On("ClearSessionCookies", mock.Anything)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusFound, rec.Code, "should redirect with 302 status")
	assert.Contains(t, rec.Header().Get("Location"), "/auth/login", "should redirect to login")
	mockSessionMgr.AssertCalled(t, "ClearSessionCookies", mock.Anything)
}

// TestAuthMiddleware_InvalidAccessTokenValidRefreshToken tests token refresh when access token is invalid
func TestAuthMiddleware_InvalidAccessTokenValidRefreshToken(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testUserID := "user-789"
	testEmail := "user@example.com"
	testAccessToken := "invalid-access-token"
	testRefreshToken := "valid-refresh-token"
	testNewAccessToken := "new-access-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return(testAccessToken, nil)
	mockService.On("GetUserByToken", mock.Anything, testAccessToken).Return(nil, errors.New("invalid token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return(testRefreshToken, nil)
	mockService.On("RefreshSession", mock.Anything, testRefreshToken).Return(&UserSession{
		UserID:       testUserID,
		Email:        testEmail,
		AccessToken:  testNewAccessToken,
		RefreshToken: testRefreshToken,
	}, nil)
	mockSessionMgr.On("UpdateAccessToken", mock.Anything, testNewAccessToken)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	c := newTestEchoContext(req)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, testUserID, c.Get(userIDKey), "user ID should be set in context")
	assert.Equal(t, testEmail, c.Get(userEmailKey), "email should be set in context")
	mockService.AssertCalled(t, "RefreshSession", mock.Anything, testRefreshToken)
	mockSessionMgr.AssertCalled(t, "UpdateAccessToken", mock.Anything, testNewAccessToken)
	mockSessionMgr.AssertNotCalled(t, "ClearSessionCookies")
}

// TestAuthMiddleware_InvalidAccessTokenNoRefreshToken tests redirect when access token invalid and no refresh token
func TestAuthMiddleware_InvalidAccessTokenNoRefreshToken(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testAccessToken := "invalid-access-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return(testAccessToken, nil)
	mockService.On("GetUserByToken", mock.Anything, testAccessToken).Return(nil, errors.New("invalid token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return("", errors.New("no refresh token"))
	mockSessionMgr.On("ClearSessionCookies", mock.Anything)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusFound, rec.Code, "should redirect with 302 status")
	assert.Contains(t, rec.Header().Get("Location"), "/auth/login", "should redirect to login")
	mockSessionMgr.AssertCalled(t, "ClearSessionCookies", mock.Anything)
}

// TestAuthMiddleware_InvalidAccessTokenRefreshFails tests redirect when both access and refresh tokens fail
func TestAuthMiddleware_InvalidAccessTokenRefreshFails(t *testing.T) {
	mockService := new(MockAuthService)
	mockSessionMgr := new(MockSessionManager)
	logger := newTestLogger()

	testAccessToken := "invalid-access-token"
	testRefreshToken := "expired-refresh-token"

	// Setup expectations
	mockSessionMgr.On("GetAccessToken", mock.Anything).Return(testAccessToken, nil)
	mockService.On("GetUserByToken", mock.Anything, testAccessToken).Return(nil, errors.New("invalid token"))
	mockSessionMgr.On("GetRefreshToken", mock.Anything).Return(testRefreshToken, nil)
	mockService.On("RefreshSession", mock.Anything, testRefreshToken).Return(nil, errors.New("refresh failed"))
	mockSessionMgr.On("ClearSessionCookies", mock.Anything)

	middleware := AuthMiddleware(mockService, mockSessionMgr, logger)

	// Create test request and context
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	// Track if next handler was called
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	// Execute middleware
	err := middleware(nextHandler)(c)

	// Assertions
	assert.NoError(t, err)
	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusFound, rec.Code, "should redirect with 302 status")
	assert.Contains(t, rec.Header().Get("Location"), "/auth/login", "should redirect to login")
	mockSessionMgr.AssertCalled(t, "ClearSessionCookies", mock.Anything)
}

// TestGetUserID_Success tests successful user ID retrieval from context
func TestGetUserID_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	testUserID := "user-123"
	c.Set(userIDKey, testUserID)

	result := GetUserID(c)

	assert.Equal(t, testUserID, result, "should return user ID from context")
}

// TestGetUserID_NotFound tests user ID retrieval when not set in context
func TestGetUserID_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	result := GetUserID(c)

	assert.Equal(t, "", result, "should return empty string when user ID not in context")
}

// TestGetUserID_WrongType tests user ID retrieval when context value is wrong type
func TestGetUserID_WrongType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	c.Set(userIDKey, 123) // Set as integer instead of string

	result := GetUserID(c)

	assert.Equal(t, "", result, "should return empty string when context value is not a string")
}

// TestGetUserEmail_Success tests successful email retrieval from context
func TestGetUserEmail_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	testEmail := "user@example.com"
	c.Set(userEmailKey, testEmail)

	result := GetUserEmail(c)

	assert.Equal(t, testEmail, result, "should return email from context")
}

// TestGetUserEmail_NotFound tests email retrieval when not set in context
func TestGetUserEmail_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	result := GetUserEmail(c)

	assert.Equal(t, "", result, "should return empty string when email not in context")
}

// TestGetUserEmail_WrongType tests email retrieval when context value is wrong type
func TestGetUserEmail_WrongType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newTestEchoContext(req)

	c.Set(userEmailKey, map[string]string{"email": "test"}) // Set as map instead of string

	result := GetUserEmail(c)

	assert.Equal(t, "", result, "should return empty string when context value is not a string")
}
