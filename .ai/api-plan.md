# HTML API Plan - VibeFeeder

## 1. Resources

| Resource  | Database Table | Description                                     |
| --------- | -------------- | ----------------------------------------------- |
| Auth      | `auth.users`   | User authentication and session management      |
| Feeds     | `feeds`        | RSS feed sources managed by users               |
| Summaries | `summaries`    | AI-generated content summaries                  |
| Articles  | `articles`     | Fetched RSS articles (not exposed via API)      |
| Events    | `events`       | Internal analytics events (not exposed via API) |

---

## 2. Endpoints

### 2.1 Dashboard

#### GET /dashboard

Main application view with layout, filters, and containers for user/feeds/summary loaded via htmx.

**Query Parameters:**

- `search` (optional): string - Initial search value for feed filters (pre-populates search input)
- `status` (optional): enum - Initial status filter for feeds. Values: `all`, `working`, `pending`, `error`. Default: `all` (pre-selects status dropdown)
- `page` (optional): integer - Initial page number for feed list. Default: `1`

Note: These query parameters are used only to pre-populate the filter form values. The actual feed data is loaded via htmx from `/feeds` endpoint.

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type DashboardViewModel struct {
    Title     string
    UserEmail string
    Query     *feedmodels.ListFeedsQuery // Query params for feed filtering (search, status, page)
}
```

Note: `ListFeedsQuery` contains: `Search`, `Status`, `Page` fields with validation tags and `SetDefaults()` method. Page size is fixed in backend (currently 20).

**Success Response:**

- HTTP 200 OK
- Renders: Complete dashboard page with:
  - Feed filter form with id `#feed-filters` containing:
    - Search input (name="search", pre-populated from `vm.Query.Search`)
    - Status select (name="status", pre-populated from `vm.Query.Status`, default: "all")
    - Form uses `hx-get="/feeds"` with `hx-target="#feed-list-container"` and `hx-trigger="change, submit"`
  - Feed add button/form
  - Empty container `#feed-list-container` with `hx-get="/feeds"` including current filter params from `vm.Query`
  - Empty container `#summary-container` (loaded by `hx-get="/summaries/latest" hx-trigger="load"`)

**Error Responses:**

- 400 Bad Request
  - Renders: Dashboard with error message "Invalid filter parameters"
- 401 Unauthorized
  - Header: `Location: /auth/login`
  - Renders: Redirect to login page
- 500 Internal Server Error
  - Renders: Error page with "Failed to load dashboard. Please refresh the page."

**Side Effects:** None

**htmx Integration:**

