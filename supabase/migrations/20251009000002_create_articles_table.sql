-- migration: create_articles_table
-- description: creates the articles table to store articles fetched from rss feeds
-- tables affected: articles
-- special notes: includes rls policies and performance indexes for published_at queries

-- create the articles table
create table articles (
    id uuid primary key default gen_random_uuid(),
    feed_id uuid not null references feeds(id) on delete cascade,
    title text not null,
    url text not null,
    content text null,
    published_at timestamptz not null,
    created_at timestamptz not null default now(),
    
    -- prevent duplicate articles from the same feed
    constraint unique_feed_article unique(feed_id, url)
);

-- create index on feed_id for efficient joins with feeds table
create index idx_articles_feed_id on articles(feed_id);

-- create index on published_at for efficient filtering of recent articles (e.g., last 24 hours)
-- sorted descending for optimal performance when fetching latest articles first
create index idx_articles_published_at on articles(published_at desc);

-- enable row level security
alter table articles enable row level security;

-- rls policy: allow authenticated users to view articles from their feeds
-- rationale: users can only see articles from feeds they own
-- uses subquery to check feed ownership
create policy "authenticated users can view articles from their feeds"
on articles for select
to authenticated
using (
    exists (
        select 1 from feeds
        where feeds.id = articles.feed_id
        and feeds.user_id = auth.uid()
    )
);

-- rls policy: deny anonymous users from viewing articles
-- rationale: articles are private to authenticated users who own the feeds
create policy "anonymous users cannot view articles"
on articles for select
to anon
using (false);

-- note: no insert, update, or delete policies for regular users
-- articles are managed exclusively by background jobs using service role
-- this prevents users from manually creating, modifying, or deleting articles
-- only the system can perform these operations to maintain data integrity

-- add comment to table
comment on table articles is 'stores articles fetched from rss feeds';

-- add comments to columns
comment on column articles.id is 'unique identifier for the article';
comment on column articles.feed_id is 'reference to the source feed';
comment on column articles.title is 'article title';
comment on column articles.url is 'article url';
comment on column articles.content is 'article content or description';
comment on column articles.published_at is 'when the article was published';
comment on column articles.created_at is 'when the article was fetched';
