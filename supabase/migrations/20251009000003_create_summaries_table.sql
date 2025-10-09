-- migration: create_summaries_table
-- description: creates the summaries table to store ai-generated summaries
-- tables affected: summaries
-- special notes: includes rls policies; summaries are immutable once created (no update/delete policies)

-- create the summaries table
create table summaries (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users(id) on delete cascade,
    content text not null,
    created_at timestamptz not null default now()
);

-- create index on user_id for efficient joins and filtering
create index idx_summaries_user_id on summaries(user_id);

-- create index on created_at for efficient retrieval of recent summaries
-- sorted descending to optimize "get latest summary" queries
create index idx_summaries_created_at on summaries(created_at desc);

-- enable row level security
alter table summaries enable row level security;

-- rls policy: allow authenticated users to view only their own summaries
-- rationale: ensures data isolation between users
create policy "authenticated users can view their own summaries"
on summaries for select
to authenticated
using (auth.uid() = user_id);

-- rls policy: deny anonymous users from viewing summaries
-- rationale: summaries are private to authenticated users only
create policy "anonymous users cannot view summaries"
on summaries for select
to anon
using (false);

-- rls policy: allow authenticated users to insert their own summaries
-- rationale: users can generate new summaries from their articles
create policy "authenticated users can insert their own summaries"
on summaries for insert
to authenticated
with check (auth.uid() = user_id);

-- rls policy: deny anonymous users from inserting summaries
-- rationale: only authenticated users can generate summaries
create policy "anonymous users cannot insert summaries"
on summaries for insert
to anon
with check (false);

-- note: no update or delete policies for any users
-- summaries are immutable once created to maintain historical record
-- if a user wants a new summary, they generate a new one instead of modifying existing

-- add comment to table
comment on table summaries is 'stores ai-generated summaries of user articles';

-- add comments to columns
comment on column summaries.id is 'unique identifier for the summary';
comment on column summaries.user_id is 'reference to the user who generated this summary';
comment on column summaries.content is 'the generated summary text';
comment on column summaries.created_at is 'when the summary was generated';
