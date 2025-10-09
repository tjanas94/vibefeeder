# VibeFeeder - Database Schema Plan

## 1. Tables

### 1.1 feeds

Stores RSS feed sources added by users.

| Column              | Type          | Constraints                                            | Description                                                 |
| ------------------- | ------------- | ------------------------------------------------------ | ----------------------------------------------------------- |
| `id`                | `UUID`        | `PRIMARY KEY DEFAULT gen_random_uuid()`                | Unique identifier for the feed                              |
| `user_id`           | `UUID`        | `NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE` | Reference to the user who owns this feed                    |
| `name`              | `TEXT`        | `NOT NULL`                                             | User-defined name for the feed                              |
| `url`               | `TEXT`        | `NOT NULL`                                             | RSS feed URL                                                |
| `last_fetch_status` | `TEXT`        | `NULL`                                                 | Status of the last fetch attempt (e.g., 'success', 'error') |
| `last_fetch_error`  | `TEXT`        | `NULL`                                                 | Error message from the last failed fetch attempt            |
| `created_at`        | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()`                               | Timestamp when the feed was created                         |
| `updated_at`        | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()`                               | Timestamp when the feed was last updated                    |

**Constraints:**

- `UNIQUE(user_id, url)` - Prevents duplicate feeds for the same user

---

### 1.2 articles

Stores articles fetched from RSS feeds.

| Column         | Type          | Constraints                                       | Description                       |
| -------------- | ------------- | ------------------------------------------------- | --------------------------------- |
| `id`           | `UUID`        | `PRIMARY KEY DEFAULT gen_random_uuid()`           | Unique identifier for the article |
| `feed_id`      | `UUID`        | `NOT NULL REFERENCES feeds(id) ON DELETE CASCADE` | Reference to the source feed      |
| `title`        | `TEXT`        | `NOT NULL`                                        | Article title                     |
| `url`          | `TEXT`        | `NOT NULL`                                        | Article URL                       |
| `content`      | `TEXT`        | `NULL`                                            | Article content/description       |
| `published_at` | `TIMESTAMPTZ` | `NOT NULL`                                        | When the article was published    |
| `created_at`   | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()`                          | When the article was fetched      |

**Constraints:**

- `UNIQUE(feed_id, url)` - Prevents duplicate articles from the same feed

---

### 1.3 summaries

Stores AI-generated summaries.

| Column       | Type          | Constraints                                            | Description                                      |
| ------------ | ------------- | ------------------------------------------------------ | ------------------------------------------------ |
| `id`         | `UUID`        | `PRIMARY KEY DEFAULT gen_random_uuid()`                | Unique identifier for the summary                |
| `user_id`    | `UUID`        | `NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE` | Reference to the user who generated this summary |
| `content`    | `TEXT`        | `NOT NULL`                                             | The generated summary text                       |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()`                               | When the summary was generated                   |

---

### 1.4 events

Stores internal product events for metrics tracking.

| Column       | Type          | Constraints                                         | Description                                                                              |
| ------------ | ------------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `id`         | `UUID`        | `PRIMARY KEY DEFAULT gen_random_uuid()`             | Unique identifier for the event                                                          |
| `user_id`    | `UUID`        | `NULL REFERENCES auth.users(id) ON DELETE SET NULL` | Reference to the user (nullable for system events)                                       |
| `event_type` | `TEXT`        | `NOT NULL`                                          | Type of event (e.g., 'user_registered', 'user_login', 'feed_added', 'summary_generated') |
| `metadata`   | `JSONB`       | `NULL`                                              | Additional event-specific data                                                           |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()`                            | When the event occurred                                                                  |

---

## 2. Relationships

### 2.1 Entity Relationship Diagram (Text Format)

```
auth.users (Supabase built-in)
    ║
    ║ (1:N)
    ╠══════════════════> feeds
    ║                      ║
    ║                      ║ (1:N)
    ║                      ╚═══════> articles
    ║
    ║ (1:N)
    ╠══════════════════> summaries
    ║
    ║ (1:N)
    ╚══════════════════> events
```

### 2.2 Relationship Details

1. **auth.users → feeds** (One-to-Many)
   - One user can have multiple feeds
   - Each feed belongs to exactly one user
   - Cascade deletion: deleting a user deletes all their feeds

2. **feeds → articles** (One-to-Many)
   - One feed can have multiple articles
   - Each article belongs to exactly one feed
   - Cascade deletion: deleting a feed deletes all its articles

3. **auth.users → summaries** (One-to-Many)
   - One user can generate multiple summaries
   - Each summary belongs to exactly one user
   - Cascade deletion: deleting a user deletes all their summaries

4. **auth.users → events** (One-to-Many, nullable)
   - One user can generate multiple events
   - Events can exist without a user (system events)
   - Set null on deletion: deleting a user sets `user_id` to NULL in their events

---

## 3. Indexes

### 3.1 Performance Indexes

```sql
-- Foreign key indexes for join performance
CREATE INDEX idx_feeds_user_id ON feeds(user_id);
CREATE INDEX idx_articles_feed_id ON articles(feed_id);
CREATE INDEX idx_summaries_user_id ON summaries(user_id);
CREATE INDEX idx_events_user_id ON events(user_id);

