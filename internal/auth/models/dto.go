package models

// LoginRequest represents the login form data
type LoginRequest struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

// RegisterRequest represents the registration form data
type RegisterRequest struct {
	Email            string `form:"email" validate:"required,email"`
	Password         string `form:"password" validate:"required,strongpassword=50"`
	PasswordConfirm  string `form:"password_confirm" validate:"required,eqfield=Password"`
	RegistrationCode string `form:"registration_code"` // Validation handled in service based on config
}

// ForgotPasswordRequest represents the forgot password form data
type ForgotPasswordRequest struct {
	Email string `form:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the reset password form data
type ResetPasswordRequest struct {
	Token           string `form:"token" validate:"required"`
	Password        string `form:"password" validate:"required,strongpassword=50"`
	PasswordConfirm string `form:"password_confirm" validate:"required,eqfield=Password"`
}
