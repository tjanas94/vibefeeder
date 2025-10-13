# Plan Implementacji Punktu Końcowego API: GET /dashboard

## 1. Przegląd Punktu Końcowego

Celem tego punktu końcowego jest dostarczenie głównego widoku aplikacji ("dashboard"). Widok ten pełni rolę szkieletu strony, który zawiera podstawowy layout, formularze filtrów oraz puste kontenery. Treść tych kontenerów (dane użytkownika, lista kanałów RSS, ostatnie podsumowanie) jest dynamicznie ładowana za pomocą oddzielnych żądań htmx po załadowaniu strony. Endpoint wymaga uwierzytelnienia.

## 2. Szczegóły Żądania

- **Metoda HTTP:** `GET`
- **Struktura URL:** `/dashboard`
- **Parametry:**
  - **Wymagane:** Brak
  - **Opcjonalne:** Brak
- **Request Body:** Brak

## 3. Wykorzystywane Typy

- **`models.DashboardViewModel{}`**: Pusty model widoku przekazywany do szablonu Templ. Jego rola jest symboliczna, ponieważ wszystkie dane są ładowane przez htmx z innych punktów końcowych:
  - `#user-info-container` ← `GET /auth/me`
  - `#feed-list-container` ← `GET /feeds`
  - `#summary-container` ← `GET /summaries/latest`

## 4. Szczegóły Odpowiedzi

- **Sukces (200 OK):**
  - **Content-Type:** `text/html; charset=utf-8`
  - **Body:** Pełna strona HTML wyrenderowana przez szablon `dashboard/view/index.templ`, zawierająca główny layout i puste kontenery gotowe do wypełnienia przez htmx.
- **Błędy:**
  - **401 Unauthorized:**
    - **Nagłówek:** `Location: /auth/login`
    - **Body:** Puste. Odpowiedź inicjuje przekierowanie po stronie klienta na stronę logowania.
  - **500 Internal Server Error:**
    - **Content-Type:** `text/html; charset=utf-8`
    - **Body:** Strona błędu z komunikatem "Failed to load dashboard. Please refresh the page."

## 5. Przepływ Danych

1. Użytkownik wysyła żądanie `GET /dashboard`.
2. Middleware uwierzytelniający (`internal/shared/auth/middleware.go`) przechwytuje żądanie i weryfikuje sesję użytkownika.
3. Jeśli użytkownik jest uwierzytelniony, żądanie trafia do `DashboardHandler`.
4. Handler `ShowDashboard` wywołuje renderowanie szablonu `dashboard/view/index.templ` z pustym modelem `DashboardViewModel`.
5. Szablon Templ generuje pełny kod HTML strony.
6. Wygenerowany HTML jest zwracany do przeglądarki z kodem statusu 200 OK.
7. Po załadowaniu strony, atrybuty `hx-get` w kontenerach inicjują kolejne żądania do API w celu pobrania dynamicznej treści.

## 6. Względy Bezpieczeństwa

- **Uwierzytelnianie:** Dostęp do punktu końcowego `/dashboard` musi być chroniony przez middleware, który weryfikuje, czy użytkownik jest zalogowany. W przypadku braku aktywnej sesji, middleware musi przerwać przetwarzanie żądania i zwrócić odpowiedź 401 z nagłówkiem `Location` w celu przekierowania na stronę logowania.
- **Walidacja Danych:** Brak danych wejściowych od użytkownika, więc walidacja nie jest wymagana.

## 7. Obsługa Błędów

- **Brak uwierzytelnienia (401):** Obsługiwane centralnie przez middleware uwierzytelniający.
- **Błąd renderowania widoku (500):** Jeśli wystąpi błąd podczas renderowania szablonu Templ, handler powinien go przechwycić, zalogować szczegóły błędu przy użyciu standardowego loggera (`slog`), a następnie zwrócić generyczną stronę błędu 500.

## 8. Etapy Wdrożenia

1. **Utworzenie struktury katalogów:**
   - Utwórz nowy katalog `internal/dashboard`.
   - Wewnątrz `internal/dashboard` utwórz podkatalogi: `handler`, `view`.

2. **Zdefiniowanie modelu widoku:**
   - Plik `internal/dashboard/models/dto.go` już istnieje i zawiera pustą strukturę `DashboardViewModel`. Nie są wymagane żadne zmiany.

3. **Implementacja widoku (Templ):**
   - Utwórz plik `internal/dashboard/view/index.templ`.
   - Zdefiniuj w nim komponent `Index(vm models.DashboardViewModel)`, który renderuje debug view z nazwą komponentu oraz view modelem jako string.
   - Wewnątrz layoutu umieść puste kontenery z odpowiednimi atrybutami `id` oraz `hx-*` do ładowania treści:
     ```html
     <div id="feed-list-container" hx-get="/feeds" hx-trigger="load"></div>
     <div id="summary-container" hx-get="/summaries/latest" hx-trigger="load"></div>
     ```

4. **Implementacja Handlera:**
   - Utwórz plik `internal/dashboard/handler/handler.go`.
   - Zdefiniuj strukturę `DashboardHandler` z zależnościami (np. `echo.Renderer`).
   - Zaimplementuj metodę `ShowDashboard(c echo.Context) error`, która:
     - Wywołuje `c.Render()` z szablonem `dashboard.Index` i pustym `models.DashboardViewModel{}`.
     - Obsługuje ewentualny błąd renderowania, loguje go i zwraca odpowiedź 500.

5. **Rejestracja trasy:**
   - W pliku `internal/app/routes.go`, w funkcji `setupRoutes`, dodaj nową trasę.
   - Utwórz grupę dla dashboardu, która używa middleware'u uwierzytelniającego.
   - Zarejestruj trasę `GET /dashboard` i przypisz do niej metodę `dashboardHandler.ShowDashboard`.

   ```go
   // Przykład w routes.go
   dashboardHandler := dashboard.NewHandler(...) // Inicjalizacja handlera

   // Grupa wymagająca uwierzytelnienia
   authGroup := e.Group("")
   authGroup.Use(authMiddleware.RequireAuth) // Użycie middleware

   authGroup.GET("/dashboard", dashboardHandler.ShowDashboard)
   ```
