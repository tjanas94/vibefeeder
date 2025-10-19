package auth

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	"github.com/tjanas94/vibefeeder/internal/auth/view"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service        *Service
	sessionManager *SessionManager
	logger         *slog.Logger
	requireRegCode bool // Whether registration code is required
}

// NewHandler creates a new auth handler
func NewHandler(service *Service, sessionManager *SessionManager, logger *slog.Logger, requireRegCode bool) *Handler {
	return &Handler{
		service:        service,
		sessionManager: sessionManager,
		logger:         logger,
		requireRegCode: requireRegCode,
	}
}

// ShowLoginPage renders the login page
func (h *Handler) ShowLoginPage(c echo.Context) error {
	props := view.LoginPageProps{
		ShowConfirmedToast:    c.QueryParam("confirmed") == "true",
		ShowResetSuccessToast: c.QueryParam("reset_success") == "true",
	}

	// Handle confirmation errors
	errorParam := c.QueryParam("error")
	if errorParam != "" {
		props.ShowErrorAlert = true
		switch errorParam {
		case "missing_token":
			props.ErrorMessage = "Invalid confirmation link. Please check your email and try again."
		case "invalid_token":
			props.ErrorMessage = "Confirmation link is invalid or expired. Please register again or contact support."
		default:
			props.ErrorMessage = "An error occurred during confirmation. Please try again."
		}
	}

	return c.Render(http.StatusOK, "", view.LoginPage(props))
}

// HandleLogin processes login form submission
func (h *Handler) HandleLogin(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind login request", "error", err)
		return c.Render(http.StatusBadRequest, "", view.LoginForm(view.LoginPageProps{
			GeneralError: "Invalid form data",
		}))
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		fieldErrors := validator.ParseFieldErrors(err)

		props := view.LoginPageProps{
			Email:         req.Email,
			EmailError:    fieldErrors["Email"],
			PasswordError: fieldErrors["Password"],
		}

		return c.Render(http.StatusUnprocessableEntity, "", view.LoginForm(props))
	}

	// Attempt login
	session, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		props := view.LoginPageProps{
			Email:        req.Email,
			GeneralError: "Invalid email or password",
		}
		return c.Render(http.StatusUnauthorized, "", view.LoginForm(props))
	}

	// Set session cookies
	h.sessionManager.SetSessionCookies(c, session)

	// Redirect to dashboard using HX-Redirect header
	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return c.NoContent(http.StatusOK)
}

// ShowRegisterPage renders the registration page
func (h *Handler) ShowRegisterPage(c echo.Context) error {
	return c.Render(http.StatusOK, "", view.RegisterPage(view.RegisterPageProps{
		RequireCode: h.requireRegCode,
	}))
}

// HandleRegister processes registration form submission
func (h *Handler) HandleRegister(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind register request", "error", err)
		return c.Render(http.StatusBadRequest, "", view.RegisterForm(view.RegisterPageProps{
			GeneralError: "Invalid form data",
			RequireCode:  h.requireRegCode,
		}))
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		fieldErrors := validator.ParseFieldErrors(err)

		props := view.RegisterPageProps{
			Email:                req.Email,
			RegistrationCode:     req.RegistrationCode,
			EmailError:           fieldErrors["Email"],
			PasswordError:        fieldErrors["Password"],
			PasswordConfirmError: fieldErrors["PasswordConfirm"],
			RequireCode:          h.requireRegCode,
		}

		return c.Render(http.StatusUnprocessableEntity, "", view.RegisterForm(props))
	}

	// Attempt registration
	if err := h.service.Register(c.Request().Context(), req); err != nil {
		props := view.RegisterPageProps{
			Email:            req.Email,
			RegistrationCode: req.RegistrationCode,
			RequireCode:      h.requireRegCode,
		}

		switch err {
		case ErrUserAlreadyExists:
			props.EmailError = "User with this email already exists"
		case ErrInvalidRegistrationCode:
			props.RegistrationCodeError = "Invalid registration code"
		default:
			props.GeneralError = "Failed to register account. Please try again later"
		}

		return c.Render(http.StatusUnprocessableEntity, "", view.RegisterForm(props))
	}

	// Show pending confirmation view
	return c.Render(http.StatusOK, "", view.RegistrationPending(view.RegistrationPendingProps{
		Email: req.Email,
	}))
}

