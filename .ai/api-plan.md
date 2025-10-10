# REST API Plan - VibeFeeder

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

### 2.2 Dashboard

#### GET /dashboard

Main application view showing feeds and latest summary.

**Query Parameters:** None

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type DashboardViewModel struct {
    User UserInfo
    Feeds []FeedItemViewModel
    LatestSummary *SummaryViewModel
    ShowEmptyState bool
}

type UserInfo struct {
    Email string
}

type FeedItemViewModel struct {
    ID string
    Name string
    URL string
    HasError bool
    ErrorMessage string
    UpdatedAt time.Time
}

type SummaryViewModel struct {
    ID string
    Content string
    CreatedAt time.Time
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Complete dashboard page with feed list and summary sections

**Error Responses:**

- 401 Unauthorized
  - Header: `Location: /auth/login`
  - Renders: Redirect to login page
- 500 Internal Server Error
  - Renders: Error page with "Failed to load dashboard. Please refresh the page."

**Side Effects:** None

---

### 2.3 Feeds Management

#### GET /feeds

List all feeds for authenticated user (returns HTML partial).

**Query Parameters:** None

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type FeedListViewModel struct {
    Feeds []FeedItemViewModel
    ShowEmptyState bool
}

type FeedItemViewModel struct {
    ID string
    Name string
    URL string
    HasError bool
    ErrorMessage string
    UpdatedAt time.Time
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Feed list partial HTML (for htmx swap)

**Error Responses:**

- 401 Unauthorized
  - HTTP 401
  - Renders: Login redirect partial or error message
- 500 Internal Server Error
  - HTTP 500
  - Renders: Error message partial "Failed to load feeds"

**Side Effects:** None

---

#### POST /feeds

Create a new RSS feed.

**Query Parameters:** None

**Request Formdata:**

```
name: string (required, non-empty)
url: string (required, valid URL format)
```

**View Model (Templ):**

```go
type FeedListViewModel struct {
    Feeds []FeedItemViewModel
    ShowEmptyState bool
}

type FeedFormErrorViewModel struct {
    NameError string
    URLError string
    GeneralError string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Updated feed list partial HTML with new feed included

**Error Responses:**

- 400 Bad Request
  - Renders: `FeedFormErrorViewModel` with specific field errors:
    - `NameError = "Feed name is required"`
    - `URLError = "Invalid URL format"` or "Unable to fetch feed from this URL"`
- 409 Conflict
  - Renders: `FeedFormErrorViewModel` with `URLError = "You have already added this feed"`
- 500 Internal Server Error
  - Renders: `FeedFormErrorViewModel` with `GeneralError = "Failed to add feed. Please try again."`

**Side Effects:**

- Creates new record in `feeds` table
- Validates URL by attempting to fetch feed metadata
- Records `feed_added` event in `events` table
- Triggers immediate article fetch for new feed (async)

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

#### POST /feeds/{id}

Update existing feed.

**Query Parameters:** None

**Request Formdata:**

```
name: string (required, non-empty)
url: string (required, valid URL format)
```

**View Model (Templ):**

```go
type FeedItemViewModel struct {
    ID string
    Name string
    URL string
    HasError bool
    ErrorMessage string
    UpdatedAt time.Time
}

type FeedFormErrorViewModel struct {
    NameError string
    URLError string
    GeneralError string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Updated feed item partial HTML

**Error Responses:**

- 400 Bad Request
  - Renders: `FeedFormErrorViewModel` with specific field errors
- 404 Not Found
  - HTTP 404
  - Renders: Error message partial "Feed not found"
- 409 Conflict
  - Renders: `FeedFormErrorViewModel` with `URLError = "A feed with this URL already exists"`
- 500 Internal Server Error
  - Renders: `FeedFormErrorViewModel` with `GeneralError = "Failed to update feed"`

**Side Effects:**

- Updates record in `feeds` table
- Validates URL by attempting to fetch feed metadata
- Resets `last_fetch_status` and `last_fetch_error` if URL changed
- Triggers immediate article fetch if URL changed (async)

---

#### DELETE /feeds/{id}

Delete a feed and all associated articles.

**Query Parameters:** None

**Request Formdata:** None (DELETE request)

**View Model (Templ):**

```go
type FeedDeleteSuccessViewModel struct {
    Message string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: Success message partial "Feed deleted successfully" or triggers list refresh

**Error Responses:**

- 401 Unauthorized
  - HTTP 401
  - Renders: Error message partial
- 404 Not Found
  - HTTP 404
  - Renders: Error message partial "Feed not found"
- 500 Internal Server Error
  - HTTP 500
  - Renders: Error message partial "Failed to delete feed"

**Side Effects:**

- Deletes record from `feeds` table
- Cascading deletion of associated `articles` records (via database FK constraint)

---

### 2.4 Summaries

#### GET /summaries/latest

Get the most recent summary for authenticated user.

**Query Parameters:** None

**Request Formdata:** None (GET request)

**View Model (Templ):**

```go
type SummaryDisplayViewModel struct {
    Summary *SummaryViewModel
    ShowEmptyState bool
    CanGenerate bool // true if user has at least one working feed
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
    Summary *SummaryViewModel
    ShowEmptyState bool
    CanGenerate bool
}

type SummaryErrorViewModel struct {
    ErrorMessage string
}
```

**Success Response:**

- HTTP 200 OK
- Renders: New summary display partial with generated content and timestamp

**Error Responses:**

- 400 Bad Request
  - Renders: `SummaryErrorViewModel` with `ErrorMessage = "You must add at least one RSS feed before generating a summary"`
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
  - Must be accessible (HTTP HEAD request returns 2xx)
  - Must return valid RSS feed XML
  - Unique per user (enforced by database constraint)
- Error messages:
  - Empty: "Feed URL is required"
  - Invalid format: "Invalid URL format. Must start with http:// or https://"
  - Inaccessible: "Unable to fetch feed from this URL. Please check the address."
  - Invalid feed: "This URL does not point to a valid RSS feed"
  - Duplicate: "You have already added this feed"

**Validation Process:**

1. Trim whitespace from name and url
2. Check required fields are non-empty
3. Validate URL format using regex
4. Attempt HTTP HEAD request to URL
5. If HEAD succeeds, attempt GET request
6. Parse response as XML and validate RSS structure
7. Check for duplicate URL in user's feeds
8. If all pass, accept feed

#### Summaries

**Business Rules:**

- User must have at least one feed with `last_fetch_status = 'success'`
- Must have at least one article published in the last 24 hours
- Summary generation timeout: 60 seconds
- If AI API fails, retry once before returning error

**Validation Process:**

1. Check user has feeds: `SELECT COUNT(*) FROM feeds WHERE user_id = ? AND last_fetch_status = 'success'`
2. If count = 0, return error "You must add at least one RSS feed before generating a summary"
3. Query articles from last 24 hours across all user feeds
4. If no articles found, return error "No articles found from the last 24 hours"
5. Prepare articles for AI API (format: title, content, published date)
6. Call OpenRouter AI API with timeout
7. If timeout or error, retry once
8. If still fails, return error "Failed to generate summary. Please try again later."
9. Save summary to database
10. Return rendered summary

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
2. Attempt to fetch feed URL:
   - HTTP HEAD request with 10-second timeout
   - If fails, return 400 with error "Unable to fetch feed from this URL"
3. Attempt to parse feed:
   - HTTP GET request with 10-second timeout
   - Parse as XML
   - Validate RSS 2.0 structure
   - If fails, return 400 with error "This URL does not point to a valid RSS feed"
4. Check for duplicate: `SELECT COUNT(*) FROM feeds WHERE user_id = ? AND url = ?`
   - If exists, return 409 with error "You have already added this feed"
5. Insert feed: `INSERT INTO feeds (user_id, name, url, last_fetch_status) VALUES (?, ?, ?, 'pending')`
6. Record `feed_added` event: `INSERT INTO events (user_id, event_type, metadata) VALUES (?, 'feed_added', '{"feed_id": "..."}')`
7. Trigger async article fetch job for this feed
8. Return updated feed list HTML partial

#### Feed Update Flow

1. Validate ownership: check feed belongs to authenticated user (RLS handles this)
2. Validate form input (name non-empty, url valid format)
3. If URL changed, perform same validation as Feed Addition (steps 2-3)
4. Check for duplicate if URL changed: `SELECT COUNT(*) FROM feeds WHERE user_id = ? AND url = ? AND id != ?`
5. Update feed: `UPDATE feeds SET name = ?, url = ?, last_fetch_status = ?, last_fetch_error = NULL WHERE id = ?`
   - Set `last_fetch_status = 'pending'` if URL changed
6. If URL changed, trigger async article fetch job for this feed
7. Return updated feed item HTML partial

#### Feed Deletion Flow

1. Validate ownership: check feed belongs to authenticated user (RLS handles this)
2. Delete feed: `DELETE FROM feeds WHERE id = ?`
   - Cascading deletion removes associated articles automatically
3. Return success message or trigger feed list refresh

#### Summary Generation Flow

1. Check user has working feeds:
   ```sql
   SELECT COUNT(*) FROM feeds
   WHERE user_id = ? AND (last_fetch_status = 'success' OR last_fetch_status IS NULL)
   ```
2. If count = 0, return 400 with error
3. Query articles from last 24 hours:
   ```sql
   SELECT a.id, a.title, a.content, a.url, a.published_at, f.name AS feed_name
   FROM articles a
   JOIN feeds f ON a.feed_id = f.id
   WHERE f.user_id = ? AND a.published_at >= NOW() - INTERVAL '24 hours'
   ORDER BY a.published_at DESC
   LIMIT 100
   ```
4. If no articles, return 404 with error "No articles found from the last 24 hours"
5. Prepare prompt for AI:

   ```
   Generate a concise summary of the following articles published in the last 24 hours.
   The summary should be 3-5 paragraphs and highlight the most important information and trends.

   Articles:
   [for each article: Title, Source, Published Date, Content excerpt]
   ```

6. Call OpenRouter API:
   - Model: configurable (default: `openai/gpt-4o-mini` for cost efficiency)
   - Max tokens: 1000
   - Temperature: 0.3 (for consistency)
   - Timeout: 60 seconds
7. If API call fails, retry once after 2-second delay
8. If still fails, return 503 with error
9. Save summary: `INSERT INTO summaries (user_id, content) VALUES (?, ?)`
10. Record event: `INSERT INTO events (user_id, event_type, metadata) VALUES (?, 'summary_generated', '{"article_count": N}')`
11. Return summary display HTML partial with new summary

#### Article Fetch Background Job

**Trigger:** Runs every 1 hour via cron job

**Process:**

1. Get all feeds: `SELECT id, url FROM feeds WHERE last_fetch_status != 'disabled'`
2. For each feed:
   - Fetch feed URL with 30-second timeout
   - Parse XML and extract articles
   - For each article:
     - Check if exists: `SELECT COUNT(*) FROM articles WHERE feed_id = ? AND url = ?`
     - If not exists: `INSERT INTO articles (feed_id, title, url, content, published_at) VALUES (...)`
   - Update feed status: `UPDATE feeds SET last_fetch_status = 'success', last_fetch_error = NULL, updated_at = NOW() WHERE id = ?`
   - If any error occurs:
     - Update feed status: `UPDATE feeds SET last_fetch_status = 'error', last_fetch_error = ? WHERE id = ?`
3. Log job completion with stats (total feeds processed, new articles added, errors encountered)

**Error Handling:**

- Individual feed errors don't stop the job
- Errors logged to `feeds.last_fetch_error` for user visibility
- Critical errors (database unavailable) stop job and alert admin
