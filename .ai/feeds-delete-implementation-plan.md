# API Endpoint Implementation Plan: DELETE /feeds/{id}

## 1. Przegląd punktu końcowego

Ten punkt końcowy jest odpowiedzialny za usunięcie istniejącego kanału (`feed`) należącego do uwierzytelnionego użytkownika. Usunięcie kanału powoduje również kaskadowe usunięcie wszystkich powiązanych z nim artykułów (`articles`) dzięki ograniczeniom klucza obcego w bazie danych. Po pomyślnym usunięciu, interfejs użytkownika jest informowany o konieczności odświeżenia listy kanałów za pomocą nagłówka HTMX.

## 2. Szczegóły żądania

- **Metoda HTTP:** `DELETE`
- **Struktura URL:** `/feeds/{id}`
- **Parametry:**
  - **Wymagane:**
    - `id` (parametr ścieżki): Unikalny identyfikator (UUID) kanału, który ma zostać usunięty.
  - **Opcjonalne:** Brak.
- **Request Body:** Brak.

## 3. Wykorzystywane typy

- **Command Model:** Nie jest wymagany. `id` jest pobierane z URL, a `user_id` z kontekstu żądania.
- **View Model:**
  - Sukces: Brak (odpowiedź `204 No Content`).
  - Błąd: `feed.ErrorView(message string)` zostanie użyty do wyrenderowania komunikatu o błędzie w komponencie `error.templ`.

## 4. Szczegóły odpowiedzi

- **Sukces (204 No Content):**
  - **Kod stanu:** `204`
  - **Nagłówki:**
    - `HX-Trigger: refreshFeedList` - instruuje HTMX do wywołania zdarzenia, które odświeży listę kanałów.
  - **Ciało odpowiedzi:** Brak.
- **Błędy:**
  - **401 Unauthorized:**
    - Zwracane przez middleware, jeśli użytkownik nie jest uwierzytelniony.
  - **404 Not Found:**
    - **Kod stanu:** `404`
    - **Nagłówki:**
      - `HX-Retarget: #feed-item-{id}-errors`
      - `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Fragment HTML wyrenderowany przez `feed.ErrorView("Feed not found")`.
  - **500 Internal Server Error:**
    - **Kod stanu:** `500`
    - **Nagłówki:**
      - `HX-Retarget: #feed-item-{id}-errors`
      - `HX-Reswap: innerHTML`
    - **Ciało odpowiedzi:** Fragment HTML wyrenderowany przez `feed.ErrorView("Failed to delete feed")`.

## 5. Przepływ danych

1. Użytkownik klika przycisk usuwania, co inicjuje żądanie `DELETE /feeds/{id}`.
2. Middleware `Auth` weryfikuje token JWT, a w przypadku sukcesu umieszcza `user_id` w kontekście żądania (`echo.Context`).
3. `FeedHandler.DeleteFeed` jest wywoływany.
4. Handler pobiera `id` kanału z parametrów URL oraz `user_id` z kontekstu.
5. Handler wywołuje metodę `FeedService.DeleteFeed(ctx, id, userID)`.
6. Serwis wywołuje metodę `FeedRepository.DeleteFeed(ctx, id, userID)`.
7. Repozytorium wykonuje zapytanie `DELETE FROM feeds WHERE id = ? AND user_id = ?`.
8. Jeśli zapytanie nie wpłynie na żaden wiersz (ponieważ `id` nie istnieje lub nie pasuje do `user_id`), repozytorium zwraca błąd `database.ErrNotFound`.
9. Baza danych automatycznie usuwa wszystkie powiązane artykuły dzięki `ON DELETE CASCADE`.
10. Serwis otrzymuje wynik z repozytorium. Jeśli wystąpił błąd `database.ErrNotFound`, propaguje go dalej. W przypadku innych błędów bazy danych, loguje je i zwraca generyczny błąd.
11. Handler na podstawie zwróconego błędu decyduje o odpowiedzi:
    - Brak błędu: wysyła odpowiedź `204 No Content` z nagłówkiem `HX-Trigger`.
    - `database.ErrNotFound`: renderuje widok błędu z komunikatem "Feed not found" i statusem `404`.
    - Inny błąd: renderuje widok błędu z komunikatem "Failed to delete feed" i statusem `500`.

## 6. Względy bezpieczeństwa

- **Uwierzytelnianie:** Dostęp do punktu końcowego musi być chroniony przez middleware `Auth`, który zapewnia, że tylko zalogowani użytkownicy mogą go używać.
- **Autoryzacja (Ochrona przed IDOR):** Logika biznesowa musi rygorystycznie sprawdzać, czy zasób (`feed`) należy do uwierzytelnionego użytkownika. Realizowane jest to poprzez dodanie warunku `user_id = ?` do klauzuli `WHERE` w zapytaniu `DELETE`. Zapobiega to sytuacji, w której użytkownik A usuwa zasoby użytkownika B.

## 7. Etapy wdrożenia

1.  **Repozytorium (`internal/feed/repository.go`):**
    - Dodać nową metodę `DeleteFeed(ctx context.Context, id, userID string) error`.
    - Metoda powinna wykonać zapytanie SQL: `DELETE FROM feeds WHERE id = ? AND user_id = ?`.
    - Jeśli `EXEC` nie zwróci błędu, ale liczba zmienionych wierszy (`RowsAffected`) wynosi 0, metoda powinna zwrócić `database.ErrNotFound`.

2.  **Serwis (`internal/feed/service.go`):**
    - Dodać nową metodę `DeleteFeed(ctx context.Context, id, userID string) error`.
    - Metoda powinna wywołać `repository.DeleteFeed` z przekazanymi argumentami.
    - Powinna obsługiwać błędy zwrócone z repozytorium, logować błędy krytyczne i zwracać je do handlera.

3.  **Handler (`internal/feed/handler.go`):**
    - Dodać nową metodę `DeleteFeed(c echo.Context) error`.
    - Pobrać `user_id` z kontekstu (`c.Get("userID")`).
    - Pobrać `id` z parametrów ścieżki (`c.Param("id")`).
    - Wywołać `service.DeleteFeed`.
    - Zaimplementować logikę obsługi odpowiedzi w zależności od wyniku:
      - Sukces: `c.Response().Header().Set("HX-Trigger", "refreshFeedList")` i `c.NoContent(http.StatusNoContent)`.
      - Błąd `database.ErrNotFound`: `c.Response().Header().Set("HX-Retarget", "#feed-item-"+id+"-errors")`, `c.Response().Header().Set("HX-Reswap", "innerHTML")` i wyrenderować widok błędu z kodem `404`.
      - Inne błędy: `c.Response().Header().Set("HX-Retarget", "#feed-item-"+id+"-errors")`, `c.Response().Header().Set("HX-Reswap", "innerHTML")` i wyrenderować widok błędu z kodem `500`.

4.  **Routing (`internal/app/routes.go`):**
    - W funkcji `setupFeedRoutes` dodać nową trasę:
      ```go
      feedsGroup.DELETE("/:id", feedHandler.DeleteFeed)
      ```
    - Upewnić się, że cała grupa `feedsGroup` jest chroniona przez middleware uwierzytelniający.
