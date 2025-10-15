# Plan Implementacji Punktu Końcowego API: GET /dashboard

## 1. Przegląd Punktu Końcowego

Celem tego punktu końcowego jest dostarczenie głównego widoku aplikacji ("dashboard"). Widok ten pełni rolę szkieletu strony, który zawiera podstawowy layout, formularze filtrów (pre-populowane z query parametrów) oraz puste kontenery. Treść tych kontenerów (lista kanałów RSS, ostatnie podsumowanie) jest dynamicznie ładowana za pomocą oddzielnych żądań htmx po załadowaniu strony.

Query parametry (`search`, `status`, `page`) służą wyłącznie do pre-populacji formularza filtrów - nie ładują bezpośrednio danych. Rzeczywiste dane feedów są pobierane przez htmx z endpointu `/feeds` z wartościami z formularza filtrów. Rozmiar strony (liczba elementów na stronę) jest stały po stronie backendu.

Endpoint wykorzystuje strukturę `feedmodels.ListFeedsQuery` do parsowania, walidacji i przechowywania query parametrów, co zapewnia spójność z endpointem `GET /feeds`.

Endpoint wymaga uwierzytelnienia.

## 2. Szczegóły Żądania

- **Metoda HTTP:** `GET`
- **Struktura URL:** `/dashboard`
- **Parametry:**
  - **Wymagane:** Brak
  - **Opcjonalne (wszystkie parsowane przez `feedmodels.ListFeedsQuery`):**
    - `search` (string): Wartość wyszukiwania do pre-populacji pola wyszukiwania w formularzu filtrów. Domyślnie: `""` (pusty string).
    - `status` (enum): Filtr statusu feedów do pre-populacji selecta w formularzu. Dopuszczalne wartości: `all`, `working`, `pending`, `error`. Domyślnie: `all`. Walidacja: `validate:"omitempty,oneof=all working error pending"`.
    - `page` (integer): Numer strony do pre-populacji w formularzu. Musi być >= 1. Domyślnie: `1`. Walidacja: `validate:"omitempty,gte=1"`.
- **Request Body:** Brak

## 3. Wykorzystywane Typy

- **`models.DashboardViewModel`**: Model widoku zawierający dane strony i wartości filtrów do pre-populacji formularza. Dane dynamiczne (lista feedów, podsumowanie) są ładowane przez htmx z innych punktów końcowych:
  - `#feed-list-container` ← `GET /feeds` (z wartościami z formularza `#feed-filters` przez `hx-include`)
  - `#summary-container` ← `GET /summaries/latest`

  ```go
  // Plik: internal/dashboard/models/dto.go
  package models

  import feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"

  type DashboardViewModel struct {
      Title     string                      // Tytuł strony (np. "Dashboard - VibeFeeder")
      UserEmail string                      // Email zalogowanego użytkownika
      Query     *feedmodels.ListFeedsQuery  // Query params dla filtrów feedów
  }
  ```

- **`feedmodels.ListFeedsQuery`**: Struktura współdzielona z endpointem `GET /feeds`, zawierająca query parametry z walidacją:

  ```go
  // Plik: internal/feed/models/queries.go
  package models

  type ListFeedsQuery struct {
      UserID string `query:"-"`                                                           // User ID (ustawiany przez handler, nie z query)
      Search string `query:"search"`                                                      // Wyszukiwanie
      Status string `query:"status" validate:"omitempty,oneof=all working error pending"` // Filtr statusu
      Page   int    `query:"page" validate:"omitempty,gte=1"`                             // Numer strony
  }

  func (q *ListFeedsQuery) SetDefaults() {
      if q.Status == "" { q.Status = "all" }
      if q.Page == 0 { q.Page = 1 }
  }
  ```

## 4. Szczegóły Odpowiedzi

- **Sukces (200 OK):**
  - **Content-Type:** `text/html; charset=utf-8`
  - **Body:** Pełna strona HTML wyrenderowana przez szablon `dashboard/view/index.templ`, zawierająca:
    - Główny layout z nawigacją (tytuł z `vm.Title`, email użytkownika z `vm.UserEmail`)
    - Formularz filtrów (`#feed-filters`) z pre-populowanymi wartościami z `vm.Query`:
      - `<input name="search" value="{vm.Query.Search}">`
      - `<select name="status">` z wybraną opcją `{vm.Query.Status}`
      - `<input type="hidden" name="page" value="1">` (reset przy zmianie filtrów)
    - Puste kontenery gotowe do wypełnienia przez htmx
