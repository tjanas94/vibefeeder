# Plan Testów dla Projektu VibeFeeder

## 1. Wprowadzenie i Cele Testowania

### 1.1. Wprowadzenie

Niniejszy dokument opisuje kompleksowy plan testów dla aplikacji VibeFeeder. Aplikacja jest czytnikiem kanałów RSS/Atom z funkcją generowania streszczeń artykułów przy użyciu AI. Plan obejmuje różne poziomy i typy testów, mające na celu zapewnienie wysokiej jakości, bezpieczeństwa, wydajności i stabilności aplikacji przed wdrożeniem na środowisko produkcyjne.

### 1.2. Cele Testowania

Główne cele procesu testowania to:

- **Weryfikacja Funkcjonalności:** Zapewnienie, że wszystkie kluczowe funkcje aplikacji działają zgodnie z założeniami, w tym uwierzytelnianie, zarządzanie kanałami (CRUD), pobieranie i parsowanie treści oraz generowanie streszczeń.
- **Zapewnienie Bezpieczeństwa:** Identyfikacja i eliminacja potencjalnych luk bezpieczeństwa, takich jak XSS, CSRF, SQL Injection, a w szczególności SSRF.
- **Ocena Użyteczności i Doświadczenia Użytkownika (UX):** Sprawdzenie, czy interfejs użytkownika jest intuicyjny, responsywny i zgodny z zasadami htmx i Alpine.js.
- **Zapewnienie Stabilności i Wydajności:** Weryfikacja, że aplikacja działa stabilnie pod obciążeniem, a kluczowe procesy (jak `fetcher`) są odporne na błędy i zoptymalizowane.
- **Wykrywanie Regresji:** Zapewnienie, że nowe zmiany nie wprowadzają błędów w istniejących, działających już częściach systemu.
- **Weryfikacja Odporności:** Testowanie zachowania aplikacji w warunkach problemów sieciowych, opóźnień i ograniczonych zasobów.

## 2. Zakres Testów

Testy obejmą następujące moduły i funkcjonalności:

- **Moduł `auth` (Uwierzytelnianie i Autoryzacja):**
  - Rejestracja użytkownika (z opcjonalnym kodem).
  - Logowanie i wylogowywanie.
  - Zarządzanie sesją (w tym odświeżanie tokenów).
  - Resetowanie hasła.
  - Ochrona endpointów wymagających autoryzacji.
- **Moduł `feed` (Zarządzanie Kanałami):**
  - Dodawanie, edycja, usuwanie i listowanie kanałów (CRUD).
  - Walidacja formularzy.
  - Paginacja i filtrowanie listy kanałów.
- **Moduł `fetcher` (Pobieranie Treści):**
  - Okresowe pobieranie i parsowanie kanałów RSS/Atom.
  - Obsługa różnych statusów HTTP (200, 304, 301, 4xx, 5xx).
  - Logika ponowień (exponential backoff) i obsługa błędów sieciowych.
  - Ochrona przed SSRF.
  - Zapisywanie nowych artykułów do bazy danych.
  - Parsowanie różnych formatów RSS/Atom (testowanie fuzzing).
- **Moduł `summary` (Generowanie Streszczeń):**
  - Generowanie streszczeń na żądanie.
  - Obsługa błędów (brak artykułów, błędy API AI).
  - Integracja z usługą OpenRouter.
  - Rate limiting.
- **Moduł `dashboard` (Panel Główny):**
  - Dynamiczne ładowanie listy kanałów.
  - Integracja z filtrowaniem i paginacją.
- **Komponenty `shared` (Współdzielone):**
  - Warstwa dostępu do bazy danych (`database`).
  - Logika renderowania widoków (`view`, `templ`).
  - Middleware (logowanie, CSRF, autoryzacja).
- **Interfejs Użytkownika (UI):**
  - Responsywność i poprawne wyświetlanie na różnych urządzeniach.
  - Interaktywność oparta na htmx i Alpine.js.
  - Spójność wizualna (Tailwind, DaisyUI).

## 3. Typy Testów

### 3.1. Testy Jednostkowe (Unit Tests)

- **Cel:** Weryfikacja małych, izolowanych fragmentów kodu (funkcji, metod).
- **Zakres:**
  - Logika biznesowa w serwisach (np. `feed.Service`, `auth.Service`).
  - Funkcje pomocnicze (np. `fetcher.calculateNextFetch`, `auth.validateRegistrationCode`).
  - Logika transformacji danych (np. `models.NewFeedItemFromDB`).
  - Walidatory niestandardowe (`shared/validator`).
  - Komponenty Templ (renderowanie z różnymi danymi wejściowymi).
- **Narzędzia:** Standardowy pakiet `testing` w Go, `stretchr/testify` (dla asercji i mocków).

