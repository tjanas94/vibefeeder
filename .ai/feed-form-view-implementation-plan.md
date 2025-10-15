# Plan implementacji widoku: Formularz i usuwanie feedu

## 1. Przegląd

Celem jest implementacja interfejsu użytkownika do zarządzania feedami RSS, zgodnie z historyjkami użytkownika US-005, US-007 i US-008. Obejmuje to stworzenie komponentów `templ` dla formularza dodawania/edycji feedu oraz modala potwierdzającego usunięcie. Interakcje będą realizowane dynamicznie przy użyciu `htmx` do komunikacji z serwerem i `Alpine.js` do zarządzania stanem po stronie klienta, a wszystko to w ramach okien modalnych `DaisyUI`.

## 2. Routing widoku

Komponenty nie będą dostępne pod statycznym adresem URL. Będą one dynamicznie ładowane w obrębie głównego panelu aplikacji (`Dashboard`):

- **Formularz dodawania/edycji:** Treść formularza będzie ładowana do modala za pomocą żądania `GET /feeds/new` lub `GET /feeds/{id}/edit`.
- **Modal potwierdzenia usunięcia:** Treść modala będzie ładowana dynamicznie za pomocą żądania `GET /feeds/{id}/delete`.

## 3. Struktura komponentów

Hierarchia komponentów będzie zintegrowana z istniejącym widokiem panelu głównego.

```
- DashboardView (Widok główny)
  - FeedListContainer (Kontener listy feedów, nasłuchujący na `refreshFeedList`)
    - FeedListItem (Pojedynczy element listy)
      - Przycisk "Edit" (`onclick` otwierający modal, `hx-get` ładujący formularz)
      - Przycisk "Delete" (`onclick` otwierający modal, `hx-get` ładujący treść potwierdzenia)
  - GlobalModals (Komponent zawierający definicje modali)
    - FeedFormModal (Modal dla formularza)
      - #feed-form-container (Kontener na dynamicznie ładowaną treść formularza)
        - FeedFormView (Komponent `templ` z formularzem)
    - DeleteConfirmationModal (Modal potwierdzenia usunięcia)
      - #delete-confirmation-container (Kontener na dynamicznie ładowaną treść potwierdzenia)
        - DeleteConfirmationView (Komponent `templ` z treścią potwierdzenia)
```

## 4. Szczegóły komponentów

### `FeedFormView`

- **Opis komponentu:** Komponent `templ` renderujący formularz do dodawania lub edycji feedu. Jest ładowany dynamicznie do modala.
- **Główne elementy:**
  - `<form>` z atrybutami `hx-post` ustawionymi dynamicznie na `/feeds` lub `/feeds/{id}`.
  - `input` dla nazwy (`name="name"`, label "Name") i adresu URL (`name="url"`, label "URL").
  - Komunikaty błędów wyświetlane pod odpowiednimi polami formularza (z `Errors.NameError` i `Errors.URLError`).
  - Ogólny komunikat błędu na górze formularza (z `Errors.GeneralError`) dla błędów nie związanych z konkretnymi polami.
  - Przyciski "Save" i "Cancel".
- **Obsługiwane interakcje:**
  - `hx-post`: Wysłanie formularza z `hx-swap="outerHTML"` aby zastąpić cały formularz w przypadku błędu.
  - `hx-on::after-request`: Zamknięcie modala po pomyślnym zapisie (status `204`). `if(event.detail.xhr.status === 204) this.closest('dialog').close()`
- **Obsługiwana walidacja:**
  - `name`: `required`, `max=255` (po stronie klienta: `required`).
  - `url`: `required`, `url` (po stronie klienta: `required`, `type="url"`).
- **Typy:** `ViewModel: FeedFormViewModel`.
- **Propsy:** `Props: FeedFormViewModel`.

### `DeleteConfirmationView`

- **Opis komponentu:** Komponent `templ` renderujący treść modala potwierdzającego usunięcie. Jest ładowany dynamicznie do modala.
- **Główne elementy:**
  - Tekst potwierdzenia z nazwą feedu, np. "Are you sure you want to delete `<nazwa feedu>`?".
  - Przycisk "Confirm" z atrybutem `hx-delete` ustawionym na `/feeds/{id}`.
  - Przycisk "Cancel" zamykający modal.
  - Wskaźnik ładowania `htmx-indicator` powiązany z przyciskiem "Confirm".
  - Kontener na błędy (`<div id="feed-delete-modal-errors">`).
