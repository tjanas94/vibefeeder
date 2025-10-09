-- migration: create_feeds_table
-- description: creates the feeds table to store rss feed sources added by users
-- tables affected: feeds
-- special notes: includes rls policies for multi-tenancy isolation

-- create the feeds table
create table feeds (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users(id) on delete cascade,
    name text not null,
    url text not null,
    last_fetch_status text null,
    last_fetch_error text null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    
    -- prevent duplicate feeds for the same user
    constraint unique_user_feed unique(user_id, url)
);

-- create index on user_id for efficient joins and filtering
create index idx_feeds_user_id on feeds(user_id);

-- enable row level security
alter table feeds enable row level security;

-- rls policy: allow authenticated users to view only their own feeds
-- rationale: ensures data isolation between users
create policy "authenticated users can view their own feeds"
on feeds for select
to authenticated
using (auth.uid() = user_id);

-- rls policy: allow anonymous users no access to feeds
-- rationale: feeds are private to authenticated users only
create policy "anonymous users cannot view feeds"
on feeds for select
to anon
using (false);

-- rls policy: allow authenticated users to insert their own feeds
-- rationale: users can add new rss feeds to their account
create policy "authenticated users can insert their own feeds"
on feeds for insert
to authenticated
with check (auth.uid() = user_id);

-- rls policy: deny anonymous users from inserting feeds
-- rationale: only authenticated users can create feeds
create policy "anonymous users cannot insert feeds"
on feeds for insert
to anon
with check (false);

-- rls policy: allow authenticated users to update their own feeds
-- rationale: users can modify their feed names or urls
create policy "authenticated users can update their own feeds"
on feeds for update
to authenticated
using (auth.uid() = user_id)
with check (auth.uid() = user_id);

-- rls policy: deny anonymous users from updating feeds
-- rationale: only authenticated users can modify their feeds
create policy "anonymous users cannot update feeds"
on feeds for update
to anon
using (false);

-- rls policy: allow authenticated users to delete their own feeds
-- rationale: users can remove feeds they no longer want to follow
-- note: cascade deletion will automatically remove associated articles
create policy "authenticated users can delete their own feeds"
on feeds for delete
to authenticated
using (auth.uid() = user_id);

-- rls policy: deny anonymous users from deleting feeds
-- rationale: only authenticated users can delete their feeds
create policy "anonymous users cannot delete feeds"
on feeds for delete
to anon
using (false);

-- add comment to table
comment on table feeds is 'stores rss feed sources added by users';

-- add comments to columns
comment on column feeds.id is 'unique identifier for the feed';
comment on column feeds.user_id is 'reference to the user who owns this feed';
comment on column feeds.name is 'user-defined name for the feed';
comment on column feeds.url is 'rss feed url';
comment on column feeds.last_fetch_status is 'status of the last fetch attempt (e.g., success, error)';
comment on column feeds.last_fetch_error is 'error message from the last failed fetch attempt';
comment on column feeds.created_at is 'timestamp when the feed was created';
comment on column feeds.updated_at is 'timestamp when the feed was last updated';
