package auth

// UserSession represents the authenticated user session data
type UserSession struct {
	UserID       string
	Email        string
	AccessToken  string
	RefreshToken string
}