### 3.2. Testy Integracyjne (Integration Tests)

- **Cel:** Weryfikacja współpracy między różnymi komponentami systemu.
- **Zakres:**
  - **API Endpoints (Handlery):** Symulowanie żądań HTTP do handlerów Echo i weryfikacja odpowiedzi (kody statusu, nagłówki, treść HTML/htmx).
  - **Warstwa Repozytorium:** Testowanie operacji na bazie danych (CRUD) z użyciem dedykowanej, testowej bazy danych.
  - **Integracja Serwis-Repozytorium:** Sprawdzenie, czy serwisy poprawnie wykorzystują repozytoria do operacji na danych.
  - **Integracja z Zewnętrznymi API (Mock):** Testowanie logiki klienta AI (`summary.Service`) z użyciem mockowanego serwera HTTP symulującego OpenRouter API.
  - **Testowanie Komponentów Templ:** Weryfikacja poprawności renderowanego HTML, w tym eskejpowania danych użytkownika.
- **Narzędzia:** `net/http/httptest`, `stretchr/testify`, `testcontainers-go` (do zarządzania testową bazą danych PostgreSQL).

### 3.3. Testy End-to-End (E2E)

- **Cel:** Symulacja rzeczywistych scenariuszy użytkowania aplikacji w przeglądarce.
- **Zakres:**
  - Pełny cykl życia użytkownika: rejestracja, logowanie, dodanie kanału, wygenerowanie streszczenia, wylogowanie.
  - Interakcje UI napędzane przez htmx i Alpine.js (otwieranie modali, przełączanie filtrów, paginacja).
  - Walidacja formularzy po stronie klienta i serwera.
  - Oczekiwanie na żądania AJAX/htmx i dynamiczne aktualizacje DOM.
- **Narzędzia:** Playwright (doskonałe wsparcie dla htmx, natywne API do czekania na żądania sieciowe, szybkie wykonanie).

### 3.4. Testy Wydajnościowe (Performance Tests)

- **Cel:** Weryfikacja wydajności i stabilności aplikacji pod obciążeniem.
- **Zakres:**
  - **Testy Obciążeniowe (Load Testing):**
    - 100 równoczesnych użytkowników przeglądających dashboard.
    - 1000 kanałów do pobrania przez fetcher w ciągu 5 minut.
    - 50 równoległych żądań generowania streszczeń.
  - **Testy Stresowe (Stress Testing):** Określenie limitów systemu poprzez stopniowe zwiększanie obciążenia.
  - **Testy Wydajności Bazy Danych:** Pomiar czasu odpowiedzi dla operacji CRUD przy dużej liczbie rekordów.
  - **Testy Skalowania:** Weryfikacja zachowania aplikacji przy rosnącej liczbie użytkowników i kanałów.
- **Narzędzia:** `k6` (nowoczesne, napisane w Go, łatwe w użyciu, dobre wsparcie dla CI/CD).
- **Metryki:**
  - Średni czas odpowiedzi (p50, p95, p99).
  - Liczba żądań na sekundę (RPS).
  - Wskaźnik błędów (error rate).
  - Zużycie pamięci i CPU.

### 3.5. Testy Bezpieczeństwa

- **Cel:** Identyfikacja i zapobieganie lukom bezpieczeństwa.
- **Zakres:**
  - **SSRF:** Testy jednostkowe dla walidatora `ssrf.ValidateIP` oraz testy integracyjne dla `fetcher.HTTPClient` próbujące uzyskać dostęp do adresów prywatnych, localhost i metadanych chmury.
  - **CSRF:** Testy automatyczne próbujące wysłać żądania POST/PATCH/DELETE bez poprawnego tokenu CSRF.
  - **XSS:** Weryfikacja (manualna i automatyczna), czy dane wejściowe od użytkownika (np. nazwy kanałów, treść artykułów) są poprawnie eskejpowane w widokach Templ.
  - **SQL Injection:** Testy próbujące wstrzyknąć złośliwy SQL przez parametry zapytań (ORM Supabase powinien chronić, ale weryfikacja jest konieczna).
  - **Kontrola Dostępu:** Testy próbujące uzyskać dostęp do chronionych zasobów bez autoryzacji lub jako inny użytkownik.
  - **Bezpieczeństwo Sesji:** Weryfikacja poprawnego ustawiania flag ciasteczek (HttpOnly, Secure, SameSite).
  - **Rate Limiting:** Testy próbujące przekroczyć limity żądań dla wrażliwych endpointów.
- **Narzędzia:**
  - Testy jednostkowe/integracyjne.
  - `gosec` (statyczna analiza kodu Go pod kątem bezpieczeństwa).
  - `OWASP ZAP` (dynamiczne skanowanie aplikacji webowej).

