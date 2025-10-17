# Plan wdrożenia usługi cyklicznego pobierania feedów

## 1. Opis usługi

Usługa `FeedFetcher` będzie działać w tle jako cykliczne zadanie (background job). Jej głównym celem jest automatyczne i "uprzejme" (polite) pobieranie treści z kanałów RSS/Atom dodanych przez użytkowników, parsowanie artykułów i zapisywanie ich w bazie danych. Usługa będzie implementować mechanizmy takie jak warunkowe pobieranie, obsługa kodów statusu HTTP, ponawianie prób po błędach oraz ograniczanie liczby zapytań, aby zminimalizować obciążenie serwerów docelowych.

Nowa usługa zostanie umieszczona w `internal/fetcher/service.go`.

## 2. Opis konstruktora

Konstruktor `NewFeedFetcherService` zainicjuje nową instancję serwisu, przyjmując wszystkie niezbędne zależności, strukturę konfiguracyjną oraz główny kontekst aplikacji.

```go
// internal/fetcher/service.go

type FetcherConfig struct {
    FetchInterval       time.Duration
    SuccessInterval     time.Duration
    WorkerCount         int
    BatchSize           int
    DomainDelay         time.Duration
    JobTimeout          time.Duration
    RequestTimeout      time.Duration
    MaxArticlesPerFeed  int
    MaxResponseBodySize int64
}

type FeedFetcherService struct {
    dbClient      *database.Client
    logger        *slog.Logger
    httpClient    *http.Client
    config        FetcherConfig
    appCtx        context.Context // Główny kontekst aplikacji

    // Mapa do śledzenia ostatniego czasu zapytania dla danej domeny (rate limiting)
    domainLastRequest map[string]time.Time
    // Mutex do synchronizacji dostępu do mapy domainLastRequest
    mu sync.Mutex
}

func NewFeedFetcherService(
    dbClient *database.Client,
    logger *slog.Logger,
    config FetcherConfig,
    appCtx context.Context,
) *FeedFetcherService {
    return &FeedFetcherService{
        dbClient:      dbClient,
        logger:        logger,
        httpClient: &http.Client{
            Timeout: config.RequestTimeout,
        },
        config:        config,
        appCtx:        appCtx,
        domainLastRequest: make(map[string]time.Time),
    }
}
```

### Konfiguracja

Parametry usługi powinny być zgrupowane w strukturze `FetcherConfig` i wstrzykiwane do konstruktora. Wartości te powinny być ładowane z głównej konfiguracji aplikacji (np. z pliku `.env`):

- `FETCHER_INTERVAL`: Czas (w sekundach), co jaki serwis "budzi się", by sprawdzić listę feedów do pobrania. **Domyślnie: 300 sekund (5 minut).**
- `FETCHER_SUCCESS_INTERVAL`: **Minimalny** domyślny czas (w sekundach) do następnego sprawdzenia feeda po udanym pobraniu. Zostanie użyty, jeśli `Cache-Control` z serwera będzie krótszy. **Domyślnie: 3600 sekund (1 godzina).**
- `FETCHER_WORKERS`: Liczba jednoczesnych gorutyn (workerów) do pobierania feedów. **Domyślnie: 10.**
- `FETCHER_BATCH_SIZE`: Maksymalna liczba feedów do przetworzenia w jednej partii. **Domyślnie: 1000.**
- `FETCHER_DOMAIN_DELAY`: Minimalny czas (w sekundach) pomiędzy kolejnymi zapytaniami do tej samej domeny. **Domyślnie: 3 sekundy.**
- `FETCHER_REQUEST_TIMEOUT`: Maksymalny czas (w sekundach) na pojedyncze zapytanie HTTP (łącznie z przekierowaniami). **Domyślnie: 30 sekund.**
- `FETCHER_JOB_TIMEOUT`: Maksymalny czas (w sekundach) na przetworzenie całego zadania dla jednego feeda (pobranie, parsowanie, zapis do bazy). Musi być dłuższy niż `FETCHER_REQUEST_TIMEOUT`. **Domyślnie: 45 sekund.**
- `FETCHER_MAX_ARTICLES`: Maksymalna liczba artykułów do zapisania z pojedynczego feeda. Chroni przed spamem w bazie danych. **Domyślnie: 100.**
- `FETCHER_MAX_BODY_SIZE_MB`: Maksymalny rozmiar odpowiedzi HTTP w megabajtach. Chroni przed atakami typu zip bomb. **Domyślnie: 2 MB.**

