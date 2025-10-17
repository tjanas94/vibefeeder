# Refactoring Plan: FeedFetcherService

## Problem Statement

`FeedFetcherService` currently violates the Single Responsibility Principle (SRP) by handling multiple concerns:

1. Harmonogramowanie i orkiestracja (scheduling + batch processing)
2. Zarządzanie współbieżnością (worker pool, semaphores, goroutines)
3. Rate limiting per domain
4. Komunikacja HTTP i obsługa redirectów
5. Parsowanie feedów (RSS/Atom)
6. Logika biznesowa (zapis artykułów, update statusów)
7. Kalkulacje i strategie retry/backoff
8. Graceful shutdown

**Current state:** ~700 linii kodu w `service.go`

## Target Architecture

```
FeedFetcherService (orchestrator)
  ├── Scheduler (harmonogramowanie + batch processing)
  ├── WorkerPool (concurrency management)
  ├── RateLimiter (domain rate limiting)
  ├── FeedFetcher (HTTP + parsing)
  ├── HTTPResponseHandler (status codes logic)
  ├── FeedStatusManager (update status + save articles)
  └── Repository (data access)
```

## File Structure

```
internal/fetcher/
├── calculations.go          # PURE FUNCTIONS - timing/backoff calculations only
├── rate_limiter.go          # NEW - Rate limiting per domain
├── worker_pool.go           # NEW - Concurrency management
├── scheduler.go             # NEW - Harmonogramowanie i batch processing
├── feed_fetcher.go          # NEW - HTTP requests + feed parsing + transformFeedItems()
├── http_response_handler.go # NEW - Obsługa HTTP status codes
├── feed_status_manager.go   # NEW - Update statusów + zapis artykułów
├── models.go                # NEW - Shared types (Article, FetchDecision)
├── service.go               # REFACTOR - Slim orchestrator
├── repository.go            # NO CHANGES
└── http_client.go           # NO CHANGES
```

## Component Responsibilities

### 1. RateLimiter (`rate_limiter.go`)

**Responsibility:** Zapewnia rate limiting per domain

**Key Methods:**

- `NewRateLimiter(domainDelay time.Duration) *RateLimiter`
- `WaitIfNeeded(feedURL string)` - blokuje jeśli zbyt szybko dla danej domeny

**State:**

- `domainLastRequest map[string]time.Time`
- `mu sync.Mutex`
- `domainDelay time.Duration`

**Dependencies:** None (pure Go)

---

### 2. WorkerPool (`worker_pool.go`)

**Responsibility:** Zarządzanie współbieżnością z limitem workerów

**Key Methods:**

- `NewWorkerPool(workerCount int, logger *slog.Logger) *WorkerPool`
- `Process(ctx context.Context, items []interface{}, processFn func(interface{}))`

**Features:**

- Semaphore dla limitu workerów
- WaitGroup dla synchronizacji
- Panic recovery per worker
- Context cancellation support

**Dependencies:**

- `log/slog` (logging)

---

### 3. HTTPResponseHandler (`http_response_handler.go`)

**Responsibility:** Obsługa różnych HTTP status codes i decyzja o dalszych akcjach

**Key Types:**

```go
type FetchDecision struct {
    ShouldRetry      bool
    NextFetchTime    time.Time
    Status           string
    ErrorMessage     *string
    ETag             *string
    LastModified     *string
    NewURL           *string
    Articles         []Article
}
```

**Key Methods:**

- `NewHTTPResponseHandler(logger, successInterval) *HTTPResponseHandler`
- `HandleResponse(resp *http.Response, feed, feedData, retryCount) FetchDecision`
- Private handlers:
  - `handleSuccess()` - 200 OK
  - `handleNotModified()` - 304
  - `handlePermanentRedirect()` - 301/308
  - `handleTemporaryRedirect()` - 302/307
  - `handleUnauthorized()` - 401
  - `handleTooManyRequests()` - 429
  - `handleNotFound()` - 404/410
  - `handleClientError()` - 4xx
  - `handleServerError()` - 5xx