- Dashboard renders layout with filter forms pre-populated from `vm.Query` (ListFeedsQuery)
- Feed list container uses `hx-get="/feeds"` with `hx-trigger="load, refreshFeedList from:body"` and `hx-include="#feed-filters"`
- Filter form (#feed-filters) uses `hx-get="/feeds"` to reload feed list on change/submit
- Summary container uses `hx-get="/summaries/latest"` with `hx-trigger="load"`
- All feed mutations (POST, PATCH, DELETE) use `hx-include="#feed-filters"` to preserve current filter state
- After mutations, server can return `HX-Trigger: refreshFeedList` to reload feed list with current filters
- Pagination links include current query params: `/feeds?search={Query.Search}&status={Query.Status}&page={N}`

---

### 2.2 Feeds Management

#### GET /feeds

List all feeds for authenticated user (returns HTML partial for htmx).

**Query Parameters:**

- `search` (optional): string - Case-insensitive search by feed name
- `status` (optional): enum - Filter by feed status. Values: `all`, `working`, `pending`, `error`. Default: `all`
  - `all` - no filter on status
  - `working` - WHERE `last_fetch_status = 'success'`
  - `pending` - WHERE `last_fetch_status IS NULL`
  - `error` - WHERE `last_fetch_status IN ('temporary_error', 'permanent_error', 'unauthorized')`
- `page` (optional): integer - Page number (1-indexed). Default: `1`
  **Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type FeedListViewModel struct {
    Feeds []FeedItemViewModel
    ShowEmptyState bool
    Pagination PaginationViewModel
}

type FeedItemViewModel struct {
    ID string
    Name string
    URL string
    HasError bool
    ErrorMessage string
    LastFetchedAt time.Time
}

type PaginationViewModel struct {
    CurrentPage int
    TotalPages int
    TotalItems int
    HasPrevious bool
    HasNext bool
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Feed list partial HTML (for htmx swap into `#feed-list-container`)

**Error Responses:**

- 400 Bad Request
  - HTTP 400
  - Renders: Error message partial "Invalid query parameters"
- 401 Unauthorized
  - HTTP 401
  - Renders: Login redirect partial or error message
- 500 Internal Server Error
  - HTTP 500
  - Renders: Error message partial "Failed to load feeds"

**Side Effects:** None

**htmx Notes:**

- Called on dashboard load with `hx-trigger="load"`
- Listens for `refreshFeedList` event from body to refresh after mutations
- Called by filter form with `hx-trigger="input changed delay:500ms from:#search, change from:#status"`
- Pagination links use `hx-include="#feed-filters"` to preserve filter state
- Triggered by successful POST /feeds, PATCH /feeds/{id}, and DELETE /feeds/{id} operations

---

#### POST /feeds

Create a new RSS feed and trigger feed list refresh.

**Query Parameters:** None

**Request Formdata:**

```
name: string (required, non-empty)
url: string (required, valid URL format)
```

**View Model (Templ):**

Error only:

```go
type FeedFormErrorViewModel struct {
    NameError string
    URLError string
    GeneralError string
}
```

**Success Response:**

- HTTP 204 No Content
- Header: `HX-Trigger: refreshFeedList`
- Triggers `GET /feeds` on `#feed-list-container` to refresh the list

**Error Responses:**

- 400 Bad Request
  - Headers: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with specific field errors:
    - `NameError = "Feed name is required"`
    - `URLError = "Invalid URL format"`
- 409 Conflict
  - Headers: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with `URLError = "You have already added this feed"`
- 500 Internal Server Error
  - Headers: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with `GeneralError = "Failed to add feed. Please try again."`

**Side Effects:**

- Creates new record in `feeds` table with `fetch_after = NOW() + INTERVAL '5 minutes'`
- Records `feed_added` event in `events` table
- Background job will fetch articles after 5 minutes

**htmx Notes:**

- Form uses `hx-swap="none"` to ignore response body (only cares about HX-Trigger header)
- `#feed-list-container` listens for `refreshFeedList` event and calls `GET /feeds` with current filters
- Clean separation: POST only mutates, GET only renders

---

#### GET /feeds/{id}/edit

Render edit form for specific feed.

**Query Parameters:** None

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type FeedEditFormViewModel struct {
    FeedID string
    Name string
    URL string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Feed edit form partial with current values pre-filled

**Error Responses:**

- 401 Unauthorized
  - HTTP 401
  - Renders: Error message partial
- 404 Not Found
  - HTTP 404
  - Renders: Error message partial "Feed not found"
- 500 Internal Server Error
  - HTTP 500
  - Renders: Error message partial "Failed to load feed"

**Side Effects:** None

---

#### PATCH /feeds/{id}

Update existing feed and trigger feed list refresh.

**Query Parameters:** None

**Request Formdata:**

```
name: string (required, non-empty)
url: string (required, valid URL format)
```

**View Model (Templ):**

Error only:

```go
type FeedFormErrorViewModel struct {
    NameError string
    URLError string
    GeneralError string
}
```

**Success Response:**

- HTTP 204 No Content
- Header: `HX-Trigger: refreshFeedList`
- Triggers `GET /feeds` on `#feed-list-container` to refresh the list

**Error Responses:**

- 400 Bad Request
  - Headers: `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with specific field errors
- 404 Not Found
  - HTTP 404
  - Headers: `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
  - Renders: Error message partial "Feed not found"
- 409 Conflict
  - Headers: `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with `URLError = "A feed with this URL already exists"`
- 500 Internal Server Error
  - Headers: `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
  - Renders: `FeedFormErrorViewModel` with `GeneralError = "Failed to update feed"`

**Side Effects:**

- Updates record in `feeds` table
- If URL changed:
  - Resets `last_fetch_status`, `last_fetch_error`, `last_modified`, `etag` to NULL
  - Sets `fetch_after = NOW() + INTERVAL '5 minutes'`
  - Background job will fetch articles after 5 minutes
- If only name changed:
  - Only updates `name` field

**htmx Notes:**

- Form uses `hx-swap="none"` to ignore response body
- `#feed-list-container` listens for `refreshFeedList` event and calls `GET /feeds` with current filters
- Clean separation: POST only mutates, GET only renders

---

#### DELETE /feeds/{id}

Delete a feed and all associated articles, trigger feed list refresh.

**Query Parameters:** None

**Request Formdata:** None (DELETE request)

**View Model (Templ):**

None (no view model for success, errors render partials)

**Success Response:**

- HTTP 204 No Content
- Header: `HX-Trigger: refreshFeedList`
- Triggers `GET /feeds` on `#feed-list-container` to refresh the list

**Error Responses:**

- 401 Unauthorized
  - HTTP 401
  - Headers: `HX-Retarget: #feed-item-{id}-errors`, `HX-Reswap: innerHTML`
  - Renders: Error message partial
- 404 Not Found
  - HTTP 404
  - Headers: `HX-Retarget: #feed-item-{id}-errors`, `HX-Reswap: innerHTML`
  - Renders: Error message partial "Feed not found"
- 500 Internal Server Error
  - HTTP 500
  - Headers: `HX-Retarget: #feed-item-{id}-errors`, `HX-Reswap: innerHTML`
  - Renders: Error message partial "Failed to delete feed"

**Side Effects:**

- Deletes record from `feeds` table
- Cascading deletion of associated `articles` records (via database FK constraint)

**htmx Notes:**

- Delete button uses `hx-swap="none"` to ignore response body
- `#feed-list-container` listens for `refreshFeedList` event and calls `GET /feeds` with current filters
- Clean separation: DELETE only mutates, GET only renders

---

### 2.3 Summaries

#### GET /summaries/latest

Get the most recent summary for authenticated user.

**Query Parameters:** None

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type SummaryDisplayViewModel struct {
    Summary      *SummaryViewModel
    CanGenerate  bool
    ErrorMessage string
}

type SummaryViewModel struct {
    ID string
    Content string
    CreatedAt time.Time
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Summary display partial with content and timestamp, or empty state with "Generate" button

**Error Responses:**

- 401 Unauthorized
  - HTTP 401
  - Renders: Error message partial
- 500 Internal Server Error
  - HTTP 500
  - Renders: Error message partial "Failed to load summary"

**Side Effects:** None

---

#### POST /summaries

Generate a new AI summary from articles published in the last 24 hours.

**Query Parameters:** None

**Request Formdata:** None (empty POST)

**View Model (Templ):**

```go
type SummaryDisplayViewModel struct {
    Summary      *SummaryViewModel
    CanGenerate  bool
    ErrorMessage string
}

type SummaryErrorViewModel struct {
    ErrorMessage string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: New summary display partial with generated content and timestamp

**Error Responses:**

- 404 Not Found
  - Renders: `SummaryErrorViewModel` with `ErrorMessage = "No articles found from the last 24 hours"`
- 500 Internal Server Error
  - Renders: `SummaryErrorViewModel` with `ErrorMessage = "Failed to generate summary. Please try again later."`
- 503 Service Unavailable
  - Renders: `SummaryErrorViewModel` with `ErrorMessage = "AI service is temporarily unavailable"`

**Side Effects:**

- Queries `articles` table for articles from last 24 hours (using `idx_articles_published_at` index)
- Joins with `feeds` table to get user's feeds
- Calls OpenRouter.ai API for summary generation
- Creates new record in `summaries` table
- Records `summary_generated` event in `events` table

---

## 3. Authentication and Authorization

### Authentication Mechanism

**Provider:** Supabase Auth

**Implementation Details:**

- Server-side session management using Supabase Auth tokens
- Authentication state stored in HTTP-only secure cookies
- Session tokens validated on every request to protected endpoints

**Token Flow:**

1. User submits credentials via `/auth/register` or `/auth/login`
2. Server validates credentials with Supabase Auth API
3. On success, Supabase returns access token and refresh token
4. Server stores tokens in HTTP-only, secure, SameSite cookies
5. Subsequent requests include cookies automatically
6. Server validates token on each protected endpoint request
7. Refresh token used to obtain new access token when expired

**Protected Endpoints:**
All endpoints except:

- `POST /auth/register`
- `POST /auth/login`
- Any static assets

### Authorization

**Row-Level Security (RLS):**

- All database queries automatically filtered by `auth.uid()` via Supabase RLS policies
- Users can only access their own feeds, articles (via feeds), and summaries
- No additional authorization checks needed in application code beyond authentication

**Supabase Service Role:**

- Background jobs use service role to bypass RLS when fetching articles for all users
- Service role key stored securely in environment variables
- Never exposed to client or in API responses

**CSRF Protection:**

- Echo CSRF middleware enabled for all state-changing operations (POST, PUT, DELETE)
- CSRF tokens embedded in forms and validated on submission
- For htmx requests, CSRF token included in custom header `X-CSRF-Token`

---

## 4. Validation and Business Logic

### 4.1 Validation Rules by Resource

#### Feeds

**Field: name**

- Type: string
- Required: Yes
- Constraints:
  - Non-empty after trimming whitespace
  - Maximum length: 255 characters
- Error messages:
  - Empty: "Feed name is required"
  - Too long: "Feed name must be less than 255 characters"

**Field: url**

- Type: string (URL)
- Required: Yes
- Constraints:
  - Valid URL format (must start with http:// or https://)
  - Unique per user (enforced by database constraint)
- Error messages:
  - Empty: "Feed URL is required"
  - Invalid format: "Invalid URL format. Must start with http:// or https://"
  - Duplicate: "You have already added this feed"

**Validation Process:**

1. Trim whitespace from name and url
2. Check required fields are non-empty
3. Validate URL format using regex (must start with http:// or https://)
4. Check for duplicate URL in user's feeds
5. If all pass, accept feed (actual feed validation happens in background job)

#### Summaries

**Business Rules:**

- Must have at least one article published in the last 24 hours
- Summary generation timeout: 60 seconds
- If AI API fails, retry once before returning error

**Validation Process:**

1. Query articles from last 24 hours across all user feeds
2. If no articles found, return error "No articles found from the last 24 hours"
3. Prepare articles for AI API (format: title, content, published date)
4. Call OpenRouter AI API with timeout
5. If timeout or error, retry once
6. If still fails, return error "Failed to generate summary. Please try again later."
7. Save summary to database
8. Return rendered summary

### 4.2 Business Logic Implementation

#### User Registration Flow

1. Validate form input (email format, password match, privacy acceptance)
2. Call Supabase Auth API: `auth.signUp({ email, password })`
3. Handle Supabase response:
   - Success: proceed to step 4
   - Error "User already registered": return 409 with error message
   - Other error: return 500 with generic error message
4. Create session cookies with returned tokens
5. Record `user_registered` event: `INSERT INTO events (user_id, event_type) VALUES (?, 'user_registered')`
6. Redirect to `/dashboard` with empty state (no feeds)

#### User Login Flow

1. Validate form input (email and password non-empty)
2. Call Supabase Auth API: `auth.signInWithPassword({ email, password })`
3. Handle Supabase response:
   - Success: proceed to step 4
   - Error "Invalid credentials": return 401 with error message "Invalid email or password"
   - Other error: return 500 with generic error message
4. Create session cookies with returned tokens
5. Record `user_login` event: `INSERT INTO events (user_id, event_type) VALUES (?, 'user_login')`
6. Redirect to `/dashboard`

#### Feed Addition Flow

1. Validate form input (name non-empty, url valid format)
2. Check for duplicate: `SELECT COUNT(*) FROM feeds WHERE user_id = ? AND url = ?`
   - If exists, return 409 with error "You have already added this feed"
3. Insert feed: `INSERT INTO feeds (user_id, name, url, fetch_after) VALUES (?, ?, ?, NOW() + INTERVAL '5 minutes')`
4. Record `feed_added` event: `INSERT INTO events (user_id, event_type, metadata) VALUES (?, 'feed_added', '{"feed_id": "..."}')`
5. Trigger immediate fetch: Call `FetchFeedNow(feed_id)` to start fetching in background immediately
6. Return 204 No Content with HX-Trigger header to refresh feed list

#### Feed Update Flow

1. Validate ownership: check feed belongs to authenticated user (RLS handles this)
2. Validate form input (name non-empty, url valid format)
3. Check for duplicate if URL changed: `SELECT COUNT(*) FROM feeds WHERE user_id = ? AND url = ? AND id != ?`
   - If exists, return 409 with error "A feed with this URL already exists"
4. Update feed:
   - If URL changed:

     ```sql
     UPDATE feeds SET
       name = ?,
       url = ?,
       last_fetch_status = NULL,
       last_fetch_error = NULL,
       last_modified = NULL,
       etag = NULL,
       fetch_after = NOW() + INTERVAL '5 minutes',
       retry_count = 0
     WHERE id = ?
     ```

     - Trigger immediate fetch: Call `FetchFeedNow(feed_id)` to start fetching in background immediately

   - If only name changed:
     ```sql
     UPDATE feeds SET name = ? WHERE id = ?
     ```

5. Return 204 No Content with HX-Trigger header to refresh feed list

#### Feed Deletion Flow

1. Validate ownership: check feed belongs to authenticated user (RLS handles this)
2. Delete feed: `DELETE FROM feeds WHERE id = ?`
   - Cascading deletion removes associated articles automatically
3. Return success message or trigger feed list refresh

#### Summary Generation Flow

1. Query articles from last 24 hours:
   ```sql
   SELECT a.id, a.title, a.content, a.url, a.published_at, f.name AS feed_name
   FROM articles a
   JOIN feeds f ON a.feed_id = f.id
   WHERE f.user_id = ? AND a.published_at >= NOW() - INTERVAL '24 hours'
   ORDER BY a.published_at DESC
   LIMIT 100
   ```
2. If no articles, return 404 with error "No articles found from the last 24 hours"
3. Prepare prompt for AI:

   ```
   Generate a concise summary of the following articles published in the last 24 hours.
   The summary should be 3-5 paragraphs and highlight the most important information and trends.

   Articles:
   [for each article: Title, Source, Published Date, Content excerpt]
   ```

4. Call OpenRouter API:
   - Model: configurable (default: `openai/gpt-4o-mini` for cost efficiency)
   - Max tokens: 1000
   - Temperature: 0.3 (for consistency)
   - Timeout: 60 seconds
5. If API call fails, retry once after 2-second delay
6. If still fails, return 503 with error
7. Save summary: `INSERT INTO summaries (user_id, content) VALUES (?, ?)`
8. Record event: `INSERT INTO events (user_id, event_type, metadata) VALUES (?, 'summary_generated', '{"article_count": N}')`
9. Return summary display HTML partial with new summary

#### Article Fetch Background Job

**Architecture:** Single process with goroutine worker pool using database polling

**Configuration:**

- Worker pool size: 10 workers (configurable)
- Fetch interval: Every 5 minutes

**Trigger:** Runs every 5 minutes

**Queueing Mechanism:**

- `fetch_after` timestamp acts as natural queue mechanism for scheduled fetching
- New/updated feeds created with `fetch_after = NOW() + 5 minutes` (fallback for regular cron cycle)
- HTTP handlers trigger immediate fetch via `FetchFeedNow(feedID)` after INSERT/UPDATE
- Background job polls database for feeds WHERE `fetch_after <= NOW()` (scheduled refresh)

**Feed Status Values:**

- `success` - Feed fetched successfully, schedule next fetch in 1 hour (`fetch_after = NOW() + 1 hour`)
- `temporary_error` - Transient failure (timeout, 5xx), retry with exponential backoff (respect `fetch_after`)
- `permanent_error` - Unrecoverable error (404, 410, 200 with invalid content-type), never fetch again (`fetch_after = NULL`)
- `unauthorized` - Authentication required (401, 403), never fetch again (`fetch_after = NULL`)

**Process:**

1. Get feeds ready for fetching (ordered by priority):

   ```sql
   SELECT id, url, last_modified, etag, last_fetch_status, retry_count
   FROM feeds
   WHERE last_fetch_status NOT IN ('permanent_error', 'unauthorized')
     AND (fetch_after IS NULL OR fetch_after <= NOW())
   ORDER BY last_fetched_at NULLS FIRST
   ```

2. For each feed:
   - **Prepare HTTP request** with conditional headers:
     - If `last_modified` exists: add `If-Modified-Since` header
     - If `etag` exists: add `If-None-Match` header
   - **Fetch feed URL** with 30-second timeout
   - **Handle HTTP responses:**
     - **304 Not Modified**:
       - Update feed on success (continue to step 4)
     - **200 OK**:
       - Validate `Content-Type` (must be `application/rss+xml`, `application/xml`, `text/xml`, or contain `xml`)
       - If invalid content-type → set status to `permanent_error`
       - Parse XML and validate RSS structure
       - If invalid XML → set status to `permanent_error`
       - Extract articles (continue to step 3)
     - **301 Moved Permanently / 308 Permanent Redirect**:
       - Follow redirect (up to 5 redirects max)
       - Update feed URL in database to new location
       - Process response from new URL
     - **302 Found / 307 Temporary Redirect**:
       - Follow redirect (up to 5 redirects max)
       - Do NOT update feed URL in database
       - Process response from redirected URL
     - **400 Bad Request**:
       ```sql
       UPDATE feeds SET
         last_fetch_status = 'permanent_error',
         last_fetch_error = 'Invalid request',
         last_fetched_at = NOW(),
         fetch_after = NULL
       WHERE id = ?
       ```
     - **401 Unauthorized / 403 Forbidden**:
       ```sql
       UPDATE feeds SET
         last_fetch_status = 'unauthorized',
         last_fetch_error = 'Authentication required',
         last_fetched_at = NOW(),
         fetch_after = NULL
       WHERE id = ?
       ```
     - **404 Not Found / 410 Gone**:
       ```sql
       UPDATE feeds SET
         last_fetch_status = 'permanent_error',
         last_fetch_error = 'Feed no longer available',
         last_fetched_at = NOW(),
         fetch_after = NULL
       WHERE id = ?
       ```
     - **429 Too Many Requests**:
       - Read `Retry-After` header (in seconds or HTTP date)
       - Set `fetch_after` to specified time
       - If no `Retry-After` header, use exponential backoff
       ```sql
       UPDATE feeds SET
         last_fetch_status = 'temporary_error',
         last_fetch_error = 'Rate limited',
         last_fetched_at = NOW(),
         fetch_after = ?, -- from Retry-After header or exponential backoff
         retry_count = retry_count + 1
       WHERE id = ?
       ```
     - **5xx Server Errors / Timeouts**:
       - Apply exponential backoff: 5min → 15min → 1h → 6h (based on consecutive failures)
       ```sql
       UPDATE feeds SET
         last_fetch_status = 'temporary_error',
         last_fetch_error = ?,
         last_fetched_at = NOW(),
         fetch_after = NOW() + INTERVAL '...', -- exponential backoff based on retry_count
         retry_count = retry_count + 1
       WHERE id = ?
       ```

3. **For each article** (if feed returned 200 OK with valid XML):
   - Check if exists: `SELECT COUNT(*) FROM articles WHERE feed_id = ? AND url = ?`
   - If not exists: `INSERT INTO articles (feed_id, title, url, content, published_at) VALUES (...)`

4. **Update feed on success**:

   ```sql
   UPDATE feeds SET
     last_fetch_status = 'success',
     last_fetch_error = NULL,
     last_fetched_at = NOW(),
     last_modified = ?, -- from response headers
     etag = ?,          -- from response headers
     fetch_after = NOW() + INTERVAL '...', -- from cache headers or default 1 hour
     retry_count = 0
   WHERE id = ?
   ```

5. Log job completion with stats (feeds processed, 304 responses, new articles, errors by type)

**Polite Bot Features:**

- **Conditional Requests**: Uses `Last-Modified` and `ETag` headers to avoid redundant downloads (HTTP 304)
- **Respect Retry-After**: Honors `Retry-After` header from 429 responses
- **Exponential Backoff**: Applies increasing delays for `temporary_error` feeds based on `retry_count` (5min → 15min → 1h → 6h → 24h)
- **Permanent Error Handling**: Never retries `permanent_error` or `unauthorized` feeds
- **Fetch Scheduling**: Uses `fetch_after` column to prevent premature retries
- **Success Scheduling**: Successful fetches are scheduled 1 hour later
- **Priority Queue**: Fetches never-fetched feeds first (NULLS FIRST), then by oldest `last_fetched_at`
- **Content-Type Validation**: Rejects non-XML responses as permanent errors
- **Domain Rate Limiting**: Minimum 3 seconds delay between requests to same domain

**Error Handling:**

- Individual feed errors don't stop the job
- Errors logged to `feeds.last_fetch_error` for user visibility
- Critical errors (database unavailable) stop job and alert admin
- Network timeouts and temporary errors trigger exponential backoff
- Permanent errors (404, 410, invalid content-type) prevent future fetches
- Authorization errors (401, 403) prevent future fetches

**Resource Limits:**

- Max 10MB response size
- Max 50 articles processed per feed fetch

**SSRF Protection:**

- Validate URLs to prevent connections to private/internal networks