### 3.6. Testy Regresji Wizualnej

- **Cel:** Wykrywanie niezamierzonych zmian w wyglądzie interfejsu użytkownika.
- **Zakres:** Kluczowe widoki aplikacji (dashboard, lista kanałów, modale, formularze, komunikaty błędów).
- **Narzędzia:** `reg-suit` (open source, self-hosted, darmowe, integracja z GitHub i S3).
- **Implementacja:** Zintegrowane z testami E2E, generowanie screenshotów kluczowych stanów aplikacji i porównanie z baseline.

### 3.7. Testy Fuzzing

- **Cel:** Wykrywanie błędów i luk bezpieczeństwa poprzez testowanie z losowymi danymi wejściowymi.
- **Zakres:**
  - **Parser RSS/Atom:** Testowanie z zdeformowanymi, niepełnymi lub złośliwymi plikami XML.
  - **Walidatory:** Testowanie funkcji walidacji z nieoczekiwanymi danymi wejściowymi.
  - **Handlery HTTP:** Testowanie z nieprawidłowymi nagłówkami i parametrami.
- **Narzędzia:** Natywny fuzzing w Go (od wersji 1.18):
  ```go
  func FuzzParseFeed(f *testing.F) {
      f.Fuzz(func(t *testing.T, data []byte) {
          ParseRSSFeed(data) // Nie powinno panikować
      })
  }
  ```
- **Uruchomienie:** `go test -fuzz=FuzzParseFeed -fuzztime=30s`

### 3.8. Testy Odporności (Resilience/Chaos Testing)

- **Cel:** Weryfikacja zachowania aplikacji w warunkach problemów infrastrukturalnych.
- **Zakres:**
  - Symulacja opóźnień sieciowych (latency).
  - Symulacja przerw w połączeniu z bazą danych.
  - Symulacja timeoutów w połączeniach z zewnętrznymi API.
  - Testowanie mechanizmów retry i exponential backoff.
- **Narzędzia:** `toxiproxy` (proxy do symulacji problemów sieciowych):

  ```bash
  # Opóźnienie połączenia do bazy o 1s
  toxiproxy-cli toxic add -t latency -a latency=1000 postgres

  # Symulacja utraty pakietów
  toxiproxy-cli toxic add -t timeout -a timeout=5000 openrouter-api
  ```

- **Zastosowanie:** Opcjonalne, ale zalecane dla krytycznych procesów (fetcher, integracja z AI).

### 3.9. Contract Testing (Opcjonalne)

- **Cel:** Weryfikacja zgodności z API zewnętrznych usług (OpenRouter).
- **Zakres:**
  - Definiowanie kontraktów dla żądań i odpowiedzi API.
  - Weryfikacja, czy aplikacja wysyła poprawne żądania i obsługuje odpowiedzi zgodnie z kontraktem.
- **Narzędzia:** `Pact` (Consumer-Driven Contracts):
  ```go
  pact.AddInteraction().
      Given("AI service is available").
      UponReceiving("A request for summary").
      WithRequest(dsl.Request{
          Method: "POST",
          Path:   "/api/v1/chat/completions",
      }).
      WillRespondWith(dsl.Response{
          Status: 200,
          Body:   dsl.Match(/* expected structure */),
      })
  ```

## 4. Scenariusze Testowe dla Kluczowych Funkcjonalności

### 4.1. Uwierzytelnianie

| ID    | Scenariusz                                                      | Oczekiwany Rezultat                                                                                     | Typ Testu         | Priorytet |
| :---- | :-------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------ | :---------------- | :-------- |
| AU-01 | Pomyślna rejestracja nowego użytkownika.                        | Użytkownik zostaje utworzony, otrzymuje e-mail z linkiem aktywacyjnym, widzi stronę "potwierdź e-mail". | E2E, Integracyjny | Krytyczny |
| AU-02 | Próba rejestracji z zajętym adresem e-mail.                     | Wyświetlany jest błąd "Użytkownik o tym adresie e-mail już istnieje".                                   | E2E, Integracyjny | Wysoki    |
| AU-03 | Pomyślne logowanie z poprawnymi danymi.                         | Użytkownik zostaje zalogowany i przekierowany na `/dashboard`. Ustawiane są ciasteczka sesji.           | E2E, Integracyjny | Krytyczny |
| AU-04 | Próba logowania z niepoprawnym hasłem.                          | Wyświetlany jest błąd "Niepoprawny e-mail lub hasło".                                                   | E2E, Integracyjny | Wysoki    |
| AU-05 | Dostęp do chronionego endpointu (`/dashboard`) bez zalogowania. | Użytkownik zostaje przekierowany na stronę logowania (`/auth/login`).                                   | E2E, Integracyjny | Krytyczny |
| AU-06 | Pomyślne wylogowanie.                                           | Ciasteczka sesji są usuwane, użytkownik zostaje przekierowany na stronę logowania.                      | E2E, Integracyjny | Wysoki    |
| AU-07 | Resetowanie hasła (wysłanie linku i zmiana hasła).              | Użytkownik otrzymuje e-mail, może ustawić nowe hasło i zalogować się przy jego użyciu.                  | E2E, Integracyjny | Wysoki    |
| AU-08 | Weryfikacja flag bezpieczeństwa ciasteczek sesji.               | Ciasteczka mają ustawione flagi: HttpOnly, Secure, SameSite=Lax/Strict.                                 | Integracyjny      | Wysoki    |

