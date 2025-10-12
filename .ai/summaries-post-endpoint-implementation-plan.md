# API Endpoint Implementation Plan: POST /summaries

## 1. Przegląd punktu końcowego

Celem tego punktu końcowego jest generowanie nowego podsumowania przy użyciu AI na podstawie artykułów użytkownika opublikowanych w ciągu ostatnich 24 godzin. Endpoint jest wyzwalany przez puste żądanie POST i zwraca fragment HTML (renderowany przez Templ) zawierający nowe podsumowanie lub komunikat o błędzie.

## 2. Szczegóły żądania

- **Metoda HTTP**: `POST`
- **Struktura URL**: `/summaries`
- **Parametry**:
  - **Wymagane**: Brak. `user_id` jest pobierany z uwierzytelnionej sesji.
  - **Opcjonalne**: Brak.
- **Request Body**: Puste.

## 3. Wykorzystywane typy

- **Command Model**: `models.GenerateSummaryCommand`
- **View Models (DTOs)**:
  - `models.SummaryDisplayViewModel`
  - `models.SummaryErrorViewModel`
  - `models.SummaryViewModel`
- **Modele bazodanowe**:
  - `database.PublicSummariesInsert`
  - `database.PublicEventsInsert`

## 4. Szczegóły odpowiedzi

- **Odpowiedź sukcesu**:
  - **Kod stanu**: `200 OK`
  - **Content-Type**: `text/html`
  - **Body**: Renderowany widok `internal/summary/view/display.templ` z `SummaryDisplayViewModel` zawierającym dane nowego podsumowania.
- **Odpowiedzi błędów**:
  - **Kod stanu**: `400`, `404`, `500`, `503`
  - **Content-Type**: `text/html`
  - **Body**: Renderowany widok `internal/summary/view/error.templ` z `SummaryErrorViewModel` zawierającym odpowiedni komunikat o błędzie.

## 5. Przepływ danych

1.  Użytkownik wysyła żądanie `POST /summaries`.
2.  Middleware `AuthMiddleware` weryfikuje token JWT, pobiera `user_id` i umieszcza go w kontekście żądania.
3.  `SummaryHandler` odbiera żądanie.
4.  Handler wywołuje `SummaryService.GenerateSummary(ctx)`.
5.  `SummaryService`:
    a. Pobiera wszystkie artykuły dla danego `user_id` z tabeli `articles`, gdzie `published_at` jest w ciągu ostatnich 24 godzin.
    b. Jeśli nie ma artykułów, zwraca błąd `ErrNoArticlesFound`.
    c. Konkatenuje tytuły i treść artykułów, tworząc pojedynczy tekst (prompt) dla modelu AI.
    d. Wywołuje klienta OpenRouter.ai z przygotowanym promptem.
    e. W przypadku błędu z API AI, zwraca błąd `ErrAIServiceUnavailable`.
    f. Po otrzymaniu odpowiedzi, zapisuje nowe podsumowanie w tabeli `summaries` z `user_id` i treścią.
    g. W przypadku błędu zapisu do bazy, zwraca błąd `ErrDatabase`.
    h. Zapisuje zdarzenie `summary_generated` z `user_id` w tabeli `events`.
    i. Zwraca nowo utworzony model podsumowania do handlera.
6.  `SummaryHandler`:
    a. W przypadku sukcesu, renderuje widok `display.templ` z danymi podsumowania i zwraca odpowiedź `200 OK`.
    b. W przypadku błędu `ErrNoArticlesFound`, renderuje `error.templ` z odpowiednim komunikatem i statusem `404`.
    c. W przypadku błędów systemowych (`ErrAIServiceUnavailable`, `ErrDatabase`), loguje błąd i renderuje `error.templ` z generycznym komunikatem i statusem (`503` lub `500`).

## 6. Względy bezpieczeństwa

