# Plan implementacji widoku: Lista Feedów

## 1. Przegląd

Widok "Lista Feedów" jest kluczowym fragmentem interfejsu użytkownika, dynamicznie ładowanym do głównego panelu. Jego celem jest umożliwienie użytkownikowi przeglądania, filtrowania i zarządzania listą swoich subskrypcji RSS. Widok ten obsługuje paginację, wyszukiwanie oraz filtrowanie po statusie, a także stanowi punkt wyjścia do operacji CRUD (dodawanie, edycja, usuwanie) na pojedynczych feedach.

## 2. Routing widoku

- **Ścieżka API:** `GET /feeds`
- **Renderowanie:** Widok jest fragmentem HTML (`partial`) renderowanym przez backend w odpowiedzi na żądanie htmx.
- **Cel w DOM:** Zawartość jest wstrzykiwana do kontenera, np. `<div id="feed-list-container">`.
- **Aktualizacja URL:** Interakcje (wyszukiwanie, filtrowanie, paginacja) aktualizują adres URL przeglądarki za pomocą `hx-push-url="true"`, aby odzwierciedlić bieżący stan widoku.

## 3. Struktura komponentów

Hierarchia komponentów w pliku `internal/feed/view/list.templ`:

```
list.templ (FeedListViewModel)
├── FeedSearchFilter (komponent do filtrowania i wyszukiwania)
│   ├── Pole wyszukiwania (`<input type="search">`)
│   ├── Grupa przycisków filtrów statusu (`.btn-group` z Alpine.js)
│   └── Przycisk "Dodaj feed" (otwiera modal)
│
├── (Renderowanie warunkowe)
│   ├── `shared/view/components/empty_state.templ` (gdy `ShowEmptyState` jest `true`)
│   ├── Komunikat "Brak wyników" (gdy `len(Feeds) == 0` po filtracji)
│   └── Tabela/Lista feedów (`<table>` lub `<div>`)
│       ├── `FeedListItem` (pętla po `Feeds`)
│       │   ├── Nazwa feedu
│       │   ├── Ikona błędu z komponentem `tooltip` (gdy `HasError` jest `true`)
│       │   ├── Przycisk "Edytuj"
│       │   └── Przycisk "Usuń"
│       └── `shared/view/components/pagination.templ` (komponent paginacji)
```

## 4. Szczegóły komponentów

### `FeedSearchFilter`

- **Opis:** Pasek narzędziowy umieszczony nad listą feedów, zawierający kontrolki do interakcji z listą.
- **Główne elementy:**
  - `input[type="search"]` do wyszukiwania po nazwie.
  - `div.btn-group` z trzema przyciskami (`button`) do filtrowania po statusie: "Wszystkie", "Działające", "Z błędami".
  - `button` "Dodaj feed", który inicjuje żądanie htmx w celu otwarcia modala z formularzem.
- **Obsługiwane interakcje:**
  - Wpisywanie tekstu w polu wyszukiwania (`hx-trigger="keyup changed delay:500ms"`).
  - Kliknięcie na przyciski filtrów (`hx-trigger="click"`).
  - Kliknięcie przycisku "Dodaj feed".
- **Warunki walidacji:** Brak po stronie klienta.
- **Typy:** Brak; stan kontrolek jest mapowany bezpośrednio na parametry zapytania `GET /feeds`.
- **Propsy:** Brak.

### `FeedListItem`

- **Opis:** Reprezentuje pojedynczy wiersz na liście feedów.
- **Główne elementy:**
  - Nazwa feedu.
  - Komponent `tooltip` (DaisyUI) z ikoną błędu, wyświetlany warunkowo.
  - Przyciski "Edytuj" i "Usuń".
- **Obsługiwane interakcje:**
  - Kliknięcie "Edytuj" (`hx-get="/feeds/{id}/edit"`) ładuje formularz edycji do modala.
  - Kliknięcie "Usuń" (`hx-delete="/feeds/{id}"`) inicjuje proces usuwania (z potwierdzeniem).
