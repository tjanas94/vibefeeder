package models

import "testing"

func TestLoginPageQuery_ToViewProps_WithConfirmedFlag(t *testing.T) {
	query := LoginPageQuery{
		Confirmed: true,
	}

	props := query.ToViewProps()

	if !props.ShowConfirmedToast {
		t.Error("expected ShowConfirmedToast to be true")
	}
	if props.ShowResetSuccessToast {
		t.Error("expected ShowResetSuccessToast to be false")
	}
	if props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be false")
	}
}

func TestLoginPageQuery_ToViewProps_WithResetSuccessFlag(t *testing.T) {
	query := LoginPageQuery{
		ResetSuccess: true,
	}

	props := query.ToViewProps()

	if props.ShowConfirmedToast {
		t.Error("expected ShowConfirmedToast to be false")
	}
	if !props.ShowResetSuccessToast {
		t.Error("expected ShowResetSuccessToast to be true")
	}
	if props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be false")
	}
}

func TestLoginPageQuery_ToViewProps_WithMissingTokenError(t *testing.T) {
	query := LoginPageQuery{
		Error: "missing_token",
	}

	props := query.ToViewProps()

	if !props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be true")
	}
	expectedMsg := "Invalid confirmation link. Please check your email and try again."
	if props.ErrorMessage != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, props.ErrorMessage)
	}
}

func TestLoginPageQuery_ToViewProps_WithInvalidTokenError(t *testing.T) {
	query := LoginPageQuery{
		Error: "invalid_token",
	}

	props := query.ToViewProps()

	if !props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be true")
	}
	expectedMsg := "Confirmation link is invalid or expired. Please register again or contact support."
	if props.ErrorMessage != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, props.ErrorMessage)
	}
}

func TestLoginPageQuery_ToViewProps_WithUnknownError(t *testing.T) {
	query := LoginPageQuery{
		Error: "unknown_error",
	}

	props := query.ToViewProps()

	if !props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be true")
	}
	expectedMsg := "An error occurred during confirmation. Please try again."
	if props.ErrorMessage != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, props.ErrorMessage)
	}
}

func TestLoginPageQuery_ToViewProps_WithMultipleFlags(t *testing.T) {
	query := LoginPageQuery{
		Confirmed:    true,
		ResetSuccess: true,
		Error:        "",
	}

	props := query.ToViewProps()

	if !props.ShowConfirmedToast {
		t.Error("expected ShowConfirmedToast to be true")
	}
	if !props.ShowResetSuccessToast {
		t.Error("expected ShowResetSuccessToast to be true")
	}
	if props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be false when error is empty")
	}
}

func TestLoginPageQuery_ToViewProps_Empty(t *testing.T) {
	query := LoginPageQuery{}

	props := query.ToViewProps()

	// Check that the returned props has the expected zero values
	if props.ShowConfirmedToast {
		t.Error("expected ShowConfirmedToast to be false")
	}
	if props.ShowResetSuccessToast {
		t.Error("expected ShowResetSuccessToast to be false")
	}
	if props.ShowErrorAlert {
		t.Error("expected ShowErrorAlert to be false")
	}
	if props.ErrorMessage != "" {
		t.Errorf("expected empty ErrorMessage, got %q", props.ErrorMessage)
	}
}

func TestMapErrorMessage_MissingToken(t *testing.T) {
	msg := mapErrorMessage("missing_token")
	expected := "Invalid confirmation link. Please check your email and try again."
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestMapErrorMessage_InvalidToken(t *testing.T) {
	msg := mapErrorMessage("invalid_token")
	expected := "Confirmation link is invalid or expired. Please register again or contact support."
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestMapErrorMessage_Unknown(t *testing.T) {
	msg := mapErrorMessage("some_unknown_error")
	expected := "An error occurred during confirmation. Please try again."
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestMapErrorMessage_Empty(t *testing.T) {
	msg := mapErrorMessage("")
	expected := "An error occurred during confirmation. Please try again."
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}
