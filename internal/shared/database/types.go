package database

type PublicFeedsSelect struct {
	CreatedAt       string  `json:"created_at"`
	Etag            *string `json:"etag"`
	FetchAfter      *string `json:"fetch_after"`
	Id              string  `json:"id"`
	LastFetchError  *string `json:"last_fetch_error"`
	LastFetchStatus *string `json:"last_fetch_status"`
	LastFetchedAt   *string `json:"last_fetched_at"`
	LastModified    *string `json:"last_modified"`
	Name            string  `json:"name"`
	RetryCount      int     `json:"retry_count"`
	UpdatedAt       string  `json:"updated_at"`
	Url             string  `json:"url"`
	UserId          string  `json:"user_id"`
}

type PublicFeedsInsert struct {
	CreatedAt       *string `json:"created_at,omitempty"`
	Etag            *string `json:"etag"`
	FetchAfter      *string `json:"fetch_after"`
	Id              *string `json:"id,omitempty"`
	LastFetchError  *string `json:"last_fetch_error"`
	LastFetchStatus *string `json:"last_fetch_status"`
	LastFetchedAt   *string `json:"last_fetched_at"`
	LastModified    *string `json:"last_modified"`
	Name            string  `json:"name"`
	RetryCount      *int    `json:"retry_count,omitempty"`
	UpdatedAt       *string `json:"updated_at,omitempty"`
	Url             string  `json:"url"`
	UserId          string  `json:"user_id"`
}

type PublicFeedsUpdate struct {
	CreatedAt       *string `json:"created_at,omitempty"`
	Etag            *string `json:"etag,omitempty"`
	FetchAfter      *string `json:"fetch_after,omitempty"`
	Id              *string `json:"id,omitempty"`
	LastFetchError  *string `json:"last_fetch_error,omitempty"`
	LastFetchStatus *string `json:"last_fetch_status,omitempty"`
	LastFetchedAt   *string `json:"last_fetched_at,omitempty"`
	LastModified    *string `json:"last_modified,omitempty"`
	Name            *string `json:"name,omitempty"`
	RetryCount      *int    `json:"retry_count,omitempty"`
	UpdatedAt       *string `json:"updated_at,omitempty"`
	Url             *string `json:"url,omitempty"`
	UserId          *string `json:"user_id,omitempty"`
}

type PublicArticlesSelect struct {
	Content     *string `json:"content"`
	CreatedAt   string  `json:"created_at"`
	FeedId      string  `json:"feed_id"`
	Id          string  `json:"id"`
	PublishedAt string  `json:"published_at"`
	Title       string  `json:"title"`
	Url         string  `json:"url"`
}

type PublicArticlesInsert struct {
	Content     *string `json:"content"`
	CreatedAt   *string `json:"created_at,omitempty"`
	FeedId      string  `json:"feed_id"`
	Id          *string `json:"id,omitempty"`
	PublishedAt string  `json:"published_at"`
	Title       string  `json:"title"`
	Url         string  `json:"url"`
}

type PublicArticlesUpdate struct {
	Content     *string `json:"content,omitempty"`
	CreatedAt   *string `json:"created_at,omitempty"`
	FeedId      *string `json:"feed_id,omitempty"`
	Id          *string `json:"id,omitempty"`
	PublishedAt *string `json:"published_at,omitempty"`
	Title       *string `json:"title,omitempty"`
	Url         *string `json:"url,omitempty"`
}

type PublicSummariesSelect struct {
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
}

type PublicSummariesInsert struct {
	Content   string  `json:"content"`
	CreatedAt *string `json:"created_at,omitempty"`
	Id        *string `json:"id,omitempty"`
	UserId    string  `json:"user_id"`
}

type PublicSummariesUpdate struct {
	Content   *string `json:"content,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	Id        *string `json:"id,omitempty"`
	UserId    *string `json:"user_id,omitempty"`
}

type PublicEventsSelect struct {
	CreatedAt string      `json:"created_at"`
	EventType string      `json:"event_type"`
	Id        string      `json:"id"`
	Metadata  interface{} `json:"metadata"`
	UserId    *string     `json:"user_id"`
}

type PublicEventsInsert struct {
	CreatedAt *string     `json:"created_at,omitempty"`
	EventType string      `json:"event_type"`
	Id        *string     `json:"id,omitempty"`
	Metadata  interface{} `json:"metadata"`
	UserId    *string     `json:"user_id"`
}

type PublicEventsUpdate struct {
	CreatedAt *string     `json:"created_at,omitempty"`
	EventType *string     `json:"event_type,omitempty"`
	Id        *string     `json:"id,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
	UserId    *string     `json:"user_id,omitempty"`
}
