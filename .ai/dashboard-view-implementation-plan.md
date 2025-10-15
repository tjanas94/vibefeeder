# Plan implementacji widoku: Panel Główny (Dashboard)

## 1. Przegląd

Panel Główny (Dashboard) jest głównym widokiem dla zalogowanego użytkownika, pełniącym rolę "powłoki aplikacji" (app shell). Jego celem jest zapewnienie stałej nawigacji, formularzy filtrów oraz kontenera na dynamicznie ładowaną treść, taką jak lista feedów. Widok ten zawiera również strukturę dla okien modalnych używanych do interakcji z użytkownikiem (dodawanie/edycja feedów, wyświetlanie podsumowań).

Dashboard obsługuje query parametry (`search`, `status`, `page`), które służą do pre-populacji formularzy filtrów. Te same parametry są następnie przekazywane do endpointu `/feeds` podczas dynamicznego ładowania listy feedów.

## 2. Routing widoku

- **Ścieżka**: `/dashboard`
- **Query Parametry** (opcjonalne):
  - `search` (string): Wstępna wartość wyszukiwania do pre-populacji pola search
  - `status` (enum): Wstępny filtr statusu. Wartości: `all`, `working`, `pending`, `error`. Domyślnie: `all`
  - `page` (integer): Wstępny numer strony. Domyślnie: `1`
- **Ochrona**: Dostęp do tej ścieżki musi być chroniony przez middleware uwierzytelniający. Niezalogowani użytkownicy powinni być przekierowywani na stronę logowania (`/login`).

## 3. Struktura komponentów

Hierarchia komponentów będzie oparta na kompozycji, gdzie główny layout aplikacji (`SharedLayout`) otacza specyficzną treść strony (`DashboardPage`).

```
SharedLayout (internal/shared/view/layout.templ)
│
└── Slot na treść strony (renderuje komponent potomny)
    │
    └── DashboardPage (internal/dashboard/view/index.templ)
        │
        ├── Navbar (components.Navbar)
        │   ├── Logo aplikacji ("VibeFeeder")
        │   ├── Slot na niestandardowe przyciski (np. "Summary")
        │   ├── Adres e-mail użytkownika
        │   └── Przycisk "Logout"
        │
        ├── Główny kontener treści (`<main>`)
        │   ├── Formularz filtrów (`<form id="feed-filter-form">`)
        │   │   ├── Pole wyszukiwania (input name="search")
        │   │   ├── Wybór statusu (radio buttons name="status")
        │   │   └── Przycisk "Add Feed"
        │   └── Kontener na listę feedów (`<div id="feed-list">`)
        │
        └── Kontenery na modale (`<dialog>`)
            ├── Modal podsumowania (`#summary-modal`)
            ├── Modal formularza feedu (`#feed-form-modal`)
            └── Modal potwierdzenia usunięcia (`#delete-confirmation-modal`)
```

## 4. Szczegóły komponentów

### `SharedLayout`

- **Opis komponentu**: Współdzielony layout dla wszystkich widoków aplikacji. Zawiera podstawową strukturę HTML (`<html>`, `<head>`, `<body>`), globalne kontenery na błędy i powiadomienia (toast), oraz renderuje dynamiczną treść strony w dedykowanym slocie. Layout nie zawiera Navbara - każda strona jest odpowiedzialna za renderowanie własnej nawigacji.
- **Główne elementy**: `<html>`, `<head>`, `<body>`, kontenery globalne (`#global-errors`, `#toast-container`), `templ.Component` jako slot na treść.
- **Obsługiwane interakcje**: Brak - layout jest tylko strukturą.
- **Typy**: `view.LayoutProps`
- **Propsy**: Przyjmuje `LayoutProps` zawierający tylko `Title` (tytuł strony dla `<title>`) oraz komponent potomny do wyrenderowania w body.

### `Navbar` (komponent wielokrotnego użytku)