### 4.2. Zarządzanie Kanałami (Feed)

| ID    | Scenariusz                                                      | Oczekiwany Rezultat                                                                                             | Typ Testu         | Priorytet |
| :---- | :-------------------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------- | :---------------- | :-------- |
| FD-01 | Pomyślne dodanie nowego kanału RSS/Atom.                        | Kanał pojawia się na liście, modal zostaje zamknięty, wyświetla się toast "Kanał został dodany".                | E2E, Integracyjny | Krytyczny |
| FD-02 | Próba dodania kanału z niepoprawnym URL.                        | Formularz wyświetla błąd walidacji "Musi być poprawnym adresem URL HTTP lub HTTPS".                             | E2E, Integracyjny | Wysoki    |
| FD-03 | Próba dodania kanału, który już istnieje na koncie użytkownika. | Formularz wyświetla błąd "Już dodałeś ten kanał".                                                               | E2E, Integracyjny | Wysoki    |
| FD-04 | Edycja nazwy istniejącego kanału.                               | Nazwa kanału na liście zostaje zaktualizowana.                                                                  | E2E, Integracyjny | Średni    |
| FD-05 | Usunięcie istniejącego kanału.                                  | Kanał znika z listy, wyświetla się toast "Kanał został usunięty".                                               | E2E, Integracyjny | Wysoki    |
| FD-06 | Filtrowanie listy kanałów po nazwie.                            | Lista jest dynamicznie aktualizowana (htmx), aby pokazać tylko pasujące kanały. URL w przeglądarce się zmienia. | E2E               | Średni    |
| FD-07 | Paginacja listy kanałów.                                        | Kliknięcie na numer strony ładuje odpowiednią listę kanałów.                                                    | E2E               | Średni    |
| FD-08 | Próba dodania kanału z nazwą zawierającą znaki specjalne (XSS). | Nazwa jest poprawnie eskejpowana w widoku, brak wykonania skryptu.                                              | Integracyjny, E2E | Krytyczny |

### 4.3. Pobieranie Treści (Fetcher)

| ID    | Scenariusz                                                               | Oczekiwany Rezultat                                                                                                | Typ Testu    | Priorytet |
| :---- | :----------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------- | :----------- | :-------- |
| FT-01 | Pomyślne pobranie i sparsowanie poprawnego kanału RSS.                   | Nowe artykuły są zapisywane w bazie danych. Status kanału to `success`.                                            | Integracyjny | Krytyczny |
| FT-02 | Próba pobrania kanału, który zwraca błąd 500.                            | Status kanału to `temporary_error`, `retry_count` jest inkrementowany, `fetch_after` jest ustawiony w przyszłości. | Integracyjny | Wysoki    |
| FT-03 | Próba pobrania kanału, który zwraca błąd 404.                            | Status kanału to `permanent_error`. Serwis nie próbuje ponownie pobierać tego kanału.                              | Integracyjny | Wysoki    |
| FT-04 | Próba pobrania kanału z adresem URL wskazującym na zasób lokalny (SSRF). | Żądanie jest blokowane na poziomie klienta HTTP. Status kanału to `permanent_error` z komunikatem o błędzie.       | Integracyjny | Krytyczny |
| FT-05 | Pobranie kanału z nagłówkiem `ETag` i `Last-Modified`.                   | Kolejne żądanie wysyła nagłówki `If-None-Match` i `If-Modified-Since`.                                             | Integracyjny | Średni    |
| FT-06 | Otrzymanie odpowiedzi 304 Not Modified.                                  | Żadne nowe artykuły nie są przetwarzane. Status kanału to `success`, `last_fetched_at` jest zaktualizowane.        | Integracyjny | Średni    |
| FT-07 | Otrzymanie odpowiedzi 301 Moved Permanently.                             | URL kanału w bazie danych jest aktualizowany na nowy adres.                                                        | Integracyjny | Wysoki    |
| FT-08 | Parsowanie zdeformowanego pliku RSS/Atom (fuzzing).                      | Parser obsługuje błędy gracefully, nie powoduje panic, loguje błąd.                                                | Fuzzing      | Wysoki    |
| FT-09 | Symulacja timeoutu podczas pobierania kanału (toxiproxy).                | Mechanizm retry jest aktywowany, `retry_count` jest inkrementowany.                                                | Odporności   | Średni    |

