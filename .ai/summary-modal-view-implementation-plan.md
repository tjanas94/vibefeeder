# Plan implementacji widoku: Modal Podsumowania

## 1. Przegląd

Celem jest implementacja widoku podsumowania, który będzie wyświetlany w modalu. Widok ten pozwoli użytkownikowi na wyświetlenie ostatnio wygenerowanego podsumowania AI oraz na zainicjowanie procesu generowania nowego. Widok będzie obsługiwał różne stany, takie jak stan początkowy (brak podsumowania), stan ładowania podczas generowania, stan błędu oraz stan wyświetlania gotowego podsumowania.

## 2. Routing widoku

Widok nie jest samodzielną stroną, lecz fragmentem interfejsu ładowanym dynamicznie do modala. Interakcje opierają się na dwóch endpointach API:

- `GET /summaries/latest`: Pobiera początkowy stan widoku (ostatnie podsumowanie lub stan pusty).
- `POST /summaries`: Inicjuje generowanie nowego podsumowania i zwraca zaktualizowany widok.

## 3. Struktura komponentów

Hierarchia komponentów `templ` będzie następująca:

- `shared.view.components.Navbar`
  - `summary.view.NavbarButton`: Przycisk w nawigacji otwierający modal. Jego stan (aktywny/nieaktywny) zależy od tego, czy użytkownik może generować podsumowania.
- `shared.view.Layout`
  - `summary.view.Modal`: Komponent modala (oparty na DaisyUI), który zawiera dynamiczną treść.
    - `summary.view.Display`: Główny komponent widoku, który renderuje odpowiedni stan w zależności od otrzymanego `ViewModel`.
      - `summary.view.Content`: Wyświetla treść i datę istniejącego podsumowania oraz przycisk do generowania nowego.
      - `summary.view.EmptyState`: Wyświetlany, gdy nie ma jeszcze żadnego podsumowania. Zawiera zachętę i przycisk do generowania.
      - `summary.view.LoadingState`: Wskaźnik ładowania pokazywany podczas generowania podsumowania.
      - `summary.view.ErrorState`: Wyświetla komunikat błędu, jeśli generowanie się nie powiedzie.

## 4. Szczegóły komponentów

### `summary.view.NavbarButton`

- **Opis:** Przycisk "Summary" w głównej nawigacji. Otwiera modal podsumowania. Jest nieaktywny, jeśli użytkownik nie ma żadnych aktywnych feedów.
- **Główne elementy:** Element `<button>` lub `<a>` opakowany w `<label for="summary-modal-toggle">`. Atrybut `disabled` jest ustawiany na podstawie propsa. Tooltip (np. DaisyUI `tooltip`) wyświetla informację, dlaczego przycisk jest nieaktywny.
- **Obsługiwane interakcje:** `click` otwiera modal.
- **Propsy:**
  - `CanGenerate (bool)`: Określa, czy przycisk ma być aktywny.

### `summary.view.Modal`

- **Opis:** Kontener modala oparty na DaisyUI. Zawiera "przełącznik" (ukryty checkbox) i treść modala. Treść jest początkowo pusta i ładowana dynamicznie przez htmx.
- **Główne elementy:** `<input type="checkbox" id="summary-modal-toggle" class="modal-toggle" />`, `<div class="modal">`. Wewnątrz znajduje się kontener docelowy dla htmx, np. `<div id="summary-modal-content">`.
- **Obsługiwane interakcje:**
  - `htmx`: `hx-get="/summaries/latest"`, `hx-trigger="load"`, `hx-target="#summary-modal-content"` - do załadowania początkowej zawartości przy pierwszym otwarciu.
- **Propsy:** Brak.

### `summary.view.Display`

- **Opis:** Komponent renderujący logikę warunkową wewnątrz modala. Na podstawie `SummaryDisplayViewModel` decyduje, czy renderować `Content`, `EmptyState`, `LoadingState` czy `ErrorState`.
- **Główne elementy:** Struktura `if/else` w `templ` do renderowania warunkowego.
- **Propsy:**
  - `ViewModel (models.SummaryDisplayViewModel)`

### `summary.view.Content`

- **Opis:** Wyświetla treść i datę utworzenia podsumowania.
- **Główne elementy:** Tytuł "Daily Summary", akapit z treścią podsumowania (`ViewModel.Summary.Content`), informacja o dacie (`ViewModel.Summary.CreatedAt`), przycisk "Generate New Summary".
- **Obsługiwane interakcje:**
  - Przycisk "Generate New Summary": `hx-post="/summaries"`, `hx-target="#summary-modal-content"`, `hx-indicator="#summary-loading-state"`.
