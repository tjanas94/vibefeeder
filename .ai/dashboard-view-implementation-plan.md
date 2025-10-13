# Plan implementacji widoku: Panel Główny (Dashboard)

## 1. Przegląd

Panel Główny (Dashboard) jest głównym widokiem dla zalogowanego użytkownika, pełniącym rolę "powłoki aplikacji" (app shell). Jego celem jest zapewnienie stałej nawigacji oraz kontenera na dynamicznie ładowaną treść, taką jak lista feedów. Widok ten zawiera również strukturę dla okien modalnych używanych do interakcji z użytkownikiem (dodawanie/edycja feedów, wyświetlanie podsumowań).

## 2. Routing widoku

- **Ścieżka**: `/dashboard`
- **Ochrona**: Dostęp do tej ścieżki musi być chroniony przez middleware uwierzytelniający. Niezalogowani użytkownicy powinni być przekierowywani na stronę logowania (`/login`).

## 3. Struktura komponentów

Hierarchia komponentów będzie oparta na kompozycji, gdzie główny layout aplikacji (`SharedLayout`) otacza specyficzną treść strony (`DashboardPage`).

```
SharedLayout (internal/shared/view/layout.templ)
│
├── Navbar (zintegrowany w layoucie)
│   ├── Logo aplikacji ("VibeFeeder")
│   ├── Przycisk "Podsumowanie"
│   ├── Adres e-mail użytkownika
│   └── Przycisk "Wyloguj"
│
└── Slot na treść strony (renderuje komponent potomny)
    │
    └── DashboardPage (internal/dashboard/view/index.templ)
        │
        ├── Główny kontener treści (`<main>`)
        │   └── Kontener na listę feedów (`<div id="feed-list-container">`)
        │
        └── Kontenery na modale (`<dialog>`)
            ├── Modal podsumowania (`#summary-modal`)
            ├── Modal formularza feedu (`#feed-form-modal`)
            └── Modal potwierdzenia usunięcia (`#delete-confirmation-modal`)
```

## 4. Szczegóły komponentów

### `SharedLayout`

- **Opis komponentu**: Współdzielony layout dla wszystkich uwierzytelnionych widoków. Zawiera `<html>`, `<head>`, `<body>` oraz stały komponent `Navbar`. Renderuje dynamiczną treść strony w dedykowanym slocie.
- **Główne elementy**: `<html>`, `<body>`, `Navbar` (DaisyUI), `templ.Component` jako slot na treść.
- **Obsługiwane interakcje**: Nawigacja globalna.
- **Typy**: `view.LayoutData`
- **Propsy**: Przyjmuje `LayoutData` (do wyświetlenia tytułu strony i danych użytkownika w navbarze) oraz komponent potomny do wyrenderowania.

### `Navbar` (część `SharedLayout`)

- **Opis komponentu**: Górny, stały pasek nawigacyjny (DaisyUI `navbar`).
- **Główne elementy**: `<a>` dla logo, `<button>` dla akcji, `<div>` na e-mail użytkownika.
- **Obsługiwane interakcje**:
  - **Kliknięcie "Podsumowanie"**: Otwiera modal podsumowania (`#summary-modal`) i dynamicznie wczytuje jego treść z `GET /summaries/latest`.
  - **Kliknięcie "Wyloguj"**: Wysyła żądanie `POST /auth/logout` w celu wylogowania użytkownika.
- **Typy**: Dane pochodzą z `view.LayoutData`.

### `DashboardPage` (`index.templ`)

- **Opis komponentu**: Główna treść dla ścieżki `/dashboard`. Zawiera kontener na listę feedów oraz definicje wszystkich okien modalnych.
- **Główne elementy**: `<main>`, `<div id="feed-list-container">`, `<dialog id="summary-modal">`, `<dialog id="feed-form-modal">`, `<dialog id="delete-confirmation-modal">`.
- **Obsługiwane interakcje**:
  - **Załadowanie strony**: Element `#feed-list-container` automatycznie wyzwala żądanie `GET /feeds` za pomocą `hx-trigger="load"`.
- **Typy**: `models.DashboardViewModel` (pusty).
- **Propsy**: Brak.

### `Modals` (`<dialog>`)

- **Opis komponentu**: Trzy oddzielne modale DaisyUI (`<dialog>`) do wyświetlania podsumowania, formularza feedu i potwierdzenia usunięcia. Ich widoczność jest kontrolowana przez stan Alpine.js.
- **Główne elementy**: `<dialog>`, `<div class="modal-box">`, przycisk zamykania, oraz wewnętrzny `<div>` z `id` służący jako cel dla htmx (np. `#summary-modal-content`).
- **Obsługiwane interakcje**: Otwieranie/zamykanie za pomocą Alpine.js. Treść ładowana dynamicznie przez htmx.

## 5. Typy

- **`models.DashboardViewModel`**: Pusta struktura, ponieważ widok nie otrzymuje żadnych danych bezpośrednio od handlera.
- **`view.LayoutData` (nowy typ)**: Należy zdefiniować nową strukturę w `internal/shared/view/`, aby przekazywać dane do współdzielonego layoutu.

  ```go
  // Plik: internal/shared/view/layout.templ.go
  package view

  // LayoutData przechowuje dane wymagane przez współdzielony layout.
  type LayoutData struct {
      UserEmail       string // Email zalogowanego użytkownika.
      IsAuthenticated bool   // Flaga do warunkowego renderowania elementów.
      Title           string // Tytuł strony.
  }
  ```