- **Obsługiwane interakcje:**
  - `hx-delete`: Wysłanie żądania usunięcia.
  - `hx-on::after-request`: Zamknięcie modala po pomyślnym usunięciu (status `204`). `if(event.detail.xhr.status === 204) this.closest('dialog').close()`
- **Obsługiwana walidacja:** Brak.
- **Typy:** `ViewModel: DeleteConfirmationViewModel`.
- **Propsy:** `Props: DeleteConfirmationViewModel`.

## 5. Typy

### `FeedFormViewModel` (Nowy ViewModel)

- **Cel:** Renderowanie komponentu `FeedFormView` w odpowiednim trybie i z odpowiednimi danymi.
- **Pola:**
  - `Mode string`: Tryb formularza ("add" lub "edit").
  - `Title string`: Tytuł modala ("Add New Feed" / "Edit Feed").
  - `PostURL string`: Docelowy URL dla `hx-post` (`/feeds` lub `/feeds/{id}`).
  - `FeedID string`: ID edytowanego feedu (opcjonalne).
  - `Name string`: Aktualna nazwa (do wypełnienia formularza - z inputu użytkownika lub z bazy).
  - `URL string`: Aktualny URL (do wypełnienia formularza - z inputu użytkownika lub z bazy).
  - `Errors models.FeedFormErrorViewModel`: Błędy walidacji do wyświetlenia pod polami.

### `FeedFormErrorViewModel` (Nowy ViewModel)

- **Cel:** Przechowywanie błędów walidacji formularza feedu.
- **Pola:**
  - `NameError string`: Błąd walidacji dla pola "name" (np. "Name is required", "Name is too long").
  - `URLError string`: Błąd walidacji dla pola "url" (np. "URL is required", "Invalid URL format", "Feed with this URL already exists").
  - `GeneralError string`: Ogólny błąd nie związany z konkretnym polem (np. "Failed to save feed", "Internal server error").

### `DeleteConfirmationViewModel` (Nowy ViewModel)

- **Cel:** Renderowanie komponentu `DeleteConfirmationView` z danymi feedu do usunięcia.
- **Pola:**
  - `FeedID string`: ID feedu do usunięcia.
  - `FeedName string`: Nazwa feedu do wyświetlenia w komunikacie.
  - `DeleteURL string`: Docelowy URL dla `hx-delete` (`/feeds/{id}`).

## 6. Zarządzanie stanem

Stan będzie zarządzany głównie przez `htmx` poprzez dynamiczne ładowanie treści modali. `Alpine.js` może być używany do zarządzania stanem UI modala (otwieranie/zamykanie), ale nie jest wymagany do przechowywania danych feedu, ponieważ są one przekazywane przez serwer w `ViewModel`.

- **Implementacja:**
  - Przycisk "Edit" na elemencie listy będzie miał atrybuty `hx-get="/feeds/{id}/edit"`, `hx-target="#feed-form-container"` oraz `onclick="feed_form_modal.showModal()"`.
  - Przycisk "Delete" na elemencie listy będzie miał atrybuty `hx-get="/feeds/{id}/delete"`, `hx-target="#delete-confirmation-container"` oraz `onclick="delete_confirmation_modal.showModal()"`.
  - Przycisk "Add Feed" będzie miał atrybuty `hx-get="/feeds/new"`, `hx-target="#feed-form-container"` oraz `onclick="feed_form_modal.showModal()"`.

## 7. Integracja API

- **Formularz dodawania (`GET /feeds/new`):**
  - **Żądanie:** Brak ciała.
  - **Odpowiedź:** Fragment HTML z komponentem `FeedFormView` w trybie "add".
- **Formularz edycji (`GET /feeds/{id}/edit`):**
  - **Żądanie:** Brak ciała.
  - **Odpowiedź (Sukces):** Fragment HTML z komponentem `FeedFormView` w trybie "edit" z wypełnionymi danymi.
  - **Odpowiedź (Błąd):** `404` jeśli feed nie istnieje.
