# API Endpoint Implementation Plan: POST /feeds

## 1. Przegląd punktu końcowego

Ten punkt końcowy obsługuje tworzenie nowego kanału RSS dla uwierzytelnionego użytkownika. Po pomyślnym dodaniu kanału, inicjuje odświeżenie listy kanałów po stronie klienta za pomocą mechanizmu htmx. Endpoint jest zaprojektowany z myślą o interakcjach formularzowych i zwraca szczegółowe błędy walidacji.

## 2. Szczegóły żądania

- **Metoda HTTP**: `POST`
- **Struktura URL**: `/feeds`
- **Parametry**: Brak parametrów query.
- **Request Body**: `multipart/form-data`
  - `name`: `string` (wymagane, niepuste)
  - `url`: `string` (wymagane, prawidłowy format URL)

## 3. Wykorzystywane typy

- **Command Model**: `models.CreateFeedCommand`
  - Służy do powiązania i walidacji danych przychodzących z formularza.
- **View Model**: `models.FeedFormErrorViewModel`
  - Używany do renderowania komunikatów o błędach walidacji lub innych błędach w kontekście formularza.
- **View**: `feed.FeedFormErrors` (komponent Templ)
  - Renderuje `FeedFormErrorViewModel` do HTML.

## 4. Szczegóły odpowiedzi

- **Odpowiedź sukcesu**:
  - **Kod statusu**: `204 No Content`
  - **Nagłówki**: `HX-Trigger: refreshFeedList`
  - **Treść**: Pusta.
- **Odpowiedzi błędów**:
  - **Kod statusu**: `400 Bad Request` (Błąd walidacji)
    - **Nagłówki**: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
    - **Treść**: HTML wyrenderowany z `feed.FeedFormErrors` z błędami walidacji (`NameError` lub `URLError`).
  - **Kod statusu**: `409 Conflict` (Kanał już istnieje)
    - **Nagłówki**: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
    - **Treść**: HTML wyrenderowany z `feed.FeedFormErrors` z błędem `URLError = "You have already added this feed"`.
  - **Kod statusu**: `500 Internal Server Error` (Błąd serwera)
    - **Nagłówki**: `HX-Retarget: #feed-add-form-errors`, `HX-Reswap: innerHTML`
    - **Treść**: HTML wyrenderowany z `feed.FeedFormErrors` z błędem `GeneralError`.

## 5. Przepływ danych

1. Użytkownik wysyła formularz dodawania kanału na `POST /feeds`.
2. **Middleware**: Żądanie przechodzi przez middleware uwierzytelniający, który weryfikuje sesję użytkownika i umieszcza `user_id` w kontekście żądania.
3. **Handler (`postFeedHandler`)**:
   a. Pobiera `user_id` z kontekstu.
   b. Wiąże dane z formularza do instancji `models.CreateFeedCommand`.
   c. Waliduje obiekt komendy za pomocą `shared/validator`. W przypadku błędu, renderuje widok błędu z kodem `400`.
   d. Wywołuje metodę `feed.Service.CreateFeed` z komendą i `user_id`.
4. **Serwis (`feed.Service`)**:
   a. Odbiera komendę i `user_id`.
   b. Wywołuje repozytorium w celu wstawienia nowego rekordu do tabeli `feeds`, ustawiając `fetch_after` na 5 minut w przyszłości.
   c. Przechwytuje ewentualny błąd naruszenia unikalności z bazy danych i zwraca go jako specyficzny typ błędu (np. `ErrFeedAlreadyExists`).
   d. Jeśli wstawienie się powiedzie, wywołuje repozytorium w celu zapisania zdarzenia `feed_added` w tabeli `events`.
5. **Handler (`postFeedHandler`)**:
   a. Jeśli serwis zwróci błąd `ErrFeedAlreadyExists`, renderuje widok błędu z kodem `409`.
   b. Jeśli serwis zwróci inny błąd, renderuje ogólny widok błędu z kodem `500`.
   c. Jeśli operacja się powiedzie, wysyła odpowiedź `204 No Content` z nagłówkiem `HX-Trigger`.

## 6. Względy bezpieczeństwa

- **Uwierzytelnienie**: Endpoint musi być chroniony przez middleware, który zapewnia, że tylko zalogowani użytkownicy mogą dodawać kanały. `user_id` musi być pobierany z zaufanego źródła (kontekst po middleware), a nie z żądania.
- **Autoryzacja**: Logika biznesowa operuje w kontekście `user_id`, zapewniając, że użytkownicy mogą modyfikować tylko własne dane.
- **CSRF**: Należy zastosować middleware anty-CSRF, aby chronić endpoint przed atakami Cross-Site Request Forgery.
- **Walidacja danych**: Rygorystyczna walidacja pól `name` i `url` jest kluczowa. Użycie biblioteki `go-playground/validator` zapewnia solidną podstawę.
- **SQL Injection**: Zapytania do bazy danych muszą być wykonywane za pomocą parametryzowanych zapytań dostarczanych przez klienta Supabase, aby uniknąć luk bezpieczeństwa.

## 7. Etapy wdrożenia

1. **Repozytorium (`internal/feed/repository.go`)**:
   - Upewnij się, że istnieje metoda do wstawiania rekordu do tabeli `feeds`.
   - Dodaj metodę do wstawiania rekordu do tabeli `events`.
2. **Serwis (`internal/feed/service.go`)**:
   - Zdefiniuj stałą dla błędu `ErrFeedAlreadyExists`.
   - Zaimplementuj metodę `CreateFeed(ctx context.Context, cmd models.CreateFeedCommand, userID string) error`.
   - W metodzie `CreateFeed`, obsłuż błąd unikalności z repozytorium i zwróć `ErrFeedAlreadyExists`.
   - Po pomyślnym dodaniu kanału, wywołaj metodę repozytorium do zapisu zdarzenia.
3. **Widok (`internal/feed/view/`)**:
   - Utwórz plik `form_error.templ`, który będzie zawierał komponent `FeedFormErrors(vm models.FeedFormErrorViewModel)`. Komponent ten będzie renderował błędy w odpowiednim kontenerze HTML. Na razie minimalny debug view z nazwą komponentu i view modelem jako string.
4. **Handler (`internal/feed/handler.go`)**:
   - Zaimplementuj `postFeedHandler(c echo.Context) error`.
   - W handlerze, zintegruj logikę walidacji, wywołania serwisu i renderowania odpowiedzi (sukces `204` lub błędy `400`/`409`/`500` z użyciem komponentu `FeedFormErrors`).
5. **Routing (`internal/app/routes.go`)**:
   - Dodaj trasę `POST /feeds` do grupy tras wymagających uwierzytelnienia.
   - Przypisz `postFeedHandler` do tej trasy.
