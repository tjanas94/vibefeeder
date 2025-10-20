-- migration: remove_events_rls_policies
-- description: removes rls policies from events table since we're using service key for all event operations
-- tables affected: events
-- special notes: rls remains enabled but no policies exist; only service role can access the table

-- ============================================================================
-- drop existing rls policies from events table
-- ============================================================================

-- drop policy: authenticated users can insert their own events
-- justification: all event insertions will now use service key, not user credentials
drop policy if exists "authenticated users can insert their own events" on events;

-- drop policy: anonymous users can insert anonymous events
-- justification: all event insertions will now use service key, not user credentials
drop policy if exists "anonymous users can insert anonymous events" on events;

-- ============================================================================
-- notes
-- ============================================================================

-- rls remains enabled on the events table for security
-- with no policies defined, only service role can perform any operations
-- this ensures:
-- 1. users cannot query their own events
-- 2. users cannot insert events directly
-- 3. all event operations are controlled server-side via service key
-- 4. maintains audit trail integrity and prevents tampering