## 6. Zarządzanie stanem

Zarządzanie stanem po stronie klienta będzie realizowane za pomocą Alpine.js i ograniczy się do kontrolowania widoczności okien modalnych.

- **Komponent Alpine.js**: Główny element strony (np. `<body>`) będzie zainicjowany za pomocą `x-data`.
- **Zmienne stanu**:
  - `isSummaryModalOpen: false`
  - `isFeedModalOpen: false`
  - `isDeleteModalOpen: false`
- **Interakcja**: Przyciski będą przełączać te zmienne (`@click="isSummaryModalOpen = true"`), a atrybut `:open` w modalach (`<dialog :open="isSummaryModalOpen">`) będzie na nie reagował. Zamykanie modali będzie obsługiwane przez nasłuchiwanie na zdarzenia `HX-Trigger` wysyłane z serwera po pomyślnej akcji (`x-on:close-modals.window="..."`).

## 7. Integracja API

- **Początkowe ładowanie danych**: Po załadowaniu strony, element `<div id="feed-list-container" hx-get="/feeds" hx-trigger="load">` wykona żądanie `GET` do `/feeds`. Odpowiedź serwera (HTML) zostanie wstawiona do tego kontenera.
- **Ładowanie treści modali**: Przyciski otwierające modale będą również posiadały atrybuty htmx, np. `hx-get="/feeds/new"` z `hx-target` wskazującym na kontener wewnątrz modalu.

## 8. Interakcje użytkownika

- **Użytkownik wchodzi na `/dashboard`**: Widzi layout z nawigacją i wskaźnikiem ładowania w miejscu listy feedów. Po chwili lista (lub stan pusty) pojawia się automatycznie.
- **Użytkownik klika "Podsumowanie"**: Otwiera się modal, a w nim pojawia się treść ostatniego podsumowania, pobrana z `GET /summaries/latest`.
- **Użytkownik klika "Dodaj feed"** (w stanie pustym): Otwiera się modal z formularzem dodawania nowego feedu, pobranym z `GET /feeds/new`.
- **Użytkownik klika "Wyloguj"**: Zostaje wylogowany i przekierowany na stronę logowania.

## 9. Warunki i walidacja

Jedynym warunkiem dla tego widoku jest uwierzytelnienie użytkownika. Jest to w całości weryfikowane przez middleware po stronie serwera. Sam komponent widoku nie implementuje żadnej logiki walidacyjnej.

## 10. Obsługa błędów

- **Błąd 401 Unauthorized**: Użytkownik jest automatycznie przekierowywany na stronę logowania przez middleware.
- **Błąd 500 Internal Server Error**: Serwer zwraca dedykowaną stronę błędu.
- **Błąd ładowania htmx (np. `GET /feeds`)**: Serwer powinien zwrócić fragment HTML z komunikatem o błędzie (np. komponent `ErrorFragment`), który htmx umieści w kontenerze `#feed-list-container`. Dodatkowo, można użyć globalnych powiadomień (toast) wyzwalanych przez zdarzenie `htmx:responseError`.

## 11. Kroki implementacji

1.  **Utworzenie/Modyfikacja `SharedLayout`**: W `internal/shared/view/layout.templ` zdefiniuj główny layout aplikacji z komponentem `Navbar`. Upewnij się, że layout przyjmuje `view.LayoutData` oraz `templ.Component` jako dziecko.
2.  **Aktualizacja `DashboardHandler`**: W `internal/dashboard/handler.go` zmodyfikuj `ShowDashboard`, aby pobierał dane użytkownika (e-mail) z `echo.Context` i przekazywał je do `SharedLayout` opakowującego widok `DashboardPage`.
3.  **Implementacja `DashboardPage`**: W pliku `internal/dashboard/view/index.templ` stwórz komponent `Index`.
4.  **Struktura `DashboardPage`**: Wewnątrz `Index` użyj `SharedLayout`. Zdefiniuj element `<main>` oraz kontener `<div id="feed-list-container" hx-get="/feeds" hx-trigger="load">`. Umieść w nim wskaźnik ładowania (np. DaisyUI `spinner`).
5.  **Dodanie Modali**: W tym samym pliku dodaj struktury HTML dla trzech okien modalnych (`<dialog>`) z odpowiednimi `id` i atrybutami `x-show` lub `:open` powiązanymi ze stanem Alpine.js.
6.  **Inicjalizacja Alpine.js**: Dodaj atrybut `x-data` do elementu `<body>` w `SharedLayout`, inicjalizując zmienne stanu dla modali.
7.  **Powiązanie Interakcji**: W `Navbar` oraz w treściach ładowanych dynamicznie (np. przycisk "Dodaj feed") dodaj atrybuty `@click` do przełączania stanu Alpine.js oraz atrybuty `hx-*` do ładowania treści modali.
8.  **Implementacja Wylogowania**: Upewnij się, że przycisk "Wyloguj" poprawnie inicjuje żądanie `POST` do `/auth/logout` (np. za pomocą `hx-post`).
9.  **Stylowanie**: Użyj klas Tailwind i DaisyUI, aby zapewnić spójny wygląd, w tym dla wskaźników ładowania `htmx-indicator`.
10. **Testowanie**: Sprawdź wszystkie ścieżki interakcji: automatyczne ładowanie listy, otwieranie i ładowanie treści modali, obsługę błędów oraz proces wylogowania.