**Zależności:**

- `dbClient (*database.Client)`: Klient bazy danych do odczytu feedów i zapisu artykułów.
- `logger (*slog.Logger)`: Logger do zapisu informacji o przebiegu operacji i błędach.

## 3. Publiczne metody i pola

### `Start()`

Uruchamia główną pętlę usługi, która w odstępach czasu zdefiniowanych w `s.config.FetchInterval` wywołuje `processFeeds`. Pętla używa wewnętrznego kontekstu aplikacji (`s.appCtx`) do obsługi graceful shutdown.

### `ProcessFeeds()`

Główna metoda orkiestrująca. Pobiera z bazy danych listę feedów do przetworzenia, a następnie dla każdego z nich wywołuje `fetchSingleFeed` w osobnej gorutynie, zarządzając pulą `s.config.WorkerCount` gorutyn (np. za pomocą `semaphore`). Przekazuje wewnętrzny kontekst aplikacji do workerów.

### `FetchFeedNow(feedID uuid.UUID)`

Umożliwia natychmiastowe, asynchroniczne pobranie konkretnego feeda. Metoda nie przyjmuje kontekstu jako argumentu, lecz używa głównego kontekstu aplikacji zapisanego w strukturze serwisu. Zapewnia to, że zadanie będzie działać do końca, ale zostanie przerwane w przypadku zamknięcia całej aplikacji.

## 4. Prywatne metody i pola

### `fetchSingleFeed(feed models.Feed)`

Odpowiada za pobranie i przetworzenie _jednego_ feeda. Używa wewnętrznego kontekstu aplikacji (`s.appCtx`) do tworzenia zapytań HTTP. Implementuje logikę "polite bota".

### `applyRateLimit(domain string)`

Sprawdza, czy od ostatniego zapytania do danej domeny minął wystarczający czas. Jeśli nie, czeka wymaganą ilość czasu.

### `updateFeedStatus(ctx context.Context, feed models.Feed, status string, errorMsg *string)`

Aktualizuje status feeda w bazie danych (`last_fetch_status`, `last_fetch_error`, `last_fetched_at`, `fetch_after`).

### `saveNewArticles(ctx context.Context, feedID uuid.UUID, articles []*gofeed.Item)`

Zapisuje nowe artykuły w bazie danych, upewniając się, że duplikaty nie są wstawiane (korzystając z `UNIQUE` constraint na `(feed_id, url)`).

## 5. Obsługa błędów

Usługa musi być odporna na błędy i odpowiednio na nie reagować, używając następujących statusów w bazie danych: `success`, `temporary_error`, `permanent_error`, `unauthorized`.