- **Warunki walidacji:** Brak.
- **Typy:** `FeedItemViewModel`.
- **Propsy:** `Feed (FeedItemViewModel)`.

### `Pagination`

- **Opis:** Komponent do nawigacji między stronami listy.
- **Główne elementy:** Przyciski "Poprzednia", "Następna" oraz numery stron.
- **Obsługiwane interakcje:** Kliknięcie na link/przycisk strony (`hx-get` do odpowiedniej strony).
- **Warunki walidacji:** Brak (linki są generowane na serwerze).
- **Typy:** `sharedmodels.PaginationViewModel`.
- **Propsy:** `Pagination (sharedmodels.PaginationViewModel)`.

## 5. Typy

- **`FeedListViewModel`**: Główny model widoku przekazywany do szablonu `list.templ`.
  - `Feeds []FeedItemViewModel`: Lista feedów do wyświetlenia na bieżącej stronie.
  - `ShowEmptyState bool`: Flaga określająca, czy wyświetlić stan początkowy dla nowego użytkownika.
  - `Pagination sharedmodels.PaginationViewModel`: Dane potrzebne do renderowania komponentu paginacji.
- **`FeedItemViewModel`**: Model dla pojedynczego feedu na liście.
  - `ID string`: Identyfikator feedu.
  - `Name string`: Nazwa feedu.
  - `URL string`: Adres URL feedu.
  - `HasError bool`: `true`, jeśli ostatnie pobranie zakończyło się błędem.
  - `ErrorMessage string`: Treść błędu do wyświetlenia w tooltipie.
- **`PaginationViewModel`**: Model dla danych paginacji.
  - `CurrentPage, TotalPages, TotalItems int`: Informacje o stronicowaniu.
  - `HasPrevious, HasNext bool`: Flagi do włączania/wyłączania przycisków nawigacyjnych.

## 6. Zarządzanie stanem

- **Stan serwera (źródło prawdy):** Cały stan dotyczący danych (lista feedów, filtry, paginacja) jest zarządzany po stronie serwera. htmx służy do pobierania i renderowania tego stanu.
- **Stan klienta (Alpine.js):**
  - **Aktywny filtr:** Stan aktywnego przycisku filtra (`.btn-active`) jest zarządzany przez Alpine.js. Komponent `x-data` odczytuje początkowy stan z parametru `status` w URL, aby zapewnić spójność po odświeżeniu strony.
    ```html
    <div
      x-data="{ activeFilter: new URL(window.location.href).searchParams.get('status') || 'all' }"
    >
      <button :class="{ 'btn-active': activeFilter === 'all' }" ...>Wszystkie</button>
      ...
    </div>
    ```
  - **URL:** Stan filtrów i paginacji jest synchronizowany z adresem URL przeglądarki za pomocą atrybutu `hx-push-url="true"` w elementach wyzwalających żądania htmx.

## 7. Integracja API

- **Endpoint:** `GET /feeds`
- **Żądanie:**
  - **Metoda:** `GET`
  - **Parametry:** `search` (string), `status` (enum), `page` (int), `limit` (int).
  - **Wyzwalacze (Triggers):**
    - Wyszukiwanie: `hx-trigger="keyup changed delay:500ms"` na polu `<input>`.
    - Filtrowanie: `hx-trigger="click"` na przyciskach statusu.
    - Paginacja: `hx-trigger="click"` na linkach paginacji.
  - **Atrybuty htmx:**
    - `hx-get="/feeds"`
    - `hx-target="#feed-list-container"` (lub wewnętrzny kontener listy)
    - `hx-push-url="/dashboard"`
    - `hx-indicator=".htmx-indicator"`
    - `hx-vals` do przekazywania wartości (np. statusu filtra).
- **Odpowiedź:** Serwer zwraca kod `200 OK` z fragmentem HTML, który zastępuje zawartość elementu docelowego.

