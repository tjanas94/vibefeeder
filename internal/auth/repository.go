package auth

import (
	"context"
	"fmt"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// Repository handles data access for authentication
type Repository struct {
	db *database.Client
}

// NewRepository creates a new auth repository
func NewRepository(db *database.Client) *Repository {
	return &Repository{db: db}
}

// InsertEvent creates a new event in the database
func (r *Repository) InsertEvent(ctx context.Context, event database.PublicEventsInsert) error {
	var result []database.PublicEventsSelect
	_, err := r.db.From("events").Insert(event, false, "", "", "").ExecuteTo(&result)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}