- **Potwierdzenie usunięcia (`GET /feeds/{id}/delete`):**
  - **Żądanie:** Brak ciała.
  - **Odpowiedź (Sukces):** Fragment HTML z komponentem `DeleteConfirmationView` z danymi feedu.
  - **Odpowiedź (Błąd):** `404` jeśli feed nie istnieje.
- **Dodawanie (`POST /feeds`):**
  - **Żądanie:** `multipart/form-data` z `name` i `url`.
  - **Odpowiedź (Sukces):** `204 No Content` z nagłówkiem `HX-Trigger: refreshFeedList`.
  - **Odpowiedź (Błąd):** `400/409/500` z pełnym komponentem `FeedFormView` zawierającym wprowadzone dane użytkownika i wypełniony `FeedFormErrorViewModel` z komunikatami błędów. Formularz zastępuje poprzednią wersję dzięki `hx-swap="outerHTML"`.
- **Edycja (`POST /feeds/{id}`):**
  - **Żądanie:** `multipart/form-data` z `name` i `url`.
  - **Odpowiedź (Sukces):** `204 No Content` z `HX-Trigger: refreshFeedList`.
  - **Odpowiedź (Błąd):** `400/404/409/500` z pełnym komponentem `FeedFormView` zawierającym wprowadzone dane użytkownika i wypełniony `FeedFormErrorViewModel` z komunikatami błędów. Formularz zastępuje poprzednią wersję dzięki `hx-swap="outerHTML"`.
- **Usuwanie (`DELETE /feeds/{id}`):**
  - **Żądanie:** Brak ciała.
  - **Odpowiedź (Sukces):** `204 No Content` z `HX-Trigger: refreshFeedList`.
  - **Odpowiedź (Błąd):** `404/500` z pełnym komponentem `DeleteConfirmationView` zawierającym komunikat błędu w dedykowanym kontenerze na błędy. Modal pozostaje otwarty, umożliwiając ponowną próbę.

## 8. Interakcje użytkownika

- **Dodawanie feedu:** Kliknięcie "Add Feed" otwiera modal i jednocześnie ładuje formularz przez `htmx`. Po zapisie modal jest zamykany, a lista feedów odświeżana.
- **Edycja feedu:** Kliknięcie "Edit" otwiera modal i jednocześnie ładuje formularz z wypełnionymi danymi przez `htmx`. Po zapisie modal jest zamykany, a lista odświeżana. W przypadku błędu, formularz jest ponownie renderowany z wprowadzonymi danymi i komunikatami błędów.
- **Anulowanie:** Kliknięcie "Cancel" w formularzu lub poza modalem zamyka go bez żadnych zmian.
  </parameter>
  </invoke>
- **Usuwanie feedu:** Kliknięcie "Delete" otwiera modal i jednocześnie ładuje treść potwierdzenia przez `htmx`. Kliknięcie "Confirm" wysyła żądanie usunięcia, pokazuje wskaźnik ładowania, a po sukcesie zamyka modal i odświeża listę.

## 9. Warunki i walidacja

- Pola `name` i `url` w formularzu będą miały atrybuty `required`. Pole `url` będzie miało `type="url"`.
- Walidacja serwerowa (istnienie URL, duplikaty) będzie obsługiwana przez ponowne renderowanie całego formularza z wprowadzonymi danymi użytkownika i komunikatami błędów pod odpowiednimi polami. Formularz nie zostanie zamknięty, a wprowadzone dane zostaną zachowane.

## 10. Obsługa błędów

- **Błędy walidacji (400, 409):** Serwer zwraca kod błędu wraz z pełnym komponentem `FeedFormView` zawierającym:
  - Wprowadzone przez użytkownika wartości w polach `name` i `url`
  - Wypełniony `FeedFormErrorViewModel` z komunikatami błędów (`NameError`, `URLError` lub `GeneralError`)
  - Formularz zastępuje poprzednią wersję (`hx-swap="outerHTML"`)
  - Modal pozostaje otwarty, umożliwiając poprawienie błędów
