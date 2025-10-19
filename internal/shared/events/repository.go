package events

import (
	"context"
	"fmt"

	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

// EventRepository defines the interface for event logging
type EventRepository interface {
	RecordEvent(ctx context.Context, event database.PublicEventsInsert) error
}

// Repository handles data access for events
type Repository struct {
	db *database.Client
}

// Ensure Repository implements EventRepository interface at compile time
var _ EventRepository = (*Repository)(nil)

// NewRepository creates a new events repository
func NewRepository(db *database.Client) *Repository {
	return &Repository{db: db}
}

// RecordEvent creates a new event in the database
func (r *Repository) RecordEvent(ctx context.Context, event database.PublicEventsInsert) error {
	var result []database.PublicEventsSelect
	_, err := r.db.From("events").Insert(event, false, "", "", "").ExecuteTo(&result)
	if err != nil {
		return fmt.Errorf("failed to record event: %w", err)
	}
	return nil
}
