# Plan Refaktoryzacji: FeedFetcherService

## Uzasadnienie Problemu

`FeedFetcherService` obecnie łamie zasadę pojedynczej odpowiedzialności (SRP), zajmując się wieloma zagadnieniami:

1.  Harmonogramowanie i orkiestracja (scheduling + batch processing)
2.  Zarządzanie współbieżnością (worker pool, semafory, goroutines)
3.  Ograniczanie liczby zapytań na domenę (rate limiting)
4.  Komunikacja HTTP i obsługa przekierowań
5.  Parsowanie feedów (RSS/Atom)
6.  Logika biznesowa (zapis artykułów, aktualizacja statusów)
7.  Kalkulacje i strategie ponawiania prób (retry/backoff)
8.  Bezpieczne zamykanie (graceful shutdown)

**Stan obecny:** ~700 linii kodu w `service.go`

## Architektura Docelowa

```
FeedFetcherService (orkiestrator)
  ├── Scheduler (harmonogramowanie + przetwarzanie wsadowe)
  ├── WorkerPool (zarządzanie współbieżnością)
  ├── RateLimiter (ograniczanie zapytań na domenę)
  ├── FeedFetcher (HTTP + parsowanie)
  ├── HTTPResponseHandler (logika kodów statusu)
  ├── FeedStatusManager (aktualizacja statusu + zapis artykułów)
  └── Repository (dostęp do danych)
```

## Struktura Plików

```
internal/fetcher/
├── calculations.go          # CZYSTE FUNKCJE - tylko obliczenia czasu/backoff
├── rate_limiter.go          # NOWY - Ograniczanie zapytań na domenę
├── worker_pool.go           # NOWY - Zarządzanie współbieżnością
├── scheduler.go             # NOWY - Harmonogramowanie i przetwarzanie wsadowe
├── feed_fetcher.go          # NOWY - Zapytania HTTP + parsowanie feedów + transformFeedItems()
├── http_response_handler.go # NOWY - Obsługa kodów statusu HTTP
├── feed_status_manager.go   # NOWY - Aktualizacja statusów + zapis artykułów
├── models.go                # NOWY - Współdzielone typy (Article, FetchDecision)
├── service.go               # REFAKTOR - Odchudzony orkiestrator
├── repository.go            # BEZ ZMIAN
└── http_client.go           # BEZ ZMIAN
```

## Odpowiedzialności Komponentów

### 1. RateLimiter (`rate_limiter.go`)

**Odpowiedzialność:** Zapewnia ograniczanie liczby zapytań na domenę.

**Kluczowe Metody:**

- `NewRateLimiter(domainDelay time.Duration) *RateLimiter`
- `WaitIfNeeded(feedURL string)` - blokuje, jeśli zapytanie jest zbyt szybkie dla danej domeny

**Stan:**

- `domainLastRequest map[string]time.Time`
- `mu sync.Mutex`
- `domainDelay time.Duration`

**Zależności:** Brak (czysty Go)

---

### 2. WorkerPool (`worker_pool.go`)

**Odpowiedzialność:** Zarządzanie współbieżnością z limitem workerów.

**Kluczowe Metody:**

- `NewWorkerPool(workerCount int, logger *slog.Logger) *WorkerPool`
- `Process(ctx context.Context, items []interface{}, processFn func(interface{}))`

**Cechy:**

- Semafor dla limitu workerów
- WaitGroup dla synchronizacji
- Odzyskiwanie po panice dla każdego workera
- Wsparcie dla anulowania kontekstu

**Zależności:**

- `log/slog` (logowanie)

---

### 3. HTTPResponseHandler (`http_response_handler.go`)

**Odpowiedzialność:** Obsługa różnych kodów statusu HTTP i podejmowanie decyzji o dalszych akcjach.

**Kluczowe Typy:**

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

**Kluczowe Metody:**

- `NewHTTPResponseHandler(logger, successInterval) *HTTPResponseHandler`
- `HandleResponse(resp *http.Response, feed, feedData, retryCount) FetchDecision`
- Prywatne handlery:
  - `handleSuccess()` - 200 OK
  - `handleNotModified()` - 304
  - `handlePermanentRedirect()` - 301/308
  - `handleTemporaryRedirect()` - 302/307
  - `handleUnauthorized()` - 401
  - `handleTooManyRequests()` - 429
  - `handleNotFound()` - 404/410
  - `handleClientError()` - 4xx
  - `handleServerError()` - 5xx

