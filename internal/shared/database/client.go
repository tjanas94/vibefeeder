package database

import (
	"context"
	"fmt"

	"github.com/supabase-community/supabase-go"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

// contextKeyType is the private key type for the context.
type contextKeyType struct{}

// accessTokenContextKey is used to store the access token in the request context
// This token is used for Row Level Security (RLS) in Supabase
var accessTokenContextKey = contextKeyType{}

// Client wraps the Supabase client and provides typed access to database operations
type Client struct {
	*supabase.Client
	cfg config.SupabaseConfig
}

// New creates a new database client using the provided configuration
func New(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	client, err := supabase.NewClient(
		cfg.Supabase.URL,
		cfg.Supabase.Key,
		&supabase.ClientOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &Client{
		Client: client,
		cfg:    cfg.Supabase,
	}, nil
}

// Health checks if the database connection is working
func (c *Client) Health() error {
	if c.Client == nil {
		return fmt.Errorf("database client is not initialized")
	}
	return nil
}

// ContextWithToken adds the access token to the context for RLS
func ContextWithToken(parent context.Context, token string) context.Context {
	return context.WithValue(parent, accessTokenContextKey, token)
}

// GetAccessToken retrieves the access token from the context
// Returns error if no token is found
func GetAccessToken(ctx context.Context) (string, error) {
	token, ok := ctx.Value(accessTokenContextKey).(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access token in context")
	}
	return token, nil
}

// NewAuthenticatedClient creates a new supabase client with the access token from context
// This allows safe per-request authentication for RLS without modifying the shared client
func (c *Client) NewAuthenticatedClient(ctx context.Context) (*supabase.Client, error) {
	token, err := GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	client, err := supabase.NewClient(
		c.cfg.URL,
		c.cfg.Key,
		&supabase.ClientOptions{
			Headers: headers,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated client: %w", err)
	}

	return client, nil
}
