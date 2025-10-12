# API Endpoint Implementation Plan: GET /feeds/{id}/edit

## 1. Przegląd punktu końcowego

Celem tego punktu końcowego jest wyrenderowanie i zwrócenie formularza HTML, który pozwala użytkownikowi na edycję istniejącego kanału (feed). Formularz zostanie wstępnie wypełniony aktualnymi danymi kanału (nazwa i URL). Dostęp do tego punktu końcowego jest ograniczony wyłącznie do właściciela kanału.

## 2. Szczegóły żądania

- **Metoda HTTP:** `GET`
- **Struktura URL:** `/feeds/{id}/edit`
- **Parametry:**
  - **Wymagane:**
    - `id` (parametr ścieżki): Unikalny identyfikator (UUID) kanału, który ma być edytowany.
  - **Opcjonalne:** Brak
- **Request Body:** Brak (żądanie GET).

## 3. Wykorzystywane typy

- **View Model:**
  - `models.FeedEditFormViewModel`: Struktura danych przekazywana do szablonu `templ` w celu renderowania formularza.
    ```go
    type FeedEditFormViewModel struct {
        FeedID string `json:"feed_id"`
        Name   string `json:"name"`
        URL    string `json:"url"`
    }
    ```

## 4. Szczegóły odpowiedzi

- **Odpowiedź sukcesu (200 OK):**
  - **Content-Type:** `text/html; charset=utf-8`
  - **Body:** Fragment kodu HTML zawierający formularz edycji, wyrenderowany przez szablon `templ`, z polami `name` i `url` wypełnionymi danymi kanału.
- **Odpowiedzi błędów:**
  - **401 Unauthorized:** Zwracane, gdy użytkownik nie jest uwierzytelniony. Odpowiedź będzie renderowana przez middleware autoryzacji.
  - **404 Not Found:** Zwracane, gdy kanał o podanym `id` nie istnieje lub nie należy do uwierzytelnionego użytkownika. Renderuje fragment błędu z komunikatem "Feed not found".
  - **500 Internal Server Error:** Zwracane w przypadku błędów serwera (np. błąd bazy danych). Renderuje fragment błędu z komunikatem "Failed to load feed".

## 5. Przepływ danych

1. Użytkownik klika link/przycisk "Edytuj", co powoduje wysłanie żądania `GET` na adres `/feeds/{id}/edit`.
2. Żądanie jest przechwytywane przez middleware `auth`, który weryfikuje sesję użytkownika i umieszcza `user_id` w kontekście żądania.
3. Routing Echo kieruje żądanie do handlera `HandleFeedEditForm` w `internal/feed/handler.go`.
4. Handler wyodrębnia `id` kanału z parametrów ścieżki oraz `user_id` z kontekstu.
5. Handler wywołuje metodę serwisu `feed.Service.GetFeedForEdit(ctx, id, userID)`.
6. Serwis wywołuje metodę repozytorium, np. `feed.Repository.FindFeedByIDAndUser(ctx, id, userID)`, aby pobrać dane kanału z bazy danych.
7. Repozytorium wykonuje zapytanie SQL: `SELECT id, name, url FROM feeds WHERE id = $1 AND user_id = $2`.
8. Jeśli zapytanie nie zwróci żadnych wyników, repozytorium zwraca błąd typu `sql.ErrNoRows`, który serwis mapuje na dedykowany błąd (np. `ErrFeedNotFound`).
9. Jeśli dane zostaną znalezione, serwis mapuje wynik z bazy danych na `models.FeedEditFormViewModel` za pomocą `models.NewFeedEditFormFromDB`.
10. Handler otrzymuje model widoku od serwisu.
11. Handler renderuje szablon `internal/feed/view/edit_form.templ`, przekazując do niego otrzymany model widoku.
12. Wyrenderowany kod HTML jest wysyłany jako odpowiedź HTTP z kodem statusu 200.

## 6. Względy bezpieczeństwa

- **Autoryzacja:** Dostęp musi być ściśle kontrolowany. Zapytanie do bazy danych musi zawierać zarówno `id` kanału, jak i `user_id` pobrany z uwierzytelnionej sesji. Zapobiega to atakom typu IDOR, gdzie użytkownik mógłby próbować uzyskać dostęp do zasobów innego użytkownika.
- **Walidacja danych wejściowych:** Parametr `id` powinien być walidowany jako prawidłowy format UUID, aby zapobiec błędom w zapytaniach do bazy danych.

## 7. Obsługa błędów

- **Brak uwierzytelnienia (401):** Obsługiwane globalnie przez middleware autoryzacji.
- **Nie znaleziono zasobu (404):**
  - **Scenariusz:** `feed.Service.GetFeedForEdit` zwraca błąd `ErrFeedNotFound`.
  - **Akcja:** Handler renderuje szablon błędu (`shared/view/error_fragment.templ`) z kodem statusu 404 i komunikatem "Feed not found".
- **Błąd serwera (500):**
  - **Scenariusz:** Błąd połączenia z bazą danych, błąd renderowania szablonu lub inny nieoczekiwany błąd w serwisie lub handlerze.
  - **Akcja:** Błąd jest logowany za pomocą `slog`. Handler renderuje szablon błędu z kodem statusu 500 i komunikatem "Failed to load feed".

## 8. Etapy wdrożenia

1. **Routing:** W pliku `internal/app/routes.go` dodać nową trasę w grupie `/feeds`:
   ```go
   feedsGroup.GET("/:id/edit", feedHandler.HandleFeedEditForm)
   ```
2. **Widok (Templ):** Utworzyć nowy plik `internal/feed/view/edit_form.templ` zawierający minimalny debug view z nazwą komponentu oraz view modelem jako string.
3. **Repozytorium:** Jeśli nie istnieje, dodać metodę `FindFeedByIDAndUser` do `internal/feed/repository.go`, która pobiera pojedynczy kanał na podstawie `id` i `user_id`.
4. **Serwis:** Zaimplementować metodę `GetFeedForEdit(ctx, feedID, userID)` w `internal/feed/service.go`. Metoda ta będzie korzystać z repozytorium do pobrania danych i mapowania ich na `FeedEditFormViewModel`.
5. **Handler:** Zaimplementować metodę `HandleFeedEditForm(c echo.Context)` w `internal/feed/handler.go`. Metoda będzie:
   - Pobierać `id` i `user_id`.
   - Wywoływać serwis `GetFeedForEdit`.
   - Obsługiwać błędy (404, 500).
   - Renderować szablon `FeedEditForm` w przypadku sukcesu lub szablon błędu w przypadku porażki.