**Zależności:**

- `log/slog`
- Czyste funkcje z `calculations.go` (calculateNextFetch, calculateBackoff, parseRetryAfter)

---

### 4. FeedFetcher (`feed_fetcher.go`)

**Odpowiedzialność:** Wykonywanie zapytań HTTP i parsowanie feedów.

**Kluczowe Typy:**

```go
type FeedData struct {
    Articles []Article
    Title    string
}
```

**Kluczowe Metody:**

- `NewFeedFetcher(httpClient, responseHandler, logger) *FeedFetcher`
- `Fetch(ctx, feed, retryCount) (FetchDecision, error)` - główna metoda
- `createRequest(ctx, feed) (*http.Request, error)` - tworzy zapytanie z nagłówkami warunkowymi
- `parseFeed(body io.Reader) (*FeedData, error)` - parsuje RSS/Atom

**Funkcje na poziomie pakietu:**

- `transformFeedItems(items, now) []Article` - transformuje gofeed.Item na Article

**Zależności:**

- `HTTPClient` (http_client.go)
- `HTTPResponseHandler`
- `github.com/mmcdole/gofeed`
- `log/slog`

---

### 5. FeedStatusManager (`feed_status_manager.go`)

**Odpowiedzialność:** Aktualizacja statusów feedów i zapis artykułów.

**Kluczowe Metody:**

- `NewFeedStatusManager(repo, logger) *FeedStatusManager`
- `ApplyDecision(ctx, feed, decision) error` - główna metoda
- `saveArticles(ctx, feedID, articles) error` - prywatny pomocnik

**Logika:**

1.  Zapisuje artykuły, jeśli istnieją (nie blokuje całej operacji w razie błędu)
2.  Aktualizuje status feeda w bazie danych
3.  Resetuje `retry_count` w przypadku sukcesu

**Zależności:**

- `Repository`
- `log/slog`

---

### 6. Scheduler (`scheduler.go`)

**Odpowiedzialność:** Harmonogramowanie i orkiestracja przetwarzania wsadowego.

**Kluczowe Metody:**

- `NewScheduler(repo, workerPool, rateLimiter, feedFetcher, statusManager, logger, config, appCtx) *Scheduler`
- `Start()` - główna pętla z tickerem
- `ProcessBatch()` - pobiera i przetwarza wsad feedów
- `processSingleFeed(feed)` - przetwarza pojedynczy feed
- `handleFetchError(feed, err)` - obsługa błędów HTTP/sieciowych

**Przepływ:**

1.  Ticker co X minut
2.  `FindFeedsDueForFetch()`
3.  `WorkerPool.Process()` dla wszystkich feedów
4.  Dla każdego feeda:
    - `RateLimiter.WaitIfNeeded()`
    - `FeedFetcher.Fetch()`
    - `FeedStatusManager.ApplyDecision()`

**Zależności:**

- `Repository`
- `WorkerPool`
- `RateLimiter`
- `FeedFetcher`
- `FeedStatusManager`
- `log/slog`
- `config.FetcherConfig`

---

### 7. FeedFetcherService (`service.go`)

**Odpowiedzialność:** Główny orkiestrator, wstrzykiwanie zależności, publiczne API.

**Kluczowe Metody:**

- `NewFeedFetcherService(dbClient, logger, cfg, appCtx) *FeedFetcherService` - konfiguracja wszystkich komponentów
- `Start()` - deleguje do `Scheduler.Start()`
- `FetchFeedNow(feedID string)` - bezpośrednie użycie `FeedFetcher` (asynchronicznie)

**Stan:**

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

**Zależności:** Wszystkie komponenty (tworzy je w konstruktorze)

---

### 8. Models (`models.go`)

**Odpowiedzialność:** Współdzielone typy używane przez wiele komponentów.

**Typy:**

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

**Odpowiedzialność:** Czyste matematyczne/czasowe kalkulacje dla pobierania feedów.

**Funkcje czyste na poziomie pakietu:**

```go
// Obliczenia czasowe
func calculateNextFetch(cacheControl string, successInterval time.Duration, now time.Time) time.Time
func parseCacheControlMaxAge(cacheControl string) (time.Duration, bool)

// Strategia ponawiania
func calculateBackoff(retryCount int, now time.Time) time.Time
func parseRetryAfter(retryAfter string, retryCount int, now time.Time) time.Time
```