### 4.4. Generowanie Streszczeń (Summary)

| ID    | Scenariusz                                                     | Oczekiwany Rezultat                                                                          | Typ Testu         | Priorytet |
| :---- | :------------------------------------------------------------- | :------------------------------------------------------------------------------------------- | :---------------- | :-------- |
| SM-01 | Pomyślne wygenerowanie streszczenia artykułu.                  | Streszczenie jest wyświetlane użytkownikowi, zapisywane w bazie danych.                      | E2E, Integracyjny | Krytyczny |
| SM-02 | Próba wygenerowania streszczenia dla nieistniejącego artykułu. | Wyświetlany jest błąd "Artykuł nie został znaleziony".                                       | Integracyjny      | Wysoki    |
| SM-03 | Błąd API OpenRouter (500, rate limit).                         | Wyświetlany jest użytkownikowi przyjazny komunikat błędu, błąd jest logowany.                | Integracyjny      | Wysoki    |
| SM-04 | Przekroczenie rate limitu dla generowania streszczeń.          | Użytkownik otrzymuje komunikat "Przekroczono limit żądań, spróbuj ponownie za X sekund".     | Integracyjny      | Średni    |
| SM-05 | Symulacja timeoutu API OpenRouter (toxiproxy).                 | Żądanie kończy się błędem timeout, użytkownik otrzymuje komunikat o problemie z połączeniem. | Odporności        | Średni    |

### 4.5. Wydajność

| ID    | Scenariusz                                                | Oczekiwany Rezultat                                                                     | Typ Testu     | Priorytet |
| :---- | :-------------------------------------------------------- | :-------------------------------------------------------------------------------------- | :------------ | :-------- |
| PF-01 | 100 równoczesnych użytkowników przeglądających dashboard. | p95 czasu odpowiedzi < 500ms, error rate < 1%, CPU < 80%, pamięć stabilna.              | Wydajnościowy | Wysoki    |
| PF-02 | Fetcher pobierający 1000 kanałów w ciągu 5 minut.         | Wszystkie kanały są przetworzone, brak memory leaks, błędy są obsługiwane gracefully.   | Wydajnościowy | Wysoki    |
| PF-03 | 50 równoległych żądań generowania streszczeń.             | p95 czasu odpowiedzi < 3s, rate limiting działa poprawnie, brak błędów backend.         | Wydajnościowy | Średni    |
| PF-04 | Paginacja z 10,000+ kanałów.                              | Czas odpowiedzi dla każdej strony < 200ms, brak degradacji przy dużej liczbie rekordów. | Wydajnościowy | Średni    |

## 5. Środowisko Testowe

- **Baza Danych:** Dedykowana instancja PostgreSQL (w kontenerze Docker) zarządzana przez `testcontainers-go`, czyszczona przed każdym uruchomieniem zestawu testów integracyjnych.
- **Zależności Zewnętrzne:**
  - Mock serwer HTTP (`net/http/httptest`) do symulowania odpowiedzi z OpenRouter API.
  - Opcjonalnie: `toxiproxy` do symulacji problemów sieciowych.
- **Środowisko Uruchomieniowe:** Testy będą uruchamiane w środowisku CI/CD (GitHub Actions) na systemie Linux.
- **Przeglądarki (dla E2E):** Testy E2E będą przeprowadzane na najnowszych wersjach Chrome.
- **Narzędzia Wydajnościowe:** `k6` lub `vegeta` uruchamiane w izolowanym środowisku z kontrolowanymi zasobami.

## 6. Narzędzia do Testowania

### 6.1. Testy Jednostkowe i Integracyjne

- `testing` (standardowa biblioteka Go)
- `stretchr/testify` (asercje i mocki)
- `net/http/httptest` (testowanie handlerów HTTP)
- `testcontainers-go` (zarządzanie testową bazą danych PostgreSQL)

### 6.2. Testy E2E

- `Playwright` (TypeScript/JavaScript, doskonałe wsparcie dla htmx i dynamicznych treści)

### 6.3. Testy Wydajnościowe

- `k6` (nowoczesne, skrypty w JavaScript, łatwa integracja z CI/CD)

### 6.4. Testy Regresji Wizualnej

- `reg-suit` (open source, self-hosted, integracja z GitHub)

### 6.5. Testy Bezpieczeństwa

