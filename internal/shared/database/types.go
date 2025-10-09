package database

type PublicFeedsSelect struct {
	CreatedAt       string  `json:"created_at"`
	Id              string  `json:"id"`
	LastFetchError  *string `json:"last_fetch_error"`
	LastFetchStatus *string `json:"last_fetch_status"`
	Name            string  `json:"name"`
	UpdatedAt       string  `json:"updated_at"`
	Url             string  `json:"url"`
	UserId          string  `json:"user_id"`
}

type PublicFeedsInsert struct {
	CreatedAt       *string `json:"created_at"`
	Id              *string `json:"id"`
	LastFetchError  *string `json:"last_fetch_error"`
	LastFetchStatus *string `json:"last_fetch_status"`
	Name            string  `json:"name"`
	UpdatedAt       *string `json:"updated_at"`
	Url             string  `json:"url"`
	UserId          string  `json:"user_id"`
}

type PublicFeedsUpdate struct {
	CreatedAt       *string `json:"created_at"`
	Id              *string `json:"id"`
	LastFetchError  *string `json:"last_fetch_error"`
	LastFetchStatus *string `json:"last_fetch_status"`
	Name            *string `json:"name"`
	UpdatedAt       *string `json:"updated_at"`
	Url             *string `json:"url"`
	UserId          *string `json:"user_id"`
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
	CreatedAt   *string `json:"created_at"`
	FeedId      string  `json:"feed_id"`
	Id          *string `json:"id"`
	PublishedAt string  `json:"published_at"`
	Title       string  `json:"title"`
	Url         string  `json:"url"`
}

type PublicArticlesUpdate struct {
	Content     *string `json:"content"`
	CreatedAt   *string `json:"created_at"`
	FeedId      *string `json:"feed_id"`
	Id          *string `json:"id"`
	PublishedAt *string `json:"published_at"`
	Title       *string `json:"title"`
	Url         *string `json:"url"`
}

type PublicSummariesSelect struct {
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
}

type PublicSummariesInsert struct {
	Content   string  `json:"content"`
	CreatedAt *string `json:"created_at"`
	Id        *string `json:"id"`
	UserId    string  `json:"user_id"`
}

type PublicSummariesUpdate struct {
	Content   *string `json:"content"`
	CreatedAt *string `json:"created_at"`
	Id        *string `json:"id"`
	UserId    *string `json:"user_id"`
}

type PublicEventsSelect struct {
	CreatedAt string      `json:"created_at"`
	EventType string      `json:"event_type"`
	Id        string      `json:"id"`
	Metadata  interface{} `json:"metadata"`
	UserId    *string     `json:"user_id"`
}

type PublicEventsInsert struct {
	CreatedAt *string     `json:"created_at"`
	EventType string      `json:"event_type"`
	Id        *string     `json:"id"`
	Metadata  interface{} `json:"metadata"`
	UserId    *string     `json:"user_id"`
}

type PublicEventsUpdate struct {
	CreatedAt *string     `json:"created_at"`
	EventType *string     `json:"event_type"`
	Id        *string     `json:"id"`
	Metadata  interface{} `json:"metadata"`
	UserId    *string     `json:"user_id"`
}