**Uwaga:** `transformFeedItems()` zostało **przeniesione do `feed_fetcher.go`**, ponieważ jest to funkcja transformacji/mapowania, a nie kalkulacja.

**Zależności:** Brak (czysta biblioteka standardowa Go)

---

## Kroki Migracji

### Faza 1: Ekstrakcja (bez łamania zmian)

1.  ✅ Utwórz `models.go` - wydziel typ `Article`
2.  ✅ Utwórz `rate_limiter.go` - wydziel logikę ograniczania zapytań
3.  ✅ Utwórz `worker_pool.go` - wydziel zarządzanie współbieżnością
4.  ✅ Utwórz `http_response_handler.go` - wydziel obsługę odpowiedzi HTTP
5.  ✅ Utwórz `feed_fetcher.go` - wydziel pobieranie + parsowanie
6.  ✅ Utwórz `feed_status_manager.go` - wydziel aktualizacje statusu
7.  ✅ Utwórz `scheduler.go` - wydziel logikę harmonogramu

### Faza 2: Integracja

8.  ✅ Zrefaktoryzuj `service.go`, aby używał nowych komponentów
9.  ✅ Zaktualizuj testy jednostkowe, aby działały z nową strukturą
10. ✅ Zweryfikuj, czy cała funkcjonalność działa

### Faza 3: Czyszczenie

11. ✅ Usuń stary kod z `service.go`
12. ✅ Zaktualizuj dokumentację
13. ✅ Uruchom lintery i formatery

## Strategia Testowania

### Testy Jednostkowe

Każdy komponent powinien być niezależnie testowalny:

- **RateLimiter**: Test z mockowanym czasem, weryfikacja opóźnień
- **WorkerPool**: Test limitów współbieżności, odzyskiwania po panice, anulowania kontekstu
- **HTTPResponseHandler**: Test wszystkich kodów statusu, weryfikacja poprawnych decyzji
- **FeedFetcher**: Mock `HTTPClient`, test logiki parsowania
- **FeedStatusManager**: Mock `Repository`, weryfikacja aktualizacji w bazie danych
- **Scheduler**: Mock wszystkich zależności, test orkiestracji

### Testy E2E

- Po refaktoryzacji, istniejące testy E2E (end-to-end) muszą przechodzić bez żadnych modyfikacji.
- Gwarantuje to, że publiczne API i zachowanie całego systemu z perspektywy użytkownika końcowego nie uległy zmianie.

## Korzyści

✅ **Pojedyncza odpowiedzialność** - każdy komponent ma jedną odpowiedzialność  
✅ **Testowalność** - małe, izolowane komponenty z mockowalnymi zależnościami  
✅ **Utrzymywalność** - zmiany w jednym miejscu nie wpływają na inne  
✅ **Czytelność** - pliki po ~70-150 linii zamiast 700  
✅ **Rozszerzalność** - łatwe dodawanie nowych funkcji (np. inne strategie ponawiania)  
✅ **Zachowane czyste funkcje** - `calculations.go` pozostaje bez zmian  
✅ **Jasne zależności** - każdy komponent ma jasno zdefiniowane zależności

## Uwagi

- **Strategia czystych funkcji**:
  - `calculations.go` - tylko czyste kalkulacje (timing, backoff, parsowanie nagłówków HTTP) - NIE TWORZYMY Z NICH OBIEKTÓW
  - `transformFeedItems()` - przeniesione z `calculations.go` do `feed_fetcher.go` jako **czysta funkcja na poziomie pakietu (bez receivera)**
    - Powód: semantycznie to transformacja/mapowanie, a nie kalkulacja
    - Pozostaje czystą funkcją dla łatwego testowania
    - Blisko miejsca użycia (`parseFeed`)
- Zachowujemy styl programowania funkcyjnego tam, gdzie ma to sens
- Zachowujemy istniejące publiczne API dla kompatybilności wstecznej
- Bezpieczne zamykanie (graceful shutdown) musi działać we wszystkich komponentach
- Ograniczanie zapytań na domenę jest krytyczne - nie można tego zgubić w refaktoryzacji
- Odzyskiwanie po panice w workerach jest kluczowe dla stabilności
- Ochrona przed SSRF pozostaje w `HTTPClient` (`http_client.go`) w warstwie transportowej
- Obsługa błędów (SSRF vs błędy sieciowe) będzie w `FeedStatusManager`
