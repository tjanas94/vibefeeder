-- migration: add_performance_indexes
-- description: adds composite performance indexes for optimized query patterns
-- tables affected: feeds, articles, summaries, events
-- special notes: replaces simple indexes with composite ones for better query performance

-- ============================================================================
-- feeds table indexes
-- ============================================================================

-- drop existing simple index on user_id (will be replaced by composite indexes)
drop index if exists idx_feeds_user_id;

-- composite index for case-insensitive feed name search within user's feeds
-- usage: searching/filtering feeds by name for a specific user
-- prefix matching: covers queries filtering by user_id alone
create index idx_feeds_user_name on feeds(user_id, lower(name));

comment on index idx_feeds_user_name is 'optimizes case-insensitive feed name search within user feeds';

-- composite index for status filtering within user's feeds
-- partial index: only indexes rows where last_fetch_status is not null
-- usage: finding feeds with errors or specific status for a user
-- prefix matching: also covers queries filtering by user_id alone
create index idx_feeds_user_status 
on feeds(user_id, last_fetch_status) 
where last_fetch_status is not null;

comment on index idx_feeds_user_status is 'optimizes status filtering within user feeds (partial index)';

-- index for bot feed selection ordering by last fetch time
-- nulls first: prioritizes feeds that have never been fetched
-- usage: background job selecting next feeds to fetch
create index idx_feeds_last_fetched 
on feeds(last_fetched_at nulls first);

comment on index idx_feeds_last_fetched is 'optimizes feed selection for background fetch job (nulls first)';

-- ============================================================================
-- articles table indexes
-- ============================================================================

-- drop existing simple indexes (will be replaced by composite index)
drop index if exists idx_articles_feed_id;
drop index if exists idx_articles_published_at;

-- composite index for recent articles retrieval per feed
-- descending order: optimizes "latest articles first" queries
-- usage: fetching recent articles from a specific feed (e.g., for summary generation)
-- prefix matching: also covers queries filtering by feed_id alone
create index idx_articles_feed_published 
on articles(feed_id, published_at desc);

comment on index idx_articles_feed_published is 'optimizes recent articles retrieval per feed';

-- ============================================================================
-- summaries table indexes
-- ============================================================================

-- drop existing simple indexes (will be replaced by composite index)
drop index if exists idx_summaries_user_id;
drop index if exists idx_summaries_created_at;

-- composite index for latest summary per user lookup
-- descending order: optimizes "most recent summary first" queries
-- usage: displaying latest summary for a user, summary history
-- prefix matching: also covers queries filtering by user_id alone
create index idx_summaries_user_created 
on summaries(user_id, created_at desc);

comment on index idx_summaries_user_created is 'optimizes latest summary lookup per user';

-- ============================================================================
-- events table indexes
-- ============================================================================

-- drop existing simple indexes on user_id and created_at (will be replaced by composite)
-- keep idx_events_event_type as it serves a different query pattern
drop index if exists idx_events_user_id;
drop index if exists idx_events_created_at;

-- composite index for time-based event analytics per user
-- descending order: optimizes "recent events first" queries
-- usage: user-specific event timelines, engagement metrics per user
-- prefix matching: also covers queries filtering by user_id alone
create index idx_events_user_created 
on events(user_id, created_at desc);

comment on index idx_events_user_created is 'optimizes time-based event analytics per user';

-- note: idx_events_event_type remains as simple index
-- rationale: serves global event analytics by type (not user-specific)
