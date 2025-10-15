package models

import (
	feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"
)

// DashboardViewModel contains the data needed to render the dashboard page
type DashboardViewModel struct {
	Title     string
	UserEmail string
	Query     *feedmodels.ListFeedsQuery // Query params for feed filtering (search, status, page)
}
