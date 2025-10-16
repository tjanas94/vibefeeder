# API Endpoint Implementation Plan: GET /summaries/latest

## 1. Przegląd punktu końcowego

Celem tego punktu końcowego jest pobranie i wyświetlenie najnowszego podsumowania wygenerowanego dla uwierzytelnionego użytkownika. Jeśli użytkownik nie ma żadnych podsumowań, zostanie wyświetlony stan pusty z opcją wygenerowania nowego podsumowania, pod warunkiem że ma co najmniej jeden poprawnie skonfigurowany kanał RSS.

## 2. Szczegóły żądania

- **Metoda HTTP**: `GET`
- **Struktura URL**: `/summaries/latest`
- **Parametry**:
  - **Wymagane**: Brak.
  - **Opcjonalne**: Brak.
- **Request Body**: Brak.

## 3. Wykorzystywane typy

- **`models.SummaryDisplayViewModel`**: Główny model widoku używany do przekazania danych do szablonu `templ`.
  ```go
  type SummaryDisplayViewModel struct {
      Summary      *SummaryViewModel
      CanGenerate  bool
      ErrorMessage string
  }
  ```
- **`models.SummaryViewModel`**: Reprezentuje dane pojedynczego podsumowania.
  ```go
  type SummaryViewModel struct {
      ID        string
      Content   string
      CreatedAt time.Time
  }
  ```

## 4. Szczegóły odpowiedzi

- **Odpowiedź sukcesu (200 OK)**:
  - Zwraca kod stanu `200 OK`.
  - Renderuje komponent `summary.view.Display` z modelem `SummaryDisplayViewModel`.
    - Jeśli podsumowanie istnieje, `Summary` zawiera jego dane.
    - Jeśli podsumowanie nie istnieje, `Summary` jest `nil`.
- **Odpowiedzi błędów**:
  - **`401 Unauthorized`**: Zwracane przez middleware, jeśli użytkownik nie jest uwierzytelniony.
  - **`500 Internal Server Error`**: Zwracane, gdy wystąpi błąd serwera (np. błąd zapytania do bazy danych). Renderuje komponent `summary.view.Display` z wypełnionym `ErrorMessage`.

## 5. Przepływ danych

1.  Żądanie `GET` trafia na endpoint `/summaries/latest`.
2.  Middleware `auth` weryfikuje tożsamość użytkownika i umieszcza `userID` w `echo.Context`.
3.  Handler `summary.GetLatestSummary` jest wywoływany.
4.  Handler pobiera `userID` z kontekstu.
5.  Handler wywołuje metodę serwisu `summary.Service.GetLatestSummaryForUser(ctx, userID)`.
6.  Metoda serwisowa wykonuje dwa zapytania do bazy danych:
    a. Pobiera najnowsze podsumowanie dla `userID` (`SELECT * FROM summaries WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`).
    b. Sprawdza, czy użytkownik ma co najmniej jeden kanał z `last_fetch_status = 'success'` (`SELECT 1 FROM feeds WHERE user_id = ? AND last_fetch_status = 'success' LIMIT 1`).
7.  Serwis konstruuje i zwraca `models.SummaryDisplayViewModel` na podstawie wyników zapytań.
8.  Handler odbiera model widoku od serwisu.
9.  Jeśli wystąpił błąd, handler loguje go i renderuje widok błędu z kodem `500`.
10. Jeśli operacja się powiodła, handler renderuje widok `summary.view.Display` z otrzymanym modelem i wysyła odpowiedź HTML z kodem `200`.

## 6. Względy bezpieczeństwa

- **Uwierzytelnianie**: Dostęp do endpointu musi być chroniony przez middleware, który weryfikuje sesję użytkownika (np. za pomocą tokenu JWT).
- **Autoryzacja**: Wszystkie zapytania do bazy danych muszą być ściśle powiązane z `userID` pobranym z uwierzytelnionego kontekstu sesji. Zapobiega to możliwości odczytania danych należących do innego użytkownika.

## 7. Rozważania dotyczące wydajności

- **Indeksowanie bazy danych**: Aby zapewnić szybkie działanie zapytań, należy upewnić się, że istnieją odpowiednie indeksy w bazie danych:
  - W tabeli `summaries` na kolumnach `(user_id, created_at)`.
  - W tabeli `feeds` na kolumnach `(user_id, last_fetch_status)`.

## 8. Etapy wdrożenia

1.  **Baza danych**:
    - Zweryfikować istnienie indeksów na `summaries(user_id, created_at DESC)` i `feeds(user_id, last_fetch_status)`. Jeśli nie istnieją, dodać odpowiednie migracje.
2.  **Warstwa serwisu (`internal/summary/service.go`)**:
    - Zaimplementować nową metodę `GetLatestSummaryForUser(ctx context.Context, userID string) (*models.SummaryDisplayViewModel, error)`.
    - Wewnątrz metody dodać logikę do pobierania najnowszego podsumowania z bazy danych.
    - Dodać logikę do sprawdzania, czy użytkownik może generować nowe podsumowania.
    - Zmapować wyniki na `SummaryDisplayViewModel`.
3.  **Warstwa handlera (`internal/summary/handler.go`)**:
    - Zaimplementować nową metodę `GetLatestSummary(c echo.Context) error`.
    - Wewnątrz metody pobrać `userID` z `c.Get("user_id")`.
    - Wywołać metodę serwisu `GetLatestSummaryForUser`.
    - Obsłużyć przypadki sukcesu i błędu, renderując odpowiednie widoki (`display.templ` lub `error.templ`).
4.  **Routing (`internal/app/routes.go`)**:
    - Dodać nową trasę `GET "/summaries/latest"` w grupie tras wymagających uwierzytelnienia.
    - Powiązać trasę z nowo utworzoną metodą handlera `summary.GetLatestSummary`.
5.  **Warstwa widoku (`internal/summary/view/display.templ`)**:
    - Na tym etapie tworzymy prosty debug view z nazwą komponentu i modelem widoku w stringu.