- **Opis komponentu**: Górny pasek nawigacyjny (DaisyUI `navbar`) z slotem na niestandardowe przyciski. Jest renderowany przez poszczególne strony (np. w `DashboardPage`), nie jest częścią `SharedLayout`.
- **Główne elementy**: `<nav>`, logo aplikacji (link do `/dashboard`), slot na niestandardowe przyciski (`{ children... }`), e-mail użytkownika, przycisk wylogowania.
- **Obsługiwane interakcje**:
  - **Kliknięcie logo**: Przekierowanie do `/dashboard`.
  - **Kliknięcie "Logout"**: Wysyła żądanie `POST /auth/logout` w celu wylogowania użytkownika.
  - **Niestandardowe przyciski**: Przekazywane przez slot children (np. przycisk "Summary" w dashboardzie).
- **Typy**: `components.NavbarProps`
- **Propsy**: Przyjmuje `NavbarProps` z polem `UserEmail` oraz opcjonalne dzieci (custom buttons).

### `DashboardPage` (`index.templ`)

- **Opis komponentu**: Główna treść dla ścieżki `/dashboard`. Zawiera navbar z przyciskiem podsumowania, formularz filtrów (w tym przycisk dodawania feedu), kontener na listę feedów oraz definicje wszystkich okien modalnych.
- **Główne elementy**:
  - `<nav>` - nawigacja (komponent `Navbar`)
  - `<main>` - główny kontener
  - `<form id="feed-filter-form">` - formularz filtrów z polami search, status (radio buttons) i przyciskiem "Add Feed"
  - `<div id="feed-list">` - kontener na listę feedów
  - `<dialog>` - modale dla różnych interakcji
- **Obsługiwane interakcje**:
  - **Załadowanie strony**: Element `#feed-list` automatycznie wyzwala żądanie `GET /feeds` za pomocą `hx-trigger="load, refreshFeedList from:document"` i `hx-include="#feed-filter-form"`, przekazując wartości z formularza filtrów.
  - **Zmiana filtrów**: Formularz `#feed-filter-form` wyzwala `GET /feeds` z `hx-trigger="change from:input[type=radio], keyup changed delay:500ms from:input[type=search]"` i `hx-target="#feed-list"`, aktualizując listę feedów bez przeładowania strony.
  - **Kliknięcie przycisku "Summary"**: Otwiera modal podsumowania i ładuje treść z `GET /summaries/latest`.
  - **Kliknięcie przycisku "Add Feed"**: Otwiera modal formularza feedu i ładuje treść z `GET /feeds/new` do `#feed-form-modal-content`.
- **Typy**: `models.DashboardViewModel` (zawiera `Title`, `UserEmail` i `Query` z typu `feedmodels.ListFeedsQuery`).
- **Propsy**: Przyjmuje `DashboardViewModel` z pre-populowaną strukturą `Query` zawierającą wartości filtrów z query parametrów.

### `Modals` (`<dialog>`)

- **Opis komponentu**: Trzy oddzielne modale DaisyUI (`<dialog>`) do wyświetlania podsumowania, formularza feedu i potwierdzenia usunięcia. Ich widoczność jest kontrolowana przez stan Alpine.js.
- **Główne elementy**: `<dialog>`, `<div class="modal-box">`, przycisk zamykania, oraz wewnętrzny `<div>` z `id` służący jako cel dla htmx (np. `#summary-modal-content`).
- **Obsługiwane interakcje**: Otwieranie/zamykanie za pomocą Alpine.js. Treść ładowana dynamicznie przez htmx.

## 5. Typy