- **Propsy:**
  - `ViewModel (models.SummaryViewModel)`

### `summary.view.EmptyState`

- **Opis:** Wyświetla informację dla użytkowników, którzy nie mają jeszcze podsumowania.
- **Główne elementy:** Tekst "No summary generated yet...", przycisk "Generate Summary".
- **Obsługiwane interakcje:**
  - Przycisk "Generate Summary": `hx-post="/summaries"`, `hx-target="#summary-modal-content"`, `hx-indicator="#summary-loading-state"`.
- **Propsy:**
  - `CanGenerate (bool)`: Określa, czy przycisk generowania ma być widoczny.

### `summary.view.LoadingState`

- **Opis:** Wskaźnik ładowania (np. spinner DaisyUI) z tekstem "Generating your summary...".
- **Główne elementy:** `<div id="summary-loading-state" class="htmx-indicator">`, komponent `loading` z DaisyUI.
- **Propsy:** Brak.

### `summary.view.ErrorState`

- **Opis:** Wyświetla komunikat o błędzie.
- **Główne elementy:** Tekst błędu (`ViewModel.ErrorMessage`), przycisk "Try Again".
- **Obsługiwane interakcje:**
  - Przycisk "Try Again": `hx-post="/summaries"`, `hx-target="#summary-modal-content"`, `hx-indicator="#summary-loading-state"`.
- **Propsy:**
  - `ViewModel (models.SummaryErrorViewModel)`

## 5. Typy

Implementacja będzie korzystać z istniejących typów DTO zdefiniowanych w backendzie:

- **`models.SummaryDisplayViewModel`**: Główny model widoku przekazywany do komponentu `Display`.
  - `Summary (*SummaryViewModel)`: Dane istniejącego podsumowania (jeśli istnieje).
  - `CanGenerate (bool)`: Flaga wskazująca, czy użytkownik spełnia warunki do generowania podsumowania.
  - `ErrorMessage (string)`: Niepusty string powoduje renderowanie stanu błędu zamiast innych stanów.
  - Logika renderowania: Komponent `Display` sprawdza w kolejności:
    1. Jeśli `ErrorMessage` nie jest puste → renderuje `Error` state
    2. Jeśli `Summary` nie jest `nil` → renderuje `Content` state
    3. W przeciwnym razie → renderuje `EmptyState`
- **`models.SummaryViewModel`**: Reprezentuje pojedyncze podsumowanie.
  - `ID (string)`
  - `Content (string)`
  - `CreatedAt (time.Time)`
- **`models.SummaryErrorViewModel`**: Używany wewnętrznie przez komponent `Error` (nie jest przekazywany bezpośrednio z handlera).
  - `ErrorMessage (string)`

## 6. Zarządzanie stanem

Zarządzanie stanem jest realizowane głównie po stronie serwera i przez htmx.

- **htmx** jest odpowiedzialny za:
  - Wymianę fragmentów HTML w modalu (`hx-swap`).
  - Wyświetlanie wskaźnika ładowania (`htmx-indicator`).
  - Obsługę wywołań API (`hx-get`, `hx-post`).
- **Alpine.js** może być użyty do drobnych interakcji po stronie klienta, jeśli zajdzie taka potrzeba (np. bardziej złożone animacje), ale dla podstawowej funkcjonalności nie jest wymagany. Stan otwarcia/zamknięcia modala jest zarządzany przez mechanizmy DaisyUI (checkbox-hack).

## 7. Integracja API

- **Pobieranie początkowego widoku:**
  - Po otwarciu modala, htmx wykona żądanie `GET /summaries/latest`.
  - Backend zwróci fragment HTML wyrenderowany na podstawie `SummaryDisplayViewModel`.
  - htmx umieści ten fragment wewnątrz kontenera modala (`#summary-modal-content`).
- **Generowanie nowego podsumowania:**
  - Kliknięcie przycisku "Generate..." wyśle żądanie `POST /summaries` za pomocą htmx.
  - Podczas przetwarzania żądania, htmx pokaże wskaźnik ładowania.
  - Po pomyślnym zakończeniu, backend zwróci `200 OK` z nowym fragmentem HTML (widok `Content`), który zastąpi poprzednią zawartość modala.
  - W przypadku błędu, backend zwróci odpowiedni kod statusu (np. 404, 500) z fragmentem HTML widoku `ErrorState`.

