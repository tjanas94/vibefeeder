-- migration: add_feeds_retry_count
-- description: adds retry_count column to feeds table for exponential backoff tracking
-- tables affected: feeds
-- special notes: tracks consecutive failed fetch attempts to implement exponential backoff

-- add retry_count column to track consecutive failed fetch attempts
-- used in conjunction with fetch_after for exponential backoff strategy
-- resets to 0 on successful fetch
alter table feeds 
add column retry_count integer not null default 0;

comment on column feeds.retry_count is 'consecutive failed fetch attempts (for exponential backoff)';
