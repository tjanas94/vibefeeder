-- migration: create_updated_at_trigger
-- description: creates a reusable trigger function to automatically update updated_at timestamps
-- tables affected: feeds (with potential for reuse in future tables)
-- special notes: this function can be applied to any table with an updated_at column

-- create a generic function to update the updated_at timestamp
-- this function will be called by triggers on tables that need automatic timestamp updates
create or replace function update_updated_at_column()
returns trigger as $$
begin
    -- set the updated_at column to the current timestamp
    -- this ensures the column always reflects the last modification time
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

-- add comment to function
comment on function update_updated_at_column() is 'generic trigger function to automatically update updated_at column on row updates';

-- create trigger on feeds table to automatically update updated_at
-- fires before each update operation on the feeds table
-- ensures updated_at is always current without requiring application-level logic
create trigger set_updated_at
    before update on feeds
    for each row
    execute function update_updated_at_column();

-- add comment to trigger
comment on trigger set_updated_at on feeds is 'automatically updates updated_at timestamp whenever a feed row is modified';