- `gosec` (statyczna analiza bezpieczeństwa kodu Go)
- `OWASP ZAP` (dynamiczne skanowanie aplikacji webowej)
- `staticcheck` (zaawansowana analiza statyczna Go)
- Testy jednostkowe/integracyjne dla konkretnych luk (SSRF, CSRF, XSS)

### 6.6. Fuzzing

- Natywny fuzzing w Go (od wersji 1.18): `go test -fuzz`

### 6.7. Testy Odporności

- `toxiproxy` (symulacja problemów sieciowych, opóźnień, timeoutów)

### 6.8. Analiza Pokrycia Kodu

```bash
# Generowanie profilu pokrycia
go test -coverprofile=coverage.out ./...

# Wizualizacja w HTML
go tool cover -html=coverage.out -o coverage.html

# Wyświetlenie w terminalu
go tool cover -func=coverage.out
```

**Integracja z Taskfile:**

```yaml
coverage:
  desc: "Generate and display coverage report"
  cmds:
    - go test -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
    - go tool cover -func=coverage.out
    - echo "Coverage report saved to coverage.html"
```

### 6.9. Linting i Analiza Statyczna

```yaml
# Taskfile.yml
lint:
  desc: "Run all linters and static analysis"
  cmds:
    - golangci-lint run
    - gosec -exclude-dir=test ./...
    - staticcheck ./...
```

### 6.10. Automatyzacja Zadań

- `go-task` (task runner dla wszystkich operacji testowych)

## 7. Harmonogram Testów

- **Testy Jednostkowe i Integracyjne:**
  - Pisane na bieżąco wraz z nowymi funkcjonalnościami i poprawkami błędów.
  - Uruchamiane automatycznie przy każdym commicie w gałęziach deweloperskich (pre-commit hook).
  - Uruchamiane w CI przy każdym push i pull request.
- **Testy E2E:**
  - Uruchamiane automatycznie przed każdym mergem do głównej gałęzi (`master`).
  - Uruchamiane cyklicznie (np. co noc) na środowisku stagingowym.
- **Testy Wydajnościowe:**
  - Uruchamiane przed dużymi wdrożeniami (release).
  - Uruchamiane cyklicznie (np. co tydzień) na środowisku stagingowym.
  - Uruchamiane po optymalizacjach wydajności (dla weryfikacji poprawy).
- **Testy Bezpieczeństwa:**
  - Podstawowe testy automatyczne (`gosec`, testy SSRF/CSRF) uruchamiane w CI przy każdym push.
  - Pełne, manualne audyty bezpieczeństwa (`OWASP ZAP`) przeprowadzane kwartalnie lub przed dużymi wdrożeniami.
- **Fuzzing:**
  - Uruchamiane lokalnie podczas rozwoju parsera RSS/Atom.
  - Uruchamiane w CI z ograniczonym czasem (np. 30s) dla każdego PR dotyczącego parsera.
  - Długie sesje fuzzingu (np. 1h) uruchamiane co tydzień na dedykowanym środowisku.
- **Testy Regresji Wizualnej:**
  - Uruchamiane automatycznie przy każdym PR modyfikującym komponenty UI.
  - Wymaga manualnej akceptacji zmian wizualnych przez zespół przed mergem.

## 8. Kryteria Akceptacji Testów

- **Pokrycie Kodu:**
  - Minimalne pokrycie kodu testami jednostkowymi i integracyjnymi na poziomie **80%**.
  - Pokrycie kodu dla krytycznych modułów (`auth`, `fetcher`, `summary`) na poziomie **90%**.
- **Status Testów:**
  - **100%** testów jednostkowych, integracyjnych i E2E musi zakończyć się sukcesem, aby zmiana mogła zostać włączona do głównej gałęzi.
  - Brak błędów z `golangci-lint`, `gosec` i `staticcheck`.
- **Błędy Krytyczne:**
  - Brak jakichkolwiek niezaadresowanych błędów o priorytecie **Krytyczny** lub **Wysoki**.
- **Regresja Wizualna:**
  - Wszystkie zmiany wizualne muszą zostać zaakceptowane przez zespół deweloperski/produktowy przed mergem.
- **Wydajność:**
  - Spełnienie metryk wydajnościowych określonych w scenariuszach testowych (sekcja 4.5).
  - Brak degradacji wydajności > 10% w porównaniu do poprzedniej wersji.
- **Bezpieczeństwo:**
  - Brak nowych luk bezpieczeństwa wykrytych przez `gosec` lub `OWASP ZAP`.
  - Wszystkie znane luki bezpieczeństwa muszą być zaadresowane przed wdrożeniem.

## 9. Role i Odpowiedzialności

