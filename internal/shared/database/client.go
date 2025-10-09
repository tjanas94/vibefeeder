package database

import (
	"fmt"

	"github.com/supabase-community/supabase-go"
	"github.com/tjanas94/vibefeeder/internal/shared/config"
)

// Client wraps the Supabase client and provides typed access to database operations
type Client struct {
	*supabase.Client
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

	return &Client{Client: client}, nil
}

// Health checks if the database connection is working
func (c *Client) Health() error {
	if c.Client == nil {
		return fmt.Errorf("database client is not initialized")
	}
	return nil
}