-- Query optimization indexes
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_created_at ON events(created_at DESC);
CREATE INDEX idx_summaries_created_at ON summaries(created_at DESC);
```

### 3.2 Index Justifications

- **Foreign key indexes**: Optimize joins and cascade operations
- **idx_articles_published_at**: Enables efficient filtering for last 24 hours of articles during summary generation
- **idx_events_event_type**: Speeds up event analytics queries filtering by type
- **idx_events_created_at**: Optimizes time-based event analytics (e.g., weekly engagement metrics)
- **idx_summaries_created_at**: Enables fast retrieval of the most recent summary per user

---

## 4. Row-Level Security (RLS) Policies

All tables containing user data must have RLS enabled to ensure multi-tenancy isolation.

### 4.1 feeds Table Policies

```sql
-- Enable RLS
ALTER TABLE feeds ENABLE ROW LEVEL SECURITY;

-- Allow users to view only their own feeds
CREATE POLICY "Users can view their own feeds"
ON feeds FOR SELECT
USING (auth.uid() = user_id);

-- Allow users to insert their own feeds
CREATE POLICY "Users can insert their own feeds"
ON feeds FOR INSERT
WITH CHECK (auth.uid() = user_id);

-- Allow users to update their own feeds
CREATE POLICY "Users can update their own feeds"
ON feeds FOR UPDATE
USING (auth.uid() = user_id)
WITH CHECK (auth.uid() = user_id);

-- Allow users to delete their own feeds
CREATE POLICY "Users can delete their own feeds"
ON feeds FOR DELETE
USING (auth.uid() = user_id);
```

### 4.2 articles Table Policies

```sql
-- Enable RLS
ALTER TABLE articles ENABLE ROW LEVEL SECURITY;

-- Allow users to view articles from their feeds
CREATE POLICY "Users can view articles from their feeds"
ON articles FOR SELECT
USING (
    EXISTS (
        SELECT 1 FROM feeds
        WHERE feeds.id = articles.feed_id
        AND feeds.user_id = auth.uid()
    )
);

-- Allow system/background jobs to insert articles (service role bypass)
-- No INSERT policy for regular users - articles are created by background jobs
```

### 4.3 summaries Table Policies

```sql
-- Enable RLS
ALTER TABLE summaries ENABLE ROW LEVEL SECURITY;

-- Allow users to view only their own summaries
CREATE POLICY "Users can view their own summaries"
ON summaries FOR SELECT
USING (auth.uid() = user_id);

-- Allow users to insert their own summaries
CREATE POLICY "Users can insert their own summaries"
ON summaries FOR INSERT
WITH CHECK (auth.uid() = user_id);

-- No UPDATE or DELETE policies - summaries are immutable once created
```

### 4.4 events Table Policies

```sql
-- Enable RLS
ALTER TABLE events ENABLE ROW LEVEL SECURITY;

-- No SELECT policy for regular users - events are for internal analytics only
-- System/background jobs use service role to bypass RLS for event insertion

-- Allow authenticated users to insert events associated with themselves
CREATE POLICY "Users can insert their own events"
ON events FOR INSERT
WITH CHECK (auth.uid() = user_id OR user_id IS NULL);
```

---

## 5. Design Notes and Decisions

### 5.1 UUID v4 for Primary Keys

- All tables use `UUID` (v4) as primary keys via `gen_random_uuid()`
- Provides globally unique identifiers without coordination
- Better for distributed systems and prevents enumeration attacks
- Slightly larger storage footprint than `BIGINT`, but acceptable for MVP scale

### 5.2 Timestamp Handling

- All timestamps use `TIMESTAMPTZ` to ensure timezone consistency
- `created_at` and `updated_at` patterns for audit trails
- `published_at` in articles table respects original publication time from RSS feeds

### 5.3 Data Integrity

- Cascade deletions ensure referential integrity (user deletion cleans up all dependent data)
- Unique constraints prevent duplicate feeds per user and duplicate articles per feed
- NOT NULL constraints enforce required fields at database level

### 5.4 Multi-Tenancy Security

- RLS policies strictly enforce data isolation between users
- All user-scoped queries automatically filtered by `auth.uid()`
- Background jobs use service role to bypass RLS for system operations

### 5.5 Scalability Considerations

- No retention policies in MVP, but `created_at` indexes support future time-based partitioning
- `events` table designed for append-only operations with minimal indexes
- `articles` table may grow large; future optimization could include partitioning by `published_at`
- JSONB `metadata` in events table allows schema flexibility without migrations

### 5.6 Missing Elements (Intentionally Excluded from MVP)

- No `profiles` table (using `auth.users` directly)
- No article categorization or tagging
- No user preferences/settings table
- No article read/unread tracking
- No feed categories or folders
- No retention policies or archival strategy