- **Deweloperzy:**
  - Odpowiedzialni za pisanie testów jednostkowych i integracyjnych dla tworzonego przez siebie kodu.
  - Uruchamianie testów lokalnie przed commitowaniem (`task test`).
  - Odpowiedzialni za naprawę błędów wykrytych na wszystkich etapach testowania.
  - Pisanie testów fuzzing dla parsera RSS/Atom i innych funkcji przetwarzających dane wejściowe użytkownika.
- **Inżynier QA / Inżynier Automatyzacji Testów:**
  - Odpowiedzialny za tworzenie i utrzymanie testów E2E.
  - Konfiguracja i utrzymanie narzędzi do testowania (Playwright, k6, reg-suit, toxiproxy).
  - Konfiguracja i monitorowanie testów wydajnościowych.
  - Konfiguracja i uruchamianie skanów bezpieczeństwa (`OWASP ZAP`).
  - Monitorowanie wyników testów w CI/CD i raportowanie błędów.
  - Utrzymanie środowiska testowego i testowej bazy danych.
- **Code Reviewerzy:**
  - Weryfikacja, czy nowy kod zawiera odpowiednie testy.
  - Weryfikacja, czy testy są sensowne i pokrywają przypadki brzegowe.
  - Sprawdzenie, czy kod spełnia standardy jakości (linting, pokrycie, wydajność).
  - Akceptacja zmian wizualnych w testach regresji wizualnej.
- **DevOps / SRE:**
  - Konfiguracja i utrzymanie pipeline'ów CI/CD.
  - Zarządzanie środowiskami testowymi (staging, performance).
  - Monitorowanie wyników testów wydajnościowych i alertowanie przy degradacji.

## 10. Procedury Raportowania Błędów

- Wszystkie błędy wykryte podczas testów (manualnych i automatycznych) będą raportowane jako **Issues** w systemie śledzenia projektu (GitHub Issues).
- Każdy raport o błędzie powinien zawierać:
  - **Tytuł** opisujący problem (format: `[Typ] Krótki opis`, np. `[Bug] Błąd walidacji przy dodawaniu kanału`).
  - **Szczegółowy opis** kroków do reprodukcji błędu (krok po kroku).
  - **Oczekiwany vs. rzeczywisty rezultat** (co powinno się stać vs. co się stało).
  - **Zrzuty ekranu, logi lub fragmenty kodu** (jeśli dotyczy).
  - **Informacje o środowisku:**
    - Przeglądarka i wersja (dla błędów UI).
    - Wersja Go i systemu operacyjnego (dla błędów backend).
    - Wersja bazy danych (jeśli dotyczy).
  - **Priorytet błędu:**
    - **Krytyczny:** Blokuje podstawową funkcjonalność, uniemożliwia użytkowanie aplikacji.
    - **Wysoki:** Poważny błąd wpływający na ważną funkcjonalność lub bezpieczeństwo.
    - **Średni:** Błąd wpływający na wygodę użytkownika, ale z workaroundem.
    - **Niski:** Drobny błąd wizualny lub niedogodność.
  - **Labele:** `bug`, `security`, `performance`, `ui`, `backend`, etc.
  - **Milestone:** Wersja, w której błąd powinien zostać naprawiony.

- **Proces zarządzania błędami:**
  1. Raportowanie błędu jako Issue w GitHub.
  2. Triage przez zespół (priorytetyzacja, przypisanie).
  3. Implementacja poprawki wraz z testem regresji.
  4. Code review i weryfikacja poprawki.
  5. Merge do głównej gałęzi po przejściu wszystkich testów.
  6. Zamknięcie Issue z odwołaniem do commita/PR.

## 11. Przykładowe Komendy i Konfiguracja

### 11.1. Taskfile.yml - Przykładowe Taski

```yaml
version: "3"

tasks:
  test:
    desc: "Run all unit and integration tests"
    cmds:
      - go test -v -race ./...

  test:unit:
    desc: "Run only unit tests"
    cmds:
      - go test -v -short ./...

  test:integration:
    desc: "Run only integration tests"
    cmds:
      - go test -v -run Integration ./...

  test:e2e:
    desc: "Run E2E tests with Playwright"
    cmds:
      - npx playwright test
    # Lub z go-rod:
    # - go test -v ./e2e/...

  test:fuzz:
    desc: "Run fuzzing tests for RSS parser"
    cmds:
      - go test -fuzz=FuzzParseFeed -fuzztime=30s ./internal/fetcher/

  test:performance:
    desc: "Run performance tests with k6"
    cmds:
      - k6 run tests/performance/load-test.js

  test:security:
    desc: "Run security scans"
    cmds:
      - gosec -exclude-dir=test ./...
      - task: test:security:zap

  test:security:zap:
    desc: "Run OWASP ZAP security scan"
    cmds:
      - docker run -t owasp/zap2docker-stable zap-baseline.py -t http://localhost:8080

  coverage:
    desc: "Generate and display coverage report"
    cmds:
      - go test -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html
      - go tool cover -func=coverage.out
      - echo "Coverage report saved to coverage.html"

  lint:
    desc: "Run all linters and static analysis"
    cmds:
      - golangci-lint run
      - gosec -exclude-dir=test ./...
      - staticcheck ./...

  ci:
    desc: "Run all CI checks (lint, test, coverage)"
    cmds:
      - task: lint
      - task: test
      - task: coverage
```