**Dependencies:**

- `log/slog`
- Pure functions z `calculations.go` (calculateNextFetch, calculateBackoff, parseRetryAfter)

---

### 4. FeedFetcher (`feed_fetcher.go`)

**Responsibility:** Wykonywanie HTTP requestów i parsowanie feedów

**Key Types:**

```go
type FeedData struct {
    Articles []Article
    Title    string
}
```

**Key Methods:**

- `NewFeedFetcher(httpClient, responseHandler, logger) *FeedFetcher`
- `Fetch(ctx, feed, retryCount) (FetchDecision, error)` - główna metoda
- `createRequest(ctx, feed) (*http.Request, error)` - tworzy request z conditional headers
- `parseFeed(body io.Reader) (*FeedData, error)` - parsuje RSS/Atom

**Package-level Functions:**

- `transformFeedItems(items, now) []Article` - transforms gofeed.Item to Article

**Dependencies:**

- `HTTPClient` (http_client.go)
- `HTTPResponseHandler`
- `github.com/mmcdole/gofeed`
- `log/slog`

---

### 5. FeedStatusManager (`feed_status_manager.go`)

**Responsibility:** Aktualizacja statusów feedów i zapis artykułów

**Key Methods:**

- `NewFeedStatusManager(repo, logger) *FeedStatusManager`
- `ApplyDecision(ctx, feed, decision) error` - główna metoda
- `saveArticles(ctx, feedID, articles) error` - private helper

**Logic:**

1. Zapisuje artykuły jeśli są (non-blocking - nie failuje całej operacji)
2. Aktualizuje status feeda w DB
3. Resetuje retry_count jeśli sukces

**Dependencies:**

- `Repository`
- `log/slog`

---

### 6. Scheduler (`scheduler.go`)

**Responsibility:** Harmonogramowanie i orkiestracja batch processingu

**Key Methods:**

- `NewScheduler(repo, workerPool, rateLimiter, feedFetcher, statusManager, logger, config, appCtx) *Scheduler`
- `Start()` - główna pętla z tickerem
- `ProcessBatch()` - pobiera i przetwarza batch feedów
- `processSingleFeed(feed)` - przetwarza pojedynczy feed
- `handleFetchError(feed, err)` - obsługa błędów HTTP/network

**Flow:**

1. Ticker co X minut
2. FindFeedsDueForFetch()
3. WorkerPool.Process() dla wszystkich feedów
4. Dla każdego feeda:
   - RateLimiter.WaitIfNeeded()
   - FeedFetcher.Fetch()
   - FeedStatusManager.ApplyDecision()

**Dependencies:**

- `Repository`
- `WorkerPool`
- `RateLimiter`
- `FeedFetcher`
- `FeedStatusManager`
- `log/slog`
- `config.FetcherConfig`

---

### 7. FeedFetcherService (`service.go`)

**Responsibility:** Główny orchestrator, dependency injection, public API

**Key Methods:**

- `NewFeedFetcherService(dbClient, logger, cfg, appCtx) *FeedFetcherService` - setup wszystkich komponentów
- `Start()` - deleguje do Scheduler.Start()
- `FetchFeedNow(feedID string)` - bezpośrednie użycie FeedFetcher (async)

**State:**

```go
type FeedFetcherService struct {
    scheduler     *Scheduler
    feedFetcher   *FeedFetcher
    statusManager *FeedStatusManager
    repo          *Repository
    logger        *slog.Logger
    appCtx        context.Context
}
```

**Dependencies:** Wszystkie komponenty (tworzy je w konstruktorze)

---

### 8. Models (`models.go`)

**Responsibility:** Współdzielone typy używane przez wiele komponentów

**Types:**

```go
type Article struct {
    Title       string
    URL         string
    Content     *string
    PublishedAt time.Time
}

type FetchDecision struct {
    ShouldRetry      bool
    NextFetchTime    time.Time
    Status           string
    ErrorMessage     *string
    ETag             *string
    LastModified     *string
    NewURL           *string
    Articles         []Article
}
```