- **Błędy:**
  - **400 Bad Request:**
    - **Uwaga:** W bieżącej implementacji błędy walidacji nie zwracają 400, zamiast tego handler używa wartości domyślnych.
    - W przypadku błędu `c.Bind()` lub `c.Validate()`, handler loguje ostrzeżenie i tworzy nową instancję `ListFeedsQuery` z wartościami domyślnymi.
    - Alternatywnie można zmienić implementację, aby zwracała 400 z komunikatem błędu.
  - **401 Unauthorized:**
    - **Nagłówek:** `Location: /auth/login`
    - **Body:** Puste. Odpowiedź inicjuje przekierowanie po stronie klienta na stronę logowania.
  - **500 Internal Server Error:**
    - **Content-Type:** `text/html; charset=utf-8`
    - **Body:** Strona błędu z komunikatem "Failed to load dashboard. Please refresh the page."

## 5. Przepływ Danych

1. Użytkownik wysyła żądanie `GET /dashboard?search=golang&status=working&page=2`.
2. Middleware uwierzytelniający weryfikuje sesję użytkownika.
3. Jeśli użytkownik jest uwierzytelniony, żądanie trafia do `DashboardHandler.ShowDashboard`.
4. Handler `ShowDashboard` (zgodnie z rzeczywistą implementacją):
   - Tworzy nową instancję `feedmodels.ListFeedsQuery`.
   - Wywołuje `c.Bind(query)` aby sparsować query parametry do struktury (używa tagów `query:"..."`).
   - Jeśli `c.Bind()` zwraca błąd:
     - Loguje ostrzeżenie: `"failed to bind query parameters"`
     - Tworzy nową, pustą instancję `ListFeedsQuery`
   - Wywołuje `query.SetDefaults()` aby ustawić wartości domyślne dla pustych pól.
   - Wywołuje `c.Validate(query)` aby zwalidować parametry (używa tagów `validate:"..."`).
   - Jeśli `c.Validate()` zwraca błąd:
     - Loguje ostrzeżenie: `"invalid query parameters"`
     - Tworzy nową, pustą instancję `ListFeedsQuery` i wywołuje `SetDefaults()`
   - Pobiera/tworzy dane użytkownika (obecnie mock: `"user@example.com"`).
   - Tworzy `DashboardViewModel` z `Title`, `UserEmail` i `Query`.
5. Handler wywołuje `c.Render(http.StatusOK, "", view.Index(vm))` aby wyrenderować widok.
6. Szablon `view.Index` generuje pełny kod HTML strony z:
   - Tytułem strony z `vm.Title`
   - Emailem użytkownika z `vm.UserEmail` w nawigacji
   - Formularzem filtrów z pre-populowanymi wartościami z `vm.Query`:
     - `search="golang"`, `status="working"`, `page="1"` (reset)
   - Kontenerem `#feed-list-container` z `hx-include="#feed-filters"`
7. Wygenerowany HTML jest zwracany do przeglądarki z kodem statusu 200 OK.
8. Po załadowaniu strony, htmx inicjuje żądania:
   - `GET /feeds?search=golang&status=working&page=1` (wartości z formularza przez `hx-include`)
   - `GET /summaries/latest`

## 6. Względy Bezpieczeństwa

- **Uwierzytelnianie:** Dostęp do punktu końcowego `/dashboard` musi być chroniony przez middleware, który weryfikuje, czy użytkownik jest zalogowany. W przypadku braku aktywnej sesji, middleware musi przerwać przetwarzanie żądania i zwrócić odpowiedź 401 z nagłówkiem `Location` w celu przekierowania na stronę logowania.
- **Walidacja Danych:**
  - Query parametry są walidowane przez `feedmodels.ListFeedsQuery` z tagami `validate`:
    - `status` - `validate:"omitempty,oneof=all working error pending"`
    - `page` - `validate:"omitempty,gte=1"`
    - `search` - brak walidacji (może być dowolnym stringiem)
  - **Strategia błędów:** Bieżąca implementacja używa wartości domyślnych przy błędach bindowania/walidacji, zamiast zwracać 400.
  - To zapewnia lepsze UX - użytkownik zawsze widzi dashboard, nawet jeśli podał nieprawidłowe parametry.
- **XSS Prevention:** Wartości z `vm.Query` są automatycznie escapowane przez Templ przy renderowaniu atrybutów HTML.

## 7. Obsługa Błędów