- **Błąd serwera (500):** Serwer zwraca pełny komponent `FeedFormView` z zachowanymi danymi użytkownika i wypełnionym `GeneralError` (np. "Failed to save feed. Please try again."). Modal pozostaje otwarty.
- **Zasób nieznaleziony (404):** W przypadku edycji nieistniejącego feedu, serwer zwraca formularz z wypełnionym `GeneralError` (np. "Feed not found"). Alternatywnie, modal może zostać zamknięty z wyświetleniem powiadomienia toast.
- **Błędy usuwania (404, 500):** Serwer zwraca pełny komponent `DeleteConfirmationView` z komunikatem błędu wyświetlanym w dedykowanym kontenerze (`#feed-delete-modal-errors`). Modal pozostaje otwarty.
- **Pessimistic UI:** Podczas usuwania, przycisk "Confirm" będzie nieaktywny (`hx-indicator`), a wskaźnik ładowania widoczny do czasu otrzymania odpowiedzi od serwera.</parameter>

<old_text line=149>

1.  **Stworzenie `ViewModeli`:** Zdefiniuj struktury `FeedFormViewModel` i `DeleteConfirmationViewModel` w Go.
2.  **Stworzenie komponentu `FeedFormView.templ`:**
    - Przyjmuje `FeedFormViewModel` jako props.
    - Dynamicznie generuje atrybuty `hx-post`, `id` kontenera błędów i wartości pól.
    - Dodaje logikę `hx-on::after-request` do zamykania modala.

## 11. Kroki implementacji

1.  **Stworzenie `ViewModeli`:** Zdefiniuj struktury `FeedFormViewModel` i `DeleteConfirmationViewModel` w Go.
2.  **Stworzenie komponentu `FeedFormView.templ`:**
    - Przyjmuje `FeedFormViewModel` jako props.
    - Dynamicznie generuje atrybuty `hx-post`, `id` kontenera błędów i wartości pól.
    - Dodaje logikę `hx-on::after-request` do zamykania modala.
3.  **Stworzenie komponentu `DeleteConfirmationView.templ`:**
    - Przyjmuje `DeleteConfirmationViewModel` jako props.
    - Generuje treść potwierdzenia z nazwą feedu i przycisk z atrybutem `hx-delete`.
    - Dodaje logikę `hx-on::after-request` do zamykania modala.
4.  **Modyfikacja `DashboardView`:**
    - Dodaj `dialog` dla `FeedFormModal` z kontenerem `#feed-form-container`.
    - Dodaj `dialog` dla `DeleteConfirmationModal` z kontenerem `#delete-confirmation-container`.
5.  **Modyfikacja `FeedListItem`:**
    - Zaktualizuj przycisk "Edit", aby ładował treść do modala (`hx-get`, `hx-target`) i go otwierał (`onclick`).
    - Zaktualizuj przycisk "Delete", aby ładował treść potwierdzenia do modala (`hx-get`, `hx-target`) i go otwierał (`onclick`).
6.  **Modyfikacja `FeedHandler`:**
    - Zaimplementuj handler `GET /feeds/new`, który renderuje `FeedFormView` z `ViewModel` w trybie "add" i pustymi błędami.
    - Zaimplementuj handler `GET /feeds/{id}/edit`, który renderuje `FeedFormView` z `ViewModel` w trybie "edit", wypełnionymi danymi z bazy i pustymi błędami.
    - Zaimplementuj handler `POST /feeds`, który w przypadku błędu walidacji zwraca pełny komponent `FeedFormView` z wprowadzonymi danymi i wypełnionymi błędami.
    - Zaimplementuj handler `POST /feeds/{id}`, który w przypadku błędu walidacji zwraca pełny komponent `FeedFormView` z wprowadzonymi danymi i wypełnionymi błędami.
    - Zaimplementuj handler `GET /feeds/{id}/delete`, który renderuje `DeleteConfirmationView` z `ViewModel` zawierającym dane feedu.
    - Zaimplementuj handler `DELETE /feeds/{id}`, który w przypadku błędu zwraca pełny komponent `DeleteConfirmationView` z komunikatem błędu.
7.  **Testowanie:** Przetestuj wszystkie ścieżki użytkownika: dodawanie, edycja, usuwanie, w tym scenariusze błędów i walidacji.