1.  **Błędy sieciowe (Timeout, DNS error, Connection Refused):** Traktowane jako błędy tymczasowe. Feed zostaje oznaczony statusem `temporary_error`. Usługa powinna zaplanować ponowienie próby z użyciem mechanizmu exponential backoff, aktualizując pole `fetch_after`.
2.  **Statusy HTTP:**
    - `200 OK`: Sukces. Feed zostaje oznaczony statusem `success`. (Uwaga: jeśli po pobraniu okaże się, że treść nie jest feedem, status zostanie nadpisany - patrz punkt "Błędy parsowania").
    - `301 Moved Permanently`, `308 Permanent Redirect`: Przekierowania permanentne. Domyślny klient HTTP w Go podąży za przekierowaniem. Usługa powinna wykryć zmianę adresu URL i zaktualizować URL feeda w bazie danych. Operacja kończy się statusem `success`.
    - `302 Found`, `307 Temporary Redirect`: Przekierowania tymczasowe. Klient HTTP podąży za nimi, ale URL w bazie danych nie powinien być modyfikowany. Operacja kończy się statusem `success`.
    - `304 Not Modified`: Sukces. Feed nie został zmieniony i zostaje oznaczony statusem `success`.
    - `401 Unauthorized`, `403 Forbidden`: Błąd autoryzacji. Może oznaczać, że feed jest prywatny lub serwer zablokował naszego bota. Feed zostaje oznaczony statusem `unauthorized`, a automatyczne pobieranie dla niego zostaje wstrzymane.
    - `400 Bad Request`, `404 Not Found`, `410 Gone` oraz pozostałe `4xx` (z wyjątkiem 429, 401, 403): Błędy permanentne, które wymagają akcji użytkownika (np. poprawy URL). Feed zostaje oznaczony statusem `permanent_error`, a automatyczne pobieranie dla niego zostaje wstrzymane.
    - `429 Too Many Requests`: Błąd tymczasowy. Feed zostaje oznaczony statusem `temporary_error`. Należy odczytać nagłówek `Retry-After` i zaplanować następną próbę zgodnie z jego wartością (lub użyć exponential backoff).
    - `5xx Server Error`: Błędy tymczasowe. Feed zostaje oznaczony statusem `temporary_error`. Należy użyć exponential backoff do zaplanowania kolejnej próby.
3.  **Błędy parsowania (nieprawidłowy XML/RSS):** Błąd permanentny. Jest to typowy scenariusz, gdy status HTTP to `200 OK`, ale URL prowadzi do zwykłej strony HTML zamiast pliku z feedem. Feed zostaje oznaczony statusem `permanent_error` z odpowiednim komunikatem, a automatyczne pobieranie zostaje wstrzymane.
4.  **Błędy bazy danych:** Każda operacja na bazie danych powinna być logowana. W przypadku niepowodzenia transakcji (np. przy zapisie artykułów), cała operacja dla danego feeda powinna zostać wycofana i ponowiona po krótkiej chwili.
5.  **Panic:** Główna pętla i każda gorutyna przetwarzająca feed powinny być opakowane w `defer` z `recover()`, aby błąd w jednym feedzie nie zatrzymał całej usługi.

## 6. Kwestie bezpieczeństwa

1.  **Ograniczenie rozmiaru odpowiedzi:** Użycie `http.MaxBytesReader` do ograniczenia maksymalnego rozmiaru pobieranego pliku (np. do 2MB), aby chronić się przed atakiem typu "zip bomb".
2.  **Ograniczenie liczby artykułów:** Po sparsowaniu feeda, jeśli liczba artykułów jest duża (np. > 100), należy przetworzyć tylko część z nich, aby uniknąć zaśmiecenia bazy danych.
3.  **User-Agent:** Ustawienie jasnego `User-Agent` (np. `VibeFeeder/1.0 (+https://vibefeeder.app/bot.html)`) pozwala administratorom stron identyfikować bota i w razie potrzeby kontaktować się.
4.  **Walidacja URL-i:** Upewnić się, że usługa nie próbuje łączyć się z adresami lokalnymi lub wewnętrznymi (SSRF).
5.  **Limit przekierowań:** Klient HTTP musi mieć ustawiony limit przekierowań, aby chronić usługę przed nieskończonymi pętlami. Będziemy korzystać z domyślnego, wbudowanego w Go limitu, który wynosi 10 przekierowań na zapytanie.

## 7. Plan wdrożenia krok po kroku

### Krok 1: Utworzenie struktury i definicji serwisu

1.  Utwórz nowy katalog `internal/fetcher`.
2.  W pliku `internal/fetcher/service.go` zdefiniuj strukturę `FeedFetcherService` oraz konstruktor `NewFeedFetcherService` zgodnie z sekcją 2.
3.  Zdefiniuj szkielety metod publicznych i prywatnych zgodnie z nowymi sygnaturami (bez argumentu `context`).

### Krok 2: Implementacja głównej pętli (`Start` i `ProcessFeeds`)