---

### 9. Calculations (`calculations.go`)

**Responsibility:** Pure mathematical/timing calculations for feed fetching

**Package-level Pure Functions:**

```go
// Timing calculations
func calculateNextFetch(cacheControl string, successInterval time.Duration, now time.Time) time.Time
func parseCacheControlMaxAge(cacheControl string) (time.Duration, bool)

// Retry strategy
func calculateBackoff(retryCount int, now time.Time) time.Time
func parseRetryAfter(retryAfter string, retryCount int, now time.Time) time.Time
```

**Note:** `transformFeedItems()` has been **moved to `feed_fetcher.go`** as it's a transformation/mapping function, not a calculation.

**Dependencies:** None (pure Go stdlib)

---

## Migration Steps

### Phase 1: Extraction (without breaking changes)

1. ✅ Create `models.go` - extract Article type
2. ✅ Create `rate_limiter.go` - extract rate limiting logic
3. ✅ Create `worker_pool.go` - extract concurrency management
4. ✅ Create `http_response_handler.go` - extract HTTP response handling
5. ✅ Create `feed_fetcher.go` - extract fetching + parsing
6. ✅ Create `feed_status_manager.go` - extract status updates
7. ✅ Create `scheduler.go` - extract scheduling logic

### Phase 2: Integration

8. ✅ Refactor `service.go` to use new components
9. ✅ Update tests to work with new structure
10. ✅ Verify all functionality works

### Phase 3: Cleanup

11. ✅ Remove old code from `service.go`
12. ✅ Update documentation
13. ✅ Run linters and formatters

## Testing Strategy

### Unit Tests

Each component should be independently testable:

- **RateLimiter**: Test with mock time, verify delays
- **WorkerPool**: Test concurrency limits, panic recovery, context cancellation
- **HTTPResponseHandler**: Test all status codes, verify correct decisions
- **FeedFetcher**: Mock HTTPClient, test parsing logic
- **FeedStatusManager**: Mock Repository, verify DB updates
- **Scheduler**: Mock all dependencies, test orchestration

### Integration Tests

- Test full flow: Scheduler → WorkerPool → RateLimiter → FeedFetcher → StatusManager
- Test graceful shutdown
- Test error scenarios

## Benefits

✅ **Single Responsibility** - każdy komponent ma jedną odpowiedzialność  
✅ **Testability** - małe, izolowane komponenty z mock'owalnymi dependencies  
✅ **Maintainability** - zmiany w jednym miejscu nie wpływają na inne  
✅ **Readability** - pliki po ~70-150 linii zamiast 700  
✅ **Extensibility** - łatwe dodawanie nowych features (np. inne strategie retry)  
✅ **Pure functions preserved** - `calculations.go` pozostaje bez zmian  
✅ **Clear dependencies** - każdy komponent ma jasno zdefiniowane zależności

## Notes

### Pure Functions Strategy

- **calculations.go** - tylko czyste kalkulacje (timing, backoff, HTTP header parsing) - NIE ROBIMY ICH OBIEKTOWO
- **transformFeedItems()** - przeniesione z `calculations.go` do `feed_fetcher.go` jako **package-level pure function (bez receivera)**
  - Powód: semantycznie to transformacja/mapping, nie calculation
  - Pozostaje pure function dla łatwego testowania
  - Blisko miejsca użycia (parseFeed)
- Zachowujemy functional programming style tam gdzie ma sens
- Zachowujemy istniejące public API dla kompatybilności wstecznej
- Graceful shutdown musi działać we wszystkich komponentach
- Rate limiting per domain jest krytyczny - nie gubić tego w refactoringu
- Panic recovery w workerach jest kluczowy dla stability
- SSRF protection pozostaje w HTTPClient (http_client.go) w transport layer
- Error handling (SSRF vs network errors) będzie w FeedStatusManager
