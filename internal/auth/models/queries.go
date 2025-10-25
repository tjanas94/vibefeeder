package models

import "github.com/tjanas94/vibefeeder/internal/auth/view"

// LoginPageQuery represents the query parameters for the login page
type LoginPageQuery struct {
	Confirmed    bool   `query:"confirmed"`
	ResetSuccess bool   `query:"reset_success"`
	Error        string `query:"error"`
}

// ToViewProps converts LoginPageQuery to LoginPageProps for rendering
func (q LoginPageQuery) ToViewProps() view.LoginPageProps {
	props := view.LoginPageProps{
		ShowConfirmedToast:    q.Confirmed,
		ShowResetSuccessToast: q.ResetSuccess,
	}

	if q.Error != "" {
		props.ShowErrorAlert = true
		props.ErrorMessage = mapErrorMessage(q.Error)
	}

	return props
}

// mapErrorMessage maps error codes to user-friendly messages
func mapErrorMessage(errorCode string) string {
	switch errorCode {
	case "missing_token":
		return "Invalid confirmation link. Please check your email and try again."
	case "invalid_token":
		return "Confirmation link is invalid or expired. Please register again or contact support."
	default:
		return "An error occurred during confirmation. Please try again."
	}
}