1.  W `Start`, użyj `time.NewTicker` z interwałem `s.config.FetchInterval` do cyklicznego wywoływania `processFeeds`.
2.  Dodaj obsługę `s.appCtx.Done()` w pętli `select`, aby umożliwić graceful shutdown.
3.  W `ProcessFeeds`, zaimplementuj logikę pobierania feedów z bazy. Użyj zapytania SQL, które wybierze feedy, dla których `fetch_after` jest w przeszłości lub `NULL`, sortując po `last_fetched_at ASC NULLS FIRST`.
4.  Użyj puli `s.config.WorkerCount` gorutyn (np. `chan struct{}` działający jak semafor) do jednoczesnego przetwarzania feedów. W pętli wywołuj `go s.fetchSingleFeed(feed)`.

### Krok 3: Implementacja logiki pobierania (`fetchSingleFeed`)

0.  **Utworzenie kontekstu zadania (Job Context):** Na samym początku metody utwórz nowy kontekst z timeoutem dla całego zadania: `jobCtx, cancel := context.WithTimeout(s.appCtx, s.config.JobTimeout)`. Pamiętaj o wywołaniu `defer cancel()` na końcu. **Wszystkie kolejne operacje w tej metodzie muszą używać `jobCtx`**.
1.  **Rate Limiting:** Na początku metody wywołaj `applyRateLimit`, przekazując domenę URL-a. Metoda ta powinna używać `sync.Mutex` do bezpiecznego dostępu do mapy `domainLastRequest` i sprawdzać, czy od ostatniego zapytania minął czas zdefiniowany w `s.config.DomainDelay`.
2.  **Zapytanie HTTP:**
    - Stwórz nowe zapytanie `http.NewRequestWithContext(jobCtx, ...)`.
    - Ustaw nagłówek `User-Agent`.
    - Jeśli `feed.Etag` istnieje, ustaw nagłówek `If-None-Match`.
    - Jeśli `feed.LastModified` istnieje, ustaw nagłówek `If-Modified-Since`.
3.  **Obsługa odpowiedzi:**
    - Ogranicz rozmiar odpowiedzi za pomocą `http.MaxBytesReader`.
    - Wykonaj zapytanie.
    - Użyj `switch resp.StatusCode` do obsługi różnych kodów odpowiedzi (200, 304, 4xx, 5xx) zgodnie z sekcją 5.
4.  **Parsowanie:**
    - Użyj biblioteki `github.com/mmcdole/gofeed` do sparsowania odpowiedzi.
    - Zabezpiecz proces parsowania za pomocą `recover`.
5.  **Aktualizacja i zapis:**
    - **Po sukcesie (`200 OK`, `304 Not Modified`):**
      - Zaktualizuj `Etag` i `LastModified` na podstawie nagłówków odpowiedzi.
      - Tylko dla `200 OK`, wywołaj `saveNewArticles`.
      - Oblicz następny czas pobrania (`fetch_after`). Użyj **większej** z dwóch wartości: czasu z nagłówka `Cache-Control: max-age` (jeśli dostępny) lub naszego minimalnego interwału `s.config.SuccessInterval`. Zapewni to, że nie będziemy odpytywać serwera częściej, niż sami sobie na to pozwolimy.
      - Wywołaj `updateFeedStatus` ze statusem `success`, przekazując obliczony `fetch_after`.
    - **Po błędzie permanentnym (np. 404, błąd parsowania) lub błędzie autoryzacji (401/403):**
      - Wywołaj `updateFeedStatus` z odpowiednim statusem (`permanent_error` lub `unauthorized`) i komunikatem błędu. Ta metoda **nie powinna** ustawiać nowego `fetch_after`.
    - **Po błędzie tymczasowym (np. 5xx, timeout zadania):**
      - Oblicz następny czas pobrania (`fetch_after`) używając logiki exponential backoff.
      - Wywołaj `updateFeedStatus` ze statusem `temporary_error` (lub podobnym) i komunikatem błędu, przekazując obliczony `fetch_after`.

### Krok 4: Integracja z aplikacją

