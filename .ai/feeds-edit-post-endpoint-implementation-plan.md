# API Endpoint Implementation Plan: PATCH /feeds/{id}

## 1. Przegląd punktu końcowego

Ten punkt końcowy obsługuje aktualizację istniejącego kanału (feed) na podstawie jego ID. Jest przeznaczony do użytku z htmx, aby zapewnić dynamiczne aktualizacje interfejsu użytkownika bez przeładowywania całej strony. W przypadku powodzenia, wyzwala odświeżenie listy kanałów. W przypadku błędu, wyświetla komunikaty o błędach w formularzu edycji.

## 2. Szczegóły żądania

- **Metoda HTTP:** `PATCH`
- **Struktura URL:** `/feeds/{id}`
- **Parametry:**
  - **Ścieżki (Path):**
    - `id` (string, UUID): Wymagany identyfikator kanału do aktualizacji.
  - **Ciała Żądania (Request Body):** `multipart/form-data`
    - `name` (string): Wymagany, niepusty. Nowa nazwa dla kanału.
    - `url` (string): Wymagany, prawidłowy format URL. Nowy adres URL dla kanału.

## 3. Wykorzystywane typy

- **Command Model:** `models.UpdateFeedCommand`
  - Służy do walidacji i powiązania danych przychodzących z formularza.
  - Zawiera metody `ToUpdate()` i `ToUpdateWithURLChange()` do generowania odpowiednich struktur do aktualizacji bazy danych w zależności od tego, czy URL uległ zmianie.
- **View Model (dla błędów):** `models.FeedFormErrorViewModel`
  - Używany do renderowania komunikatów o błędach walidacji lub innych błędach operacji w komponencie `view.FormError`.
- **Templ Component:** `feed/view.FormError`
  - Renderuje błędy formularza w odpowiedzi na żądanie htmx.

## 4. Szczegóły odpowiedzi

- **Odpowiedź sukcesu (Success):**
  - **Kod stanu:** `204 No Content`
  - **Nagłówki:** `HX-Trigger: refreshFeedList`
  - **Ciało odpowiedzi:** Puste.
- **Odpowiedzi błędów (Error):**
  - **Kod stanu:** `400 Bad Request` (Błąd walidacji)
    - **Nagłówki:** `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Renderowany komponent `feed/view.FormError` z `FeedFormErrorViewModel` zawierającym błędy pól.
  - **Kod stanu:** `404 Not Found` (Kanał nie znaleziony)
    - **Nagłówki:** `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Renderowany komponent `feed/view.FormError` z `FeedFormErrorViewModel` i komunikatem "Feed not found".
  - **Kod stanu:** `409 Conflict` (URL już istnieje)
    - **Nagłówki:** `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Renderowany komponent `feed/view.FormError` z `FeedFormErrorViewModel` i błędem `URLError`.
  - **Kod stanu:** `500 Internal Server Error` (Błąd serwera)
    - **Nagłówki:** `HX-Retarget: #feed-edit-form-errors-{id}`, `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Renderowany komponent `feed/view.FormError` z `FeedFormErrorViewModel` i błędem `GeneralError`.

## 5. Przepływ danych

1.  Użytkownik przesyła formularz edycji kanału.
2.  Middleware `auth` weryfikuje sesję użytkownika i dołącza `userID` do kontekstu żądania.
3.  Handler `feed.HandleUpdate` jest wywoływany.
4.  Handler pobiera `id` z parametrów ścieżki i `userID` z kontekstu.
5.  Dane z formularza są bindowane i walidowane do struktury `models.UpdateFeedCommand`.
6.  Handler wywołuje metodę `feed.Service.UpdateFeed` z `id`, `userID` i `UpdateFeedCommand`.
7.  Serwis `feed.Service`:
    a. Pobiera istniejący kanał z repozytorium, aby zweryfikować własność i istnienie.
    b. Porównuje przychodzący URL z istniejącym.
    c. **Jeśli URL się zmienił:**
    i. Sprawdza, czy nowy URL jest już zajęty przez innego feeda tego samego użytkownika.
    ii. Jeśli jest unikalny, wywołuje repozytorium z poleceniem `cmd.ToUpdateWithURLChange()`, które resetuje pola związane z pobieraniem (`last_fetch_status`, `etag`, etc.) i ustawia `fetch_after`.
    d. **Jeśli URL się nie zmienił:**
    i. Wywołuje repozytorium z poleceniem `cmd.ToUpdate()`, które aktualizuje tylko nazwę.
    e. Zwraca odpowiedni błąd (`ErrFeedNotFound`, `ErrFeedURLConflict`) lub `nil`.