## 8. Interakcje użytkownika

- **Wyszukiwanie:** Użytkownik wpisuje frazę. Po 500ms bezczynności, wysyłane jest żądanie `GET /feeds?search=...`, a lista jest aktualizowana.
- **Filtrowanie:** Użytkownik klika przycisk statusu. Wysyłane jest żądanie `GET /feeds?status=...`, lista jest aktualizowana, a kliknięty przycisk otrzymuje klasę `.btn-active`.
- **Paginacja:** Użytkownik klika numer strony. Wysyłane jest żądanie `GET /feeds?page=...` (z zachowaniem pozostałych filtrów), a lista jest aktualizowana.
- **Wyświetlanie błędu:** Użytkownik najeżdża na ikonę błędu, co powoduje wyświetlenie tooltipa z komunikatem o błędzie.

## 9. Warunki i walidacja

Interfejs użytkownika jest zaprojektowany tak, aby uniemożliwić wysyłanie niepoprawnych wartości do API:

- **`status`:** Kontrolowany przez przyciski, które mają predefiniowane, poprawne wartości.
- **`page`:** Linki do stron są generowane na serwerze na podstawie dostępnego zakresu.
- **`search`:** Jest dowolnym ciągiem znaków; walidacja i sanitization odbywają się po stronie serwera.

## 10. Obsługa błędów

- **Błędy serwera (4xx, 5xx):** Serwer powinien zwrócić fragment HTML z komunikatem o błędzie (np. "Nie udało się załadować feedów"). htmx automatycznie umieści ten fragment w kontenerze docelowym. Warto dodać przycisk "Spróbuj ponownie", który ponowi ostatnie żądanie.
- **Błędy sieciowe:** Globalna obsługa zdarzeń htmx (np. `htmx:sendError`) może być użyta do wyświetlania generycznego komunikatu (np. toast) o problemach z połączeniem.
- **Brak wyników:** Jeśli API zwraca pustą listę `Feeds` (ale nie `ShowEmptyState`), komponent powinien wyświetlić komunikat "Nie znaleziono feedów pasujących do Twoich kryteriów".

## 11. Kroki implementacji

1.  **Stworzenie szablonu `list.templ`:** Utworzyć plik `internal/feed/view/list.templ`, który przyjmuje `FeedListViewModel` jako prop.
2.  **Implementacja struktury:** Zaimplementować główną strukturę HTML, w tym kontener docelowy dla htmx (`<div id="feed-list">`).
3.  **Implementacja `FeedSearchFilter`:**
    - Dodać pole wyszukiwania z atrybutami `hx-get`, `hx-trigger`, `hx-target`, `hx-push-url`.
    - Dodać grupę przycisków filtrów z logiką Alpine.js do zarządzania klasą `.btn-active` i atrybutami htmx do wysyłania żądań.
4.  **Implementacja logiki warunkowej:** Dodać bloki `if/else` w szablonie `templ` do obsługi `ShowEmptyState`, braku wyników i wyświetlania listy.
5.  **Implementacja `FeedListItem`:** W pętli `for` po `Feeds` stworzyć komponent dla pojedynczego elementu, w tym warunkowe renderowanie ikony błędu i tooltipa (DaisyUI). Dodać przyciski "Edytuj" i "Usuń" z odpowiednimi atrybutami `hx-*`.
6.  **Integracja `Pagination`:** Dodać komponent paginacji i przekazać do niego `ViewModel.Pagination`. Upewnić się, że linki generowane przez komponent zachowują aktywne filtry.
7.  **Styling:** Użyć klas Tailwind CSS i komponentów DaisyUI do ostylowania widoku zgodnie z projektem.
8.  **Testowanie:** Przetestować wszystkie interakcje: wyszukiwanie (w tym debounce), filtrowanie, paginację, działanie linków, obsługę stanów pustych i błędów.