// HandleConfirm handles email confirmation redirect
func (h *Handler) HandleConfirm(c echo.Context) error {
	// Get token from query parameter
	token := c.QueryParam("token")
	if token == "" {
		h.logger.Error("Email confirmation failed: missing token")
		return c.Redirect(http.StatusFound, "/auth/login?error=missing_token")
	}

	// Verify the email confirmation token
	if err := h.service.VerifyEmailConfirmation(c.Request().Context(), token); err != nil {
		h.logger.Error("Email confirmation failed", "error", err)
		return c.Redirect(http.StatusFound, "/auth/login?error=invalid_token")
	}

	// Redirect to login page with confirmation message
	return c.Redirect(http.StatusFound, "/auth/login?confirmed=true")
}

// HandleLogout processes user logout
func (h *Handler) HandleLogout(c echo.Context) error {
	// Get access token
	accessToken, err := h.sessionManager.GetAccessToken(c)
	if err == nil && accessToken != "" {
		// Attempt to logout from Supabase (best effort, don't fail if it errors)
		_ = h.service.Logout(c.Request().Context(), accessToken)
	}

	// Clear session cookies
	h.sessionManager.ClearSessionCookies(c)

	// Redirect to login page
	c.Response().Header().Set("HX-Redirect", "/auth/login")
	return c.NoContent(http.StatusOK)
}

// ShowForgotPasswordPage renders the forgot password page
func (h *Handler) ShowForgotPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "", view.ForgotPasswordPage(view.ForgotPasswordPageProps{}))
}

// HandleForgotPassword processes forgot password form submission
func (h *Handler) HandleForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind forgot password request", "error", err)
		return c.Render(http.StatusBadRequest, "", view.ForgotPasswordForm(view.ForgotPasswordPageProps{
			EmailError: "Invalid form data",
		}))
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		fieldErrors := validator.ParseFieldErrors(err)

		props := view.ForgotPasswordPageProps{
			Email:      req.Email,
			EmailError: fieldErrors["Email"],
		}

		return c.Render(http.StatusUnprocessableEntity, "", view.ForgotPasswordForm(props))
	}

	// Send reset email (always succeed for security - don't reveal if email exists)
	_ = h.service.SendPasswordResetEmail(c.Request().Context(), req.Email)

	// Show success message
	return c.Render(http.StatusOK, "", view.ForgotPasswordForm(view.ForgotPasswordPageProps{
		ShowSuccess: true,
	}))
}

// ShowResetPasswordPage renders the reset password page
func (h *Handler) ShowResetPasswordPage(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		// If no token, redirect to login
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	// Render the reset password form with token (will be verified on POST)
	return c.Render(http.StatusOK, "", view.ResetPasswordPage(view.ResetPasswordPageProps{
		Token: token,
	}))
}

// HandleResetPassword processes reset password form submission
func (h *Handler) HandleResetPassword(c echo.Context) error {
	var req models.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind reset password request", "error", err)
		return c.Render(http.StatusBadRequest, "", view.ResetPasswordForm(view.ResetPasswordPageProps{
			Token:        req.Token,
			GeneralError: "Invalid form data",
		}))
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		fieldErrors := validator.ParseFieldErrors(err)

		props := view.ResetPasswordPageProps{
			Token:                req.Token,
			PasswordError:        fieldErrors["Password"],
			PasswordConfirmError: fieldErrors["PasswordConfirm"],
		}

		return c.Render(http.StatusUnprocessableEntity, "", view.ResetPasswordForm(props))
	}

	// Attempt password reset
	if err := h.service.ResetPassword(c.Request().Context(), req.Token, req.Password); err != nil {
		props := view.ResetPasswordPageProps{
			Token: req.Token,
		}

		switch err {
		case ErrSamePassword:
			props.PasswordError = "New password must be different from your current password"
			return c.Render(http.StatusUnprocessableEntity, "", view.ResetPasswordForm(props))
		case ErrInvalidToken:
			props.GeneralError = "Password reset link is invalid or expired. Please request a new one."
			return c.Render(http.StatusBadRequest, "", view.ResetPasswordForm(props))
		default:
			props.GeneralError = "Failed to reset password. Please try again later"
			return c.Render(http.StatusBadRequest, "", view.ResetPasswordForm(props))
		}
	}

	// DO NOT set session cookies - user must log in with new password
	// Redirect to login with success message
	c.Response().Header().Set("HX-Redirect", "/auth/login?reset_success=true")
	return c.NoContent(http.StatusOK)
}
