-- migration: create_events_table
-- description: creates the events table to store internal product events for metrics tracking
-- tables affected: events
-- special notes: includes rls policies; designed for append-only operations with minimal access for regular users

-- create the events table
create table events (
    id uuid primary key default gen_random_uuid(),
    user_id uuid null references auth.users(id) on delete set null,
    event_type text not null,
    metadata jsonb null,
    created_at timestamptz not null default now()
);

-- create index on user_id for efficient filtering by user
-- nullable to support system events without user association
create index idx_events_user_id on events(user_id);

-- create index on event_type for efficient analytics queries filtering by type
-- e.g., "how many feed_added events occurred this week?"
create index idx_events_event_type on events(event_type);

-- create index on created_at for time-based analytics queries
-- sorted descending for optimal performance when fetching recent events first
-- e.g., "weekly engagement metrics" or "events from last 30 days"
create index idx_events_created_at on events(created_at desc);

-- enable row level security
alter table events enable row level security;

-- note: no select policy for regular users (authenticated or anon)
-- rationale: events are for internal analytics only; users don't need to query their own events
-- analytics dashboards and reports use service role to bypass rls

-- rls policy: allow authenticated users to insert events associated with themselves
-- rationale: enables client-side event tracking for user actions
-- allows null user_id for system events triggered by authenticated users
create policy "authenticated users can insert their own events"
on events for insert
to authenticated
with check (auth.uid() = user_id or user_id is null);

-- rls policy: allow anonymous users to insert system events only
-- rationale: enables tracking of anonymous user interactions (e.g., page views, signups)
-- user_id must be null for anonymous events
create policy "anonymous users can insert anonymous events"
on events for insert
to anon
with check (user_id is null);

-- note: no update or delete policies for any users
-- events are immutable once created for audit trail integrity
-- only service role (background jobs) can modify events if absolutely necessary

-- add comment to table
comment on table events is 'stores internal product events for metrics tracking and analytics';

-- add comments to columns
comment on column events.id is 'unique identifier for the event';
comment on column events.user_id is 'reference to the user (nullable for system events)';
comment on column events.event_type is 'type of event (e.g., user_registered, feed_added, summary_generated)';
comment on column events.metadata is 'additional event-specific data stored as json';
comment on column events.created_at is 'when the event occurred';