- **Nieprawidłowe parametry:** W bieżącej implementacji handler używa **Opcji A** (wartości domyślne):
  - Jeśli `c.Bind()` lub `c.Validate()` zwraca błąd, handler:
    - Loguje ostrzeżenie z detalami błędu
    - Tworzy nową instancję `ListFeedsQuery` z wartościami domyślnymi
    - Zwraca normalną odpowiedź 200 OK
  - Przykłady:
    - `/dashboard?status=invalid` → używa `status="all"`
    - `/dashboard?page=0` → używa `page=1`
- **Brak uwierzytelnienia (401):** Obsługiwane centralnie przez middleware uwierzytelniający.
- **Błąd renderowania widoku (500):** Jeśli wystąpi błąd podczas renderowania szablonu Templ, handler powinien go przechwycić, zalogować szczegóły błędu przy użyciu standardowego loggera (`slog`), a następnie zwrócić generyczną stronę błędu 500.

## 8. Etapy Wdrożenia

1. **Utworzenie struktury katalogów:**
   - Utwórz nowy katalog `internal/dashboard`.
   - Wewnątrz `internal/dashboard` utwórz podkatalogi: `handler`, `view`.

2. **Weryfikacja modelu widoku:**
   - Plik `internal/dashboard/models/dto.go` powinien zawierać:

     ```go
     package models

     import feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"

     type DashboardViewModel struct {
         Title     string
         UserEmail string
         Query     *feedmodels.ListFeedsQuery
     }
     ```

   - **Uwaga:** Model nie przechowuje już osobnych pól dla query parametrów - używa współdzielonej struktury `ListFeedsQuery`.

3. **Implementacja widoku (Templ):**
   - Utwórz plik `internal/dashboard/view/index.templ`.
   - Zdefiniuj komponent `Index(vm models.DashboardViewModel)`, który renderuje **debug view** z nazwą komponentu oraz view modelem jako string.

4. **Implementacja Handlera:**
   - Plik `internal/dashboard/handler.go` powinien zawierać:

     ```go
     package dashboard

     import (
         "log/slog"
         "net/http"

         "github.com/labstack/echo/v4"
         "github.com/tjanas94/vibefeeder/internal/dashboard/models"
         "github.com/tjanas94/vibefeeder/internal/dashboard/view"
         feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"
     )

     type Handler struct {
         logger *slog.Logger
     }

     func NewHandler(logger *slog.Logger) *Handler {
         return &Handler{logger: logger}
     }

     func (h *Handler) ShowDashboard(c echo.Context) error {
         // 1. TODO: Fetch actual user data when user service is implemented
         mockEmail := "user@example.com"

         // 2. Bind query parameters using ListFeedsQuery
         query := new(feedmodels.ListFeedsQuery)
         if err := c.Bind(query); err != nil {
             h.logger.Warn("failed to bind query parameters", "error", err)
             query = &feedmodels.ListFeedsQuery{}
         }

         // 3. Set defaults for empty values
         query.SetDefaults()

         // 4. Validate query parameters
         if err := c.Validate(query); err != nil {
             h.logger.Warn("invalid query parameters", "error", err)
             query = &feedmodels.ListFeedsQuery{}
             query.SetDefaults()
         }

         // 5. Create view model
         vm := models.DashboardViewModel{
             Title:     "Dashboard - VibeFeeder",
             UserEmail: mockEmail,
             Query:     query,
         }

         // 6. Render dashboard template
         return c.Render(http.StatusOK, "", view.Index(vm))
     }
     ```

   - **Uwaga:** Handler używa `feedmodels.ListFeedsQuery` z metodami `Bind()`, `SetDefaults()` i `Validate()`.

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

6. **Testowanie:**
   - Przetestuj różne kombinacje query parametrów:
     - `/dashboard` - domyślne wartości (status="all", page=1)
     - `/dashboard?search=golang` - pre-populacja search
     - `/dashboard?status=error` - pre-populacja status
     - `/dashboard?search=test&status=working&page=2` - wszystkie obsługiwane parametry
     - `/dashboard?status=invalid` - nieprawidłowy status (powinien użyć "all")
     - `/dashboard?page=-1` - nieprawidłowa strona (powinien użyć 1)

   - Sprawdź logi:
     - Ostrzeżenia przy błędach bindowania/walidacji
     - Brak błędów przy prawidłowych parametrach
   - Sprawdź, czy formularz filtrów jest poprawnie pre-populowany wartościami z `vm.Query`
   - Sprawdź, czy htmx poprawnie ładuje feeds z wartościami z formularza (przez `hx-include="#feed-filters"`)
