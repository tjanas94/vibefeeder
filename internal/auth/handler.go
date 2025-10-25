package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tjanas94/vibefeeder/internal/auth/models"
	"github.com/tjanas94/vibefeeder/internal/auth/view"
	sharedAuth "github.com/tjanas94/vibefeeder/internal/shared/auth"
	sharederrors "github.com/tjanas94/vibefeeder/internal/shared/errors"
	"github.com/tjanas94/vibefeeder/internal/shared/validator"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service        *Service
	sessionManager sharedAuth.SessionManager
	requireRegCode bool // Whether registration code is required
}

// NewHandler creates a new auth handler
func NewHandler(service *Service, sessionManager sharedAuth.SessionManager, requireRegCode bool) *Handler {
	return &Handler{
		service:        service,
		sessionManager: sessionManager,
		requireRegCode: requireRegCode,
	}
}

// ShowLoginPage renders the login page
func (h *Handler) ShowLoginPage(c echo.Context) error {
	var query models.LoginPageQuery
	_ = c.Bind(&query)
	props := query.ToViewProps()
	return c.Render(http.StatusOK, "", view.LoginPage(props))
}

// HandleLogin processes login form submission
func (h *Handler) HandleLogin(c echo.Context) error {
	var req models.LoginRequest
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Path 2: Handle validation errors (invalid data)
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
		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			props := view.LoginPageProps{
				Email:        req.Email,
				GeneralError: serviceErr.Message,
			}
			return c.Render(serviceErr.Code, "", view.LoginForm(props))
		}
		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// Set session cookies
	h.sessionManager.SetSessionCookies(c, session)

	// Redirect to dashboard using HX-Redirect header
	return h.htmxRedirect(c, "/dashboard")
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
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Path 2: Handle validation errors (invalid data)
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

		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			props.GeneralError = serviceErr.Message
			props.EmailError = serviceErr.FieldErrors["Email"]
			props.RegistrationCodeError = serviceErr.FieldErrors["RegistrationCode"]
			return c.Render(serviceErr.Code, "", view.RegisterForm(props))
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
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
		return c.Redirect(http.StatusFound, "/auth/login?error=missing_token")
	}

	// Verify the email confirmation token
	if err := h.service.VerifyEmailConfirmation(c.Request().Context(), token); err != nil {
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
	return h.htmxRedirect(c, "/auth/login")
}

// ShowForgotPasswordPage renders the forgot password page
func (h *Handler) ShowForgotPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "", view.ForgotPasswordPage(view.ForgotPasswordPageProps{}))
}

// HandleForgotPassword processes forgot password form submission
func (h *Handler) HandleForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Path 2: Handle validation errors (invalid data)
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
	// Path 1: Handle bind errors (invalid request format)
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Path 2: Handle validation errors (invalid data)
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

		// Path 3: Handle business errors (ServiceError)
		var serviceErr *sharederrors.ServiceError
		if errors.As(err, &serviceErr) {
			props.GeneralError = serviceErr.Message
			props.PasswordError = serviceErr.FieldErrors["Password"]
			return c.Render(serviceErr.Code, "", view.ResetPasswordForm(props))
		}

		// Path 4: Unexpected error - delegate to global error handler
		return err
	}

	// DO NOT set session cookies - user must log in with new password
	// Redirect to login with success message
	return h.htmxRedirect(c, "/auth/login?reset_success=true")
}

// htmxRedirect sets the HX-Redirect header and returns NoContent status
func (h *Handler) htmxRedirect(c echo.Context, path string) error {
	c.Response().Header().Set("HX-Redirect", path)
	return c.NoContent(http.StatusOK)
}
