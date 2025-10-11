-- migration: add_feeds_fetch_columns
-- description: adds http caching and fetch scheduling columns to feeds table
-- tables affected: feeds
-- special notes: these columns enable conditional http requests and intelligent fetch scheduling

-- add last_fetched_at column to track when the feed was last fetched
-- used by the background job to determine fetch order and frequency
alter table feeds 
add column last_fetched_at timestamptz null;

comment on column feeds.last_fetched_at is 'when the feed was last fetched by the background job';

-- add last_modified column to store http last-modified header value
-- enables conditional requests using if-modified-since header to save bandwidth
alter table feeds 
add column last_modified varchar(255) null;

comment on column feeds.last_modified is 'http last-modified header for conditional requests';

-- add etag column to store http etag header value
-- enables conditional requests using if-none-match header for efficient fetching
alter table feeds 
add column etag varchar(255) null;

comment on column feeds.etag is 'http etag header for conditional requests';

-- add fetch_after column to implement exponential backoff and respect retry-after headers
-- prevents hammering feeds that are temporarily unavailable
-- null value means "fetch as soon as possible"
alter table feeds 
add column fetch_after timestamptz null;

comment on column feeds.fetch_after is 'earliest time to fetch this feed (for exponential backoff and rate limiting)';