1.  W `cmd/vibefeeder/main.go`, utwórz instancję `FeedFetcherService`, przekazując do niej główny kontekst aplikacji.
2.  Uruchom usługę w osobnej gorutynie: `go feedFetcherService.Start()`.
3.  Przekaż instancję serwisu do handlerów HTTP, które jej potrzebują (np. do `feed.Handler`).
4.  W `feed.Handler`, w metodach odpowiedzialnych za dodawanie i aktualizację feedów:
    - Po INSERT/UPDATE wywołaj `s.fetcherService.FetchFeedNow(feed.ID)` aby zlecić natychmiastowe pobranie w tle
    - INSERT/UPDATE ustawia `fetch_after = NOW() + INTERVAL '5 minutes'` (fallback dla regular cron job)
    - `FetchFeedNow` uruchamia natychmiastowe pobranie niezależnie od `fetch_after`

### Krok 5: Testowanie i monitoring

1.  Dodaj obszerne logowanie w kluczowych miejscach: start/stop serwisu, początek/koniec przetwarzania batcha, sukces/błąd dla każdego feeda.

## 8. Architektura i kluczowe decyzje projektowe

Ta sekcja opisuje kluczowe decyzje architektoniczne i mechanizmy wewnętrzne serwisu, które nie zostały szczegółowo opisane w krokach implementacyjnych.

### Mechanizm Graceful Shutdown

Graceful shutdown opiera się na mechanizmie `context.Context` w Go.

1.  Główna funkcja aplikacji w `main.go` tworzy nadrzędny kontekst i przekazuje go do konstruktora serwisu.
2.  Aplikacja nasłuchuje na sygnały systemowe (`SIGINT`, `SIGTERM`). Po otrzymaniu sygnału, nadrzędny kontekst jest anulowany.
3.  Główna pętla serwisu w metodzie `Start` nasłuchuje na anulowanie kontekstu (`<-s.appCtx.Done()`) i po jego otrzymaniu kończy swoje działanie.
4.  Każde zadanie (`fetchSingleFeed`) tworzy swój własny kontekst z timeoutem (`jobCtx`), który jest dzieckiem głównego kontekstu aplikacji (`s.appCtx`). Oznacza to, że anulowanie `s.appCtx` natychmiast anuluje również `jobCtx`, co przerywa wszystkie trwające operacje (w tym zapytania HTTP) i zapewnia szybkie, czyste zamknięcie.

### Strategia Rate-Limitingu (Worker blokujący)

Świadomie wybrano prostsze podejście, w którym worker, napotykając na limit zapytań do danej domeny, czeka (`time.Sleep`), blokując swoje miejsce w puli.

- **Uzasadnienie:** Jest to rozwiązanie znacznie prostsze w implementacji i wystarczająco dobre dla początkowej wersji aplikacji.
- **Alternatywa (do rozważenia w przyszłości):** Bardziej złożone, ale wydajniejsze podejście polegałoby na "odkładaniu" zadania na później (np. przez aktualizację `fetch_after` w bazie) i zwalnianiu workera, aby mógł zająć się innym zadaniem. Taka optymalizacja może zostać wprowadzona, jeśli monitoring wykaże, że bezczynność workerów jest problemem.

### Mechanika puli workerów (Semafor)

Pula workerów jest realizowana za pomocą kanału buforowanego, który działa jak semafor.

1.  Pętla zlecająca zadania w `processFeeds` próbuje wysłać wartość do kanału przed uruchomieniem każdej gorutyny.
2.  Jeśli pula jest pełna (bufor kanału jest zapełniony), operacja wysyłania **blokuje pętlę**, aż zwolni się miejsce.
3.  Każdy worker po zakończeniu pracy odbiera wartość z kanału, zwalniając tym samym miejsce w puli.
    Gwarantuje to, że liczba aktywnych workerów nigdy nie przekroczy skonfigurowanego limitu.

### Sekwencyjne przetwarzanie partii (Batchy)

Główny "zegar" (`Ticker`) w metodzie `Start` regularnie inicjuje nowy cykl pracy, ale wywołanie `processFeeds(ctx)` jest operacją blokującą.

- **Efekt:** Nowa partia zadań **nigdy nie rozpocznie się, dopóki poprzednia w pełni się nie zakończy**.
- **Gwarancja:** Zapobiega to nakładaniu się na siebie głównych cykli przetwarzania i rywalizacji o zasoby, zapewniając stabilne i przewidywalne działanie serwisu.
