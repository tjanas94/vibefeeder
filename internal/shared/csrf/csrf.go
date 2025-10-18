package csrf

import "context"

// EchoContextKey is the key used to store the CSRF token in the echo.Context.
const EchoContextKey = "csrf"

// contextKey is the private key type for the standard context.
type contextKeyType struct{}

// contextKey is the private key for the standard context.
var contextKey = contextKeyType{}

// Token extracts the CSRF token from the context.Context.
// This helper is used in templ templates to include CSRF tokens in forms.
func Token(ctx context.Context) string {
	if token, ok := ctx.Value(contextKey).(string); ok {
		return token
	}
	return ""
}

// WithToken returns a new context with the CSRF token.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, contextKey, token)
}