## 8. Interakcje użytkownika

- **Użytkownik klika "Summary" w nawigacji:** Modal się otwiera, a jego treść jest dynamicznie ładowana z `GET /summaries/latest`.
- **Użytkownik klika "Generate Summary" / "Generate New Summary" / "Try Again":** Rozpoczyna się proces generowania (`POST /summaries`). Użytkownik widzi stan ładowania, a następnie wynik (nowe podsumowanie lub błąd).
- **Użytkownik klika przycisk zamknięcia lub tło modala:** Modal się zamyka.

## 9. Warunki i walidacja

- **Warunek:** Użytkownik musi mieć co najmniej jeden aktywny feed, aby móc generować podsumowanie.
- **Weryfikacja:** Backend weryfikuje ten warunek i przekazuje wynik w polu `CanGenerate` w `SummaryDisplayViewModel`.
- **Wpływ na interfejs:**
  - Jeśli `CanGenerate` jest `false`, przycisk `NavbarButton` jest nieaktywny (`disabled`) i wyświetla tooltip z wyjaśnieniem.
  - Przyciski generowania wewnątrz modala (`EmptyState`, `Content`) również są renderowane warunkowo na podstawie tej flagi.

## 10. Obsługa błędów

- **Błąd pobierania ostatniego podsumowania (`GET`):** Backend renderuje widok `Display` z wypełnionym polem `ErrorMessage` (np. "Failed to load summary. Please try again.").
- **Błąd generowania podsumowania (`POST`):**
  - **Brak artykułów (404):** Backend renderuje widok `Display` z `ErrorMessage`: "No articles found from the last 24 hours".
  - **Błąd usługi AI (503):** Backend renderuje widok `Display` z `ErrorMessage`: "AI service is temporarily unavailable".
  - **Inne błędy serwera (500):** Backend renderuje widok `Display` z `ErrorMessage`: "Failed to generate summary. Please try again.".
- Komponent `Display` automatycznie wykrywa niepuste `ErrorMessage` i renderuje odpowiedni stan błędu z przyciskiem "Try Again".

## 11. Kroki implementacji

1.  **Utworzenie komponentów widoku (`templ`):**
    - Stwórz plik `internal/summary/view/modal.templ` zawierający wszystkie komponenty: `Modal`, `Display`, `Content`, `EmptyState`, `LoadingState`, `ErrorState`.
    - Stwórz plik `internal/summary/view/navbar_button.templ` dla przycisku w nawigacji.
2.  **Integracja z layoutem:**
    - W głównym pliku layoutu (`internal/shared/view/layout.templ`) dodaj komponent `summary.view.Modal`.
    - W komponencie nawigacji (`internal/shared/view/components/navbar.templ`) dodaj `summary.view.NavbarButton`.
3.  **Modyfikacja handlerów:**
    - Zaktualizuj handler `dashboard.GetDashboard` tak, aby pobierał informację `CanGenerate` i przekazywał ją do `NavbarButton`.
    - Zaktualizuj handler `summary.GetLatestSummary`, aby renderował komponent `summary.view.Display` z odpowiednim `ViewModel`. W przypadku błędu, ustaw `ErrorMessage` w `SummaryDisplayViewModel`.
    - Zaktualizuj handler `summary.PostSummary`, aby zawsze renderował `summary.view.Display`. W przypadku błędu, ustaw odpowiedni `ErrorMessage` w `SummaryDisplayViewModel`.
4.  **Dodanie atrybutów htmx:**
    - Dodaj atrybuty `hx-*` do przycisków w komponentach `Content`, `EmptyState` i `ErrorState`, aby wyzwalały żądanie `POST /summaries`.
    - Skonfiguruj `htmx-indicator` tak, aby wskazywał na komponent `LoadingState`.
    - Dodaj atrybuty `hx-*` do `Modal`, aby ładował początkową treść przy otwarciu.
5.  **Stylowanie i Dostępność:**
    - Użyj klas DaisyUI i Tailwind CSS do ostylowania wszystkich stanów modala.
    - Dodaj odpowiednie atrybuty ARIA, aby zapewnić dostępność, zwłaszcza dla dynamicznie zmieniającej się treści.
    - Zaimplementuj tooltip dla nieaktywnego przycisku `NavbarButton`.
6.  **Testowanie:**
    - Przetestuj wszystkie ścieżki interakcji: pomyślne generowanie, generowanie z pustego stanu, obsługa błędów (brak artykułów, błąd serwera), stan nieaktywnego przycisku.