8.  Handler na podstawie wyniku z serwisu:
    a. **Sukces:** Zwraca `204 No Content` z nagłówkiem `HX-Trigger`.
    b. **Błąd:** Renderuje komponent `FormError` z odpowiednim kodem statusu i nagłówkami htmx (wszystkie błędy używają tego samego komponentu).

## 6. Względy bezpieczeństwa

- **Uwierzytelnianie:** Punkt końcowy musi być chroniony przez middleware `auth`, aby zapewnić, że tylko zalogowani użytkownicy mogą z niego korzystać.
- **Autoryzacja:** Logika biznesowa w `feed.Service` musi rygorystycznie sprawdzać, czy `user_id` kanału, który ma być zaktualizowany, pasuje do `user_id` zautoryzowanego użytkownika. Zapobiega to modyfikowaniu zasobów innych użytkowników.
- **Walidacja danych wejściowych:**
  - `id` musi być walidowane jako poprawny UUID.
  - `name` i `url` muszą być walidowane zgodnie z regułami w `UpdateFeedCommand` (`required`, `max=255`, `url`), aby zapobiec nieprawidłowym danym i potencjalnym atakom (np. XSS, chociaż renderowanie po stronie serwera z Templ zmniejsza to ryzyko).

## 7. Etapy wdrożenia

1.  **Routing:** Zarejestruj trasę `PATCH /feeds/:id` w `internal/app/routes.go`, przypisując ją do `feed.Handler.HandleUpdate` i stosując middleware `auth`.
2.  **Handler (`feed.Handler`):**
    - Zaimplementuj metodę `HandleUpdate(c echo.Context) error`.
    - Pobierz `id` z `c.Param("id")` i `userID` z kontekstu.
    - Zbinduj i zwaliduj dane formularza do `models.UpdateFeedCommand`. W przypadku błędu walidacji, renderuj `view.FormError` z kodem 400.
    - Wywołaj `h.service.UpdateFeed`.
    - Obsłuż błędy zwrócone z serwisu (`ErrFeedNotFound`, `ErrFeedURLConflict`, błędy ogólne) i renderuj odpowiednie widoki błędów z kodami 404, 409, 500.
    - W przypadku sukcesu, zwróć `c.NoContent(http.StatusNoContent)` i ustaw nagłówek `HX-Trigger`.
3.  **Service (`feed.Service`):**
    - Zaimplementuj metodę `UpdateFeed(ctx context.Context, id string, userID string, cmd models.UpdateFeedCommand) error`.
    - Dodaj nową metodę do interfejsu repozytorium: `IsURLTaken(ctx context.Context, userID string, url string, excludeFeedID string) (bool, error)`.
    - W `UpdateFeed`, pobierz kanał za pomocą istniejącej metody `FindFeedByIDAndUser`. Jeśli nie zostanie znaleziony, zwróć `ErrFeedNotFound`.
    - Jeśli `cmd.URL` różni się od URL-a pobranego kanału:
      - Wywołaj `IsURLTaken`. Jeśli `true`, zwróć `ErrFeedURLConflict`.
      - Wywołaj `repo.UpdateFeed(id, cmd.ToUpdateWithURLChange())`.
    - W przeciwnym razie, wywołaj `repo.UpdateFeed(id, cmd.ToUpdate())`.
    - Zwróć błąd z repozytorium, jeśli wystąpi.
4.  **Repository (`feed.Repository`):**
    - Użyj istniejącej metody `FindFeedByIDAndUser` - już implementuje zapytanie `SELECT` z warunkiem `WHERE id = ? AND user_id = ?`.
    - Zaimplementuj `IsURLTaken` - zapytanie z warunkiem `WHERE user_id = ? AND url = ? AND id != ?`.
    - Zaimplementuj `UpdateFeed(ctx context.Context, feedID string, update database.PublicFeedsUpdate)` - wykonuje UPDATE z warunkiem `WHERE id = ?`.
    - Metoda `UpdateFeed` musi obsługiwać struktury generowane przez `ToUpdate()` i `ToUpdateWithURLChange()`.