### 11.2. Przykład Testu Wydajnościowego (k6)

```javascript
// tests/performance/load-test.js
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 20 }, // Ramp-up do 20 użytkowników
    { duration: "1m", target: 100 }, // Ramp-up do 100 użytkowników
    { duration: "3m", target: 100 }, // Utrzymanie 100 użytkowników
    { duration: "30s", target: 0 }, // Ramp-down do 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% żądań < 500ms
    http_req_failed: ["rate<0.01"], // Error rate < 1%
  },
};

export default function () {
  const res = http.get("http://localhost:8080/dashboard");
  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

### 11.3. Przykład Testu Fuzzing

```go
// internal/fetcher/parser_test.go
func FuzzParseFeed(f *testing.F) {
    // Seed corpus - przykładowe poprawne dane
    f.Add([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Test</title></channel></rss>`))

    f.Fuzz(func(t *testing.T, data []byte) {
        // Test nie powinien panikować dla żadnych danych wejściowych
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Parser panicked: %v", r)
            }
        }()

        _, err := ParseRSSFeed(data)
        // Błędy są OK, ale nie panic
        _ = err
    })
}
```

### 11.4. Przykład Testu Komponenty Templ

```go
// internal/dashboard/view_test.go
func TestDashboardView_Render(t *testing.T) {
    data := DashboardData{
        Feeds: []Feed{
            {ID: "1", Name: "Test Feed", URL: "https://example.com/feed"},
        },
    }

    component := View(data)
    buf := new(bytes.Buffer)
    err := component.Render(context.Background(), buf)

    assert.NoError(t, err)
    html := buf.String()
    assert.Contains(t, html, "Test Feed")
    assert.Contains(t, html, "https://example.com/feed")
}

func TestDashboardView_XSSProtection(t *testing.T) {
    data := DashboardData{
        Feeds: []Feed{
            {ID: "1", Name: "<script>alert('XSS')</script>", URL: "https://example.com/feed"},
        },
    }

    component := View(data)
    buf := new(bytes.Buffer)
    err := component.Render(context.Background(), buf)

    assert.NoError(t, err)
    html := buf.String()
    // Upewnij się, że skrypt został eskejpowany
    assert.NotContains(t, html, "<script>")
    assert.Contains(t, html, "&lt;script&gt;")
}
```

## 12. Metryki i Monitorowanie Testów

### 12.1. Metryki do Śledzenia

- **Pokrycie Kodu:** Procent kodu pokrytego testami (cel: 80%+).
- **Czas Wykonania Testów:** Czas trwania całego zestawu testów (cel: < 5 min dla CI).
- **Współczynnik Niestabilności (Flaky Rate):** Procent testów, które przechodzą/nie przechodzą niekonsekvwentnie (cel: < 1%).
- **Liczba Błędów:** Liczba otwartych bugów według priorytetu.
- **Czas Naprawy Błędów (MTTR):** Średni czas od wykrycia błędu do jego naprawy.
- **Wydajność:**
  - Średni czas odpowiedzi (p50, p95, p99).
  - Throughput (RPS - requests per second).
  - Error rate.

### 12.2. Dashboardy

- **CI/CD Pipeline:** Monitorowanie statusu testów w GitHub Actions.
- **Coverage Dashboard:** Wizualizacja pokrycia kodu (np. w Codecov lub własny dashboard).
- **Performance Dashboard:** Wykresy metryk wydajności z k6 (np. w Grafana).

## 13. Przyszłe Rozszerzenia Planu Testów

- **Mutation Testing:** Weryfikacja jakości testów poprzez wprowadzanie mutacji w kodzie i sprawdzanie, czy testy je wykrywają (`go-mutesting`).
- **Property-Based Testing:** Testowanie właściwości kodu zamiast konkretnych przypadków (`gopter`).
- **A/B Testing:** Testowanie różnych wersji UI/funkcjonalności na produkcji z użytkownikami.
- **Monitoring w Produkcji:** Integracja z narzędziami monitoring (np. Sentry, DataDog) do wykrywania błędów w czasie rzeczywistym.
- **Testy Dostępności (Accessibility):** Automatyczne testy WCAG (np. `axe-core`).

---

**Ostatnia Aktualizacja:** 2025-10-20
**Wersja Dokumentu:** 2.0