- **Uwierzytelnianie**: Endpoint musi być chroniony przez `internal/shared/auth/middleware.go`, który weryfikuje token JWT i zapewnia obecność `user_id`.
- **Autoryzacja**: Wszystkie zapytania do bazy danych (pobieranie feedów, artykułów, zapis podsumowań) muszą być filtrowane przez `user_id` z kontekstu, aby zapobiec dostępowi do danych innych użytkowników.
- **Rate Limiting**: Należy zaimplementować middleware do rate limitingu (np. `golang.org/x/time/rate`) per `user_id`, aby zapobiec nadużyciom kosztownego API AI. Przykładowy limit: 1 podsumowanie na 5 minut na użytkownika.
- **Zarządzanie sekretami**: Klucz API do OpenRouter.ai musi być ładowany z zmiennych środowiskowych za pomocą `internal/shared/config` i nigdy nie może być hardkodowany w kodzie.

## 7. Obsługa błędów

Zdefiniowane zostaną dedykowane typy błędów w pakiecie `summary`, aby umożliwić handlerowi mapowanie ich na odpowiednie kody statusu HTTP.

- `ErrNoArticlesFound` (404): "No articles found from the last 24 hours"
- `ErrAIServiceUnavailable` (503): "AI service is temporarily unavailable"
- `ErrDatabase` (500): "Failed to generate summary. Please try again later."
  Wszystkie błędy 5xx będą logowane za pomocą `slog`.

## 8. Rozważania dotyczące wydajności

- **Wywołanie API AI**: Wywołanie zewnętrznego API jest operacją blokującą i może być czasochłonne. Czekamy 60s na odpowiedź, po czym zwracamy błąd `ErrAIServiceUnavailable`.
- **Generowanie w tle**: W przyszłości można rozważyć przeniesienie generowania podsumowań do zadania w tle (background job), aby natychmiastowo odpowiadać użytkownikowi, a następnie powiadamiać go o zakończeniu generowania (np. przez WebSockets lub odpytywanie po stronie klienta).

## 9. Etapy wdrożenia

1.  **Struktura plików**:
    - Utwórz pliki: `internal/summary/handler.go`, `internal/summary/service.go`, `internal/summary/errors.go`.
    - Utwórz pliki widoków: `internal/summary/view/display.templ`, `internal/summary/view/error.templ`.
2.  **Routing**: W `internal/app/routes.go`, dodaj nową trasę `POST /summaries` w grupie wymagającej autoryzacji, kierując ją do `summaryHandler.GenerateSummary`.
3.  **Błędy**: W `internal/summary/errors.go`, zdefiniuj niestandardowe typy błędów (`ErrNoArticlesFound`, etc.).
4.  **Warstwa serwisu (`SummaryService`)**:
    - Zaimplementuj metodę `GenerateSummary(ctx context.Context) (*database.PublicSummariesSelect, error)`.
    - Dodaj logikę sprawdzania artykułów, zwracając odpowiednie błędy.
    - Zaimplementuj klienta dla OpenRouter.ai w nowym pakiecie (np. `internal/shared/ai/openrouter.go`). Na tym etapie skorzystamy z mocków zamiast rzeczywistego wywołania API.
    - Dodaj logikę wywołania API AI z obsługą timeoutów i błędów.
    - Zaimplementuj logikę zapisu do tabel `summaries` i `events` przy użyciu klienta Supabase.
5.  **Warstwa handlera (`SummaryHandler`)**:
    - Zaimplementuj metodę `GenerateSummary(c echo.Context) error`.
    - Wywołaj `summaryService.GenerateSummary`.
    - Zaimplementuj mapowanie błędów zwróconych z serwisu na odpowiednie odpowiedzi HTTP (renderowanie widoków `display.templ` lub `error.templ`).
6.  **Widoki (Templ)**:
    - W `display.templ`, stwórz komponent, który przyjmuje `SummaryDisplayViewModel` i renderuje tylko nazwę komponentu i modelem widoku.
    - W `error.templ`, stwórz komponent, który przyjmuje `SummaryErrorViewModel` i renderuje tylko nazwę komponentu i modelem widoku.
7.  **Middleware**: Dodaj middleware do rate limitingu dla nowo utworzonego endpointu.
8.  **Dokumentacja**: Upewnij się, że kod jest zgodny z wytycznymi z `AGENTS.md`, a nowe funkcje są udokumentowane w komentarzach, jeśli to konieczne.