- **`models.DashboardViewModel`**: Struktura zawierająca dane dla widoku dashboardu:

  ```go
  // Plik: internal/dashboard/models/dto.go
  package models

  import feedmodels "github.com/tjanas94/vibefeeder/internal/feed/models"

  type DashboardViewModel struct {
      Title     string                      // Tytuł strony
      UserEmail string                      // Email zalogowanego użytkownika
      Query     *feedmodels.ListFeedsQuery  // Query params dla filtrów feedów (search, status, page)
  }
  ```

  **`feedmodels.ListFeedsQuery`** zawiera pola:
  - `Search string` - wartość wyszukiwania (domyślnie: "")
  - `Status string` - filtr statusu (domyślnie: "all")
  - `Page int` - numer strony (domyślnie: 1)
  - Metoda `SetDefaults()` ustawia wartości domyślne
  - Tagi walidacji dla wszystkich pól

- **`view.LayoutProps`**: Struktura przekazywana do współdzielonego layoutu:

  ```go
  // Plik: internal/shared/view/types.go
  package view

  // LayoutProps contains the data needed to render the base layout
  type LayoutProps struct {
      Title string // Tytuł strony dla tagu <title>
  }
  ```

- **`components.NavbarProps`**: Struktura przekazywana do komponentu Navbar:

  ```go
  // Plik: internal/shared/view/components/types.go
  package components

  type NavbarProps struct {
      // UserEmail is the email address of the authenticated user
      // Displayed in the navbar for user identification
      UserEmail string
  }
  ```

## 6. Zarządzanie stanem

Zarządzanie stanem po stronie klienta będzie realizowane za pomocą Alpine.js i ograniczy się do kontrolowania widoczności okien modalnych oraz przechowywania informacji o obecności feedów.

- **Komponent Alpine.js**: Główny kontener dashboardu będzie zainicjowany za pomocą `x-data` z następującymi zmiennymi stanu:
  - `openModal: null` - przechowuje nazwę aktualnie otwartego modalu ('summary', 'feed', 'delete', lub null)
  - `hasFeeds: false` - flaga informująca czy użytkownik ma jakiekolwiek feedy (używana do warunkowego wyświetlania przycisku "Summary")
  - `lastFocusedElement: null` - przechowuje referencję do ostatnio skupionego elementu przed otwarciem modalu (dla accessibility)
- **Interakcja**:
  - Przyciski otwierają modale przez wysyłanie eventu `@open-modal.window` z nazwą modalu w `$event.detail.modal`
  - Modale zamykają się przez event `@close-modal.window`
  - Status feedów jest aktualizowany przez event `@feeds-loaded.window` z `$event.detail.hasFeeds`
  - Atrybut `x-show` kontroluje widoczność elementów na podstawie stanu

## 7. Integracja API

- **Początkowe ładowanie danych**: Po załadowaniu strony, element `<div id="feed-list" hx-get="/feeds" hx-trigger="load, refreshFeedList from:document" hx-include="#feed-filter-form">` wykona żądanie `GET` do `/feeds` z wartościami z formularza filtrów. Odpowiedź serwera (HTML) zostanie wstawiona do tego kontenera.
- **Ładowanie treści modali**: Przyciski otwierające modale będą również posiadały atrybuty htmx, np. `hx-get="/feeds/new"` z `hx-target` wskazującym na kontener wewnątrz modalu (np. `#feed-form-modal-content`).
- **Filtrowanie i paginacja**: Formularz `#feed-filter-form` oraz linki paginacji używają `hx-get="/feeds"` do aktualizacji listy feedów z odpowiednimi parametrami query.
- **Odświeżanie listy**: Event `refreshFeedList` wysyłany z `document` powoduje automatyczne przeładowanie listy feedów (np. po dodaniu/usunięciu feeda).

## 8. Interakcje użytkownika

- **Użytkownik wchodzi na `/dashboard`**:
  - Widzi layout z nawigacją, formularz filtrów (z wartościami z URL query params) i wskaźnik ładowania w miejscu listy feedów.
  - Po chwili lista (lub stan pusty) pojawia się automatycznie, załadowana z `/feeds` z aktualnymi wartościami filtrów.

