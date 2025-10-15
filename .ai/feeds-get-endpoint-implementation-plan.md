# API Endpoint Implementation Plan: GET /feeds

## 1. Przegląd punktu końcowego

Celem tego punktu końcowego jest dostarczenie listy kanałów (feeds) dla uwierzytelnionego użytkownika. Odpowiedź będzie renderowana jako fragment HTML (partial) i przeznaczona do dynamicznego wstrzyknięcia w interfejsie użytkownika za pomocą htmx. Endpoint obsługuje wyszukiwanie, filtrowanie według statusu oraz paginację.

## 2. Szczegóły żądania

- **Metoda HTTP:** `GET`
- **Struktura URL:** `/feeds`
- **Parametry:**
  - **Opcjonalne:**
    - `search` (string): Fraza do wyszukiwania w nazwach kanałów (case-insensitive).
    - `status` (enum): Filtr statusu ostatniego pobrania. Dopuszczalne wartości: `all`, `working`, `pending`, `error`. Domyślna wartość: `all`.
    - `page` (integer): Numer strony wyników (1-indeksowany). Domyślna wartość: `1`.

- **Request Body:** Brak (żądanie GET).

## 3. Wykorzystywane typy

- **Query Model (do utworzenia w `internal/feed/models/queries.go`):**

  ```go
  package models

  type ListFeedsQuery struct {
      UserID string
      Search string
      Status string
      Page   int

  }
  ```

- **View Models (istniejące w `internal/feed/models/dto.go`):**
  - `FeedListViewModel`
  - `FeedItemViewModel`
  - `PaginationViewModel`

## 4. Szczegóły odpowiedzi

- **Sukces (200 OK):**
  - **Content-Type:** `text/html; charset=utf-8`
  - **Body:** Fragment HTML wyrenderowany przez komponent `feed/view/list.templ` zawierający listę kanałów i kontrolki paginacji.
- **Błędy:**
  - **400 Bad Request:** Fragment HTML z komunikatem błędu, np. "Nieprawidłowe parametry zapytania".
  - **401 Unauthorized:** Przekierowanie do strony logowania lub fragment HTML z komunikatem o braku autoryzacji.
  - **500 Internal Server Error:** Fragment HTML z komunikatem błędu, np. "Nie udało się załadować kanałów".

## 5. Przepływ danych

1.  Użytkownik wysyła żądanie `GET /feeds` z opcjonalnymi parametrami.
2.  Middleware `Auth` weryfikuje sesję użytkownika. Jeśli jest nieprawidłowa, zwraca `401 Unauthorized`. W przeciwnym razie, dodaje `user_id` do kontekstu żądania.
3.  `feed.Handler` parsuje parametry zapytania (`search`, `status`, `page`).
4.  Handler waliduje parametry: `status` musi należeć do dozwolonego enuma, `page` musi być poprawną liczbą (>= 1). W przypadku błędu walidacji zwraca `400 Bad Request`.
5.  Handler tworzy obiekt `ListFeedsQuery` z `user_id` z kontekstu i zwalidowanymi parametrami.
6.  Handler wywołuje metodę `feed.Service.ListFeeds(ctx, query)`.
7.  Serwis konstruuje zapytanie SQL do tabeli `feeds` na podstawie `ListFeedsQuery`:
    - Dodaje warunek `WHERE user_id = ?`.
    - Jeśli `search` jest podany, dodaje warunek `AND name ILIKE ?`.
    - Jeśli `status` to `working` lub `error`, dodaje odpowiedni warunek `AND last_fetch_status`.
8.  Serwis wykonuje dwa zapytania do bazy danych:
    - `COUNT(*)` z tymi samymi warunkami, aby uzyskać całkowitą liczbę pasujących rekordów.
    - `SELECT ...` z `LIMIT` i `OFFSET` (obliczonym na podstawie stałego rozmiaru strony i `page`), aby pobrać dane dla bieżącej strony.
9.  Serwis oblicza dane paginacji (`TotalPages`, `HasPrevious`, `HasNext`).
10. Serwis mapuje wyniki z bazy danych na `[]FeedItemViewModel` i konstruuje finalny `FeedListViewModel`.
11. Handler otrzymuje `FeedListViewModel` od serwisu.
12. Handler renderuje komponent `feed/view/list.templ`, przekazując do niego otrzymany view model.
13. Serwer wysyła odpowiedź HTML z kodem `200 OK`.

## 6. Względy bezpieczeństwa

- **Uwierzytelnianie i Autoryzacja:** Dostęp do endpointu musi być chroniony przez middleware, który weryfikuje tożsamość użytkownika.
- **Kontrola dostępu:** Kluczowe jest, aby każde zapytanie do bazy danych zawierało warunek `WHERE user_id = ?`, gdzie `user_id` pochodzi z zaufanego źródła (kontekstu żądania po przejściu przez middleware autoryzacji). Zapobiegnie to dostępowi do danych innych użytkowników.
- **Walidacja danych wejściowych:** Rygorystyczna walidacja parametrów `page` i `status` chroni przed nieoczekiwanym zachowaniem i potencjalnymi atakami. Użycie zparametryzowanych zapytań SQL jest obowiązkowe, aby zapobiec SQL Injection (szczególnie dla pola `search`).

## 7. Etapy wdrożenia

1.  **Model:** Utworzyć plik `internal/feed/models/queries.go` i zdefiniować w nim struct `ListFeedsQuery`.
2.  **Baza danych:** Zdefiniować metody dostępu do danych w warstwie bazy danych, które będą przyjmować parametry filtrowania i paginacji.
3.  **Serwis:**
    - Utworzyć plik `internal/feed/service.go` (jeśli nie istnieje) i zdefiniować interfejs `FeedService`.
    - Zaimplementować metodę `ListFeeds(ctx context.Context, query ListFeedsQuery) (*FeedListViewModel, error)`. Metoda ta będzie zawierać logikę biznesową pobierania, filtrowania i mapowania danych.
4.  **Widok:** Utworzyć plik `internal/feed/view/list.templ`. Komponent ten będzie przyjmował `FeedListViewModel` i renderował prosty debug view z nazwą komponentu oraz view modelem jako string.
5.  **Handler:**
    - W pliku `internal/feed/handler.go` zaimplementować metodę `HandleListFeeds(c echo.Context) error`.
    - Dodać logikę parsowania i walidacji parametrów wejściowych.
    - Dodać wywołanie serwisu `feed.Service.ListFeeds`.
    - Dodać obsługę błędów z serwisu i renderowanie odpowiednich widoków błędów.
    - Dodać renderowanie widoku sukcesu (`list.templ`).
6.  **Routing:** W pliku `internal/app/routes.go` zarejestrować nową trasę `GET /feeds` w grupie wymagającej autoryzacji, mapując ją do `feed.Handler.HandleListFeeds`.
