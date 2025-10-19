package view

// AuthLayoutProps contains the data needed to render the authentication layout
type AuthLayoutProps struct {
	Title string
}

// LoginPageProps contains the data needed to render the login page
type LoginPageProps struct {
	Email                 string
	EmailError            string
	PasswordError         string
	GeneralError          string
	ShowConfirmedToast    bool
	ShowResetSuccessToast bool
	ShowErrorAlert        bool   // Show error alert for confirmation errors
	ErrorMessage          string // Error message to display in alert
}

// RegisterPageProps contains the data needed to render the registration page
type RegisterPageProps struct {
	Email                 string
	EmailError            string
	PasswordError         string
	PasswordConfirmError  string
	RegistrationCode      string
	RegistrationCodeError string
	RequireCode           bool // Whether registration code is required (based on config)
	GeneralError          string
}

// RegistrationPendingProps contains the data needed to render the registration pending view
type RegistrationPendingProps struct {
	Email string
}

// ForgotPasswordPageProps contains the data needed to render the forgot password page
type ForgotPasswordPageProps struct {
	Email       string
	EmailError  string
	ShowSuccess bool
}

// ResetPasswordPageProps contains the data needed to render the reset password page
type ResetPasswordPageProps struct {
	Token                string
	PasswordError        string
	PasswordConfirmError string
	GeneralError         string
}