- **Użytkownik wprowadza tekst w pole wyszukiwania**:
  - Po 500ms od ostatniego naciśnięcia klawisza (debounce) lista feedów jest automatycznie odświeżana.
  - URL nie zmienia się (tylko stan formularza i treść #feed-list).

- **Użytkownik zmienia filtr statusu**:
  - Lista feedów jest natychmiast odświeżana z nowym filtrem.
  - Paginacja jest resetowana do strony 1.

- **Użytkownik klika link paginacji**:
  - Lista feedów ładuje się dla wybranej strony.
  - Aktualne wartości search i status są zachowane w URL linku.

- **Użytkownik klika "Summary"**: Otwiera się modal, a w nim pojawia się treść ostatniego podsumowania, pobrana z `GET /summaries/latest`.

- **Użytkownik klika "Add Feed"**: Otwiera się modal formularza feedu (`#feed-form-modal`) i ładuje się formularz dodawania nowego feedu z `GET /feeds/new` do `#feed-form-modal-content`.

- **Użytkownik dodaje/usuwa feed**:
  - Po pomyślnej operacji lista feedów jest automatycznie odświeżana z zachowaniem aktualnych filtrów.
  - Modal zostaje zamknięty.

- **Użytkownik klika "Logout"**: Zostaje wylogowany i przekierowany na stronę logowania.

## 9. Warunki i walidacja

Jedynym warunkiem dla tego widoku jest uwierzytelnienie użytkownika. Jest to w całości weryfikowane przez middleware po stronie serwera. Sam komponent widoku nie implementuje żadnej logiki walidacyjnej.

## 10. Obsługa błędów

- **Błąd 401 Unauthorized**: Użytkownik jest automatycznie przekierowywany na stronę logowania przez middleware.
- **Błąd 500 Internal Server Error**: Serwer zwraca dedykowaną stronę błędu.
- **Błąd ładowania htmx (np. `GET /feeds`)**: Serwer powinien zwrócić fragment HTML z komunikatem o błędzie (np. komponent `ErrorFragment`), który htmx umieści w kontenerze `#feed-list`. Dodatkowo, można użyć globalnych powiadomień (toast) wyzwalanych przez zdarzenie `htmx:responseError`.

## 11. Kroki implementacji

1.  **Weryfikacja `SharedLayout`**: Upewnij się, że `internal/shared/view/layout.templ` zawiera podstawową strukturę HTML bez Navbara. Layout powinien przyjmować `view.LayoutProps` (tylko `Title`) oraz `templ.Component` jako dziecko (slot na treść strony).

2.  **Weryfikacja typów**: - Sprawdź, czy `internal/dashboard/models/dto.go` zawiera `DashboardViewModel` z polami: `Title`, `UserEmail` i `Query`. - Sprawdź, czy `internal/shared/view/types.go` zawiera `LayoutProps` z polem `Title`. - Sprawdź, czy `internal/shared/view/components/types.go` zawiera `NavbarProps` z polem `UserEmail`.
    </parameter>

3.  **Weryfikacja komponentu `Navbar`**: Upewnij się, że `internal/shared/view/components/navbar.templ` zawiera komponent z slotem children dla niestandardowych przycisków oraz przyjmuje `NavbarProps` z polem `UserEmail`.

4.  **Aktualizacja `DashboardHandler`**:
    - W `internal/dashboard/handler.go` zmodyfikuj `ShowDashboard`, aby:
      - Tworzył instancję `feedmodels.ListFeedsQuery` i używał `c.Bind(query)` do parsowania query parametrów.
      - Wywoływał `query.SetDefaults()` aby ustawić wartości domyślne.
      - Walidował query używając `c.Validate(query)` (walidacja oparta na tagach w `ListFeedsQuery`).
      - W przypadku błędu bindowania lub walidacji, używał nowej instancji z wartościami domyślnymi.
      - Pobierał dane użytkownika (e-mail) z kontekstu.
      - Tworzył `DashboardViewModel` z `Title`, `UserEmail` i `Query`.
      - Przekazywał view model do renderowania widoku.

5.  **Implementacja formularza filtrów**: W pliku `internal/feed/view/list.templ`:
    - Stwórz komponent `FeedSearchFilter` z formularzem `<form id="feed-filter-form">` z atrybutami htmx:
      - `hx-get="/feeds"`
      - `hx-target="#feed-list"`
      - `hx-trigger="change from:input[type=radio], keyup changed delay:500ms from:input[type=search], search from:input[type=search]"`
    - Dodaj pole input `name="search"` typu `search` z wartością z props.
    - Dodaj radio buttons `name="status"` z opcjami (all, working, pending, error) i wybraną wartością z props.
    - Dodaj przycisk "Add Feed" (`type="button"`) z atrybutami:
      - `@click="lastFocusedElement = $event.target"` (Alpine.js - zapisuje element dla accessibility)
      - `hx-get="/feeds/new"` (ładuje formularz nowego feedu)
      - `hx-target="#feed-form-modal-content"` (wstawia treść do modalu)
      - `hx-trigger="click"` (trigger na kliknięcie)

6.  **Implementacja kontenera listy feedów**: W pliku `internal/dashboard/view/index.templ`:
    - Dodaj `<div id="feed-list">` z atrybutami:
      - `hx-get="/feeds"`
      - `hx-trigger="load, refreshFeedList from:document"`
      - `hx-include="#feed-filter-form"`
    - Umieść w nim wskaźnik ładowania (komponent `LoadingSpinner`).

7.  **Dodanie Modali**: Dodaj struktury HTML dla trzech okien modalnych używając komponentu `components.Modal` z odpowiednimi `ID`, `ContentID`, `AlpineStateVar` i `MaxWidth`.

8.  **Inicjalizacja Alpine.js**: Dodaj atrybut `x-data` do głównego kontenera div w `DashboardPage` (nie w `SharedLayout`), inicjalizując zmienne stanu: `openModal`, `hasFeeds`, `lastFocusedElement` oraz event handlery dla `@close-modal.window`, `@open-modal.window`, `@feeds-loaded.window`.

9.  **Powiązanie Interakcji**:
    - Przy renderowaniu `Navbar` w dashboardzie, przekaż niestandardowe przyciski (np. "Summary") przez slot children z odpowiednimi atrybutami `@click` i `hx-*`.
    - Przycisk "Add Feed" jest już częścią komponentu `FeedSearchFilter` - nie wymaga dodatkowej implementacji w dashboardzie.

10. **Weryfikacja wylogowania**: Upewnij się, że przycisk "Logout" w komponencie `Navbar` poprawnie inicjuje żądanie `POST` do `/auth/logout` za pomocą `hx-post`.

11. **Stylowanie**: Użyj klas Tailwind i DaisyUI, aby zapewnić spójny wygląd formularza filtrów, wskaźników ładowania i modali.

12. **Testowanie**: Sprawdź wszystkie ścieżki interakcji:
    - Pre-populacja filtrów z query params przy wejściu na `/dashboard?search=test&status=error&page=2`.
    - Automatyczne ładowanie listy z wartościami filtrów (używając `hx-include="#feed-filter-form"`).
    - Działanie wyszukiwania z debounce 500ms i filtrowania w czasie rzeczywistym.
    - Zachowanie filtrów przy paginacji (wszystkie parametry: search, status).
    - Zachowanie filtrów po dodaniu/usunięciu feedu (używając `hx-include="#feed-filter-form"` i eventu `refreshFeedList`).
    - Walidacja nieprawidłowych query params (np. `status=invalid`, `page=0`, `limit=1000`).
    - Otwieranie i ładowanie treści modali.
    - Obsługę błędów oraz proces wylogowania.
