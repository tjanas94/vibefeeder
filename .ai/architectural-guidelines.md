# Architektoniczne Zasady i Plan Refaktoryzacji

Ten dokument podsumowuje uzgodnione zasady architektoniczne oraz ogólny plan refaktoryzacji aplikacji VibeFeeder. Ma on służyć jako wytyczne podczas dalszych prac nad kodem.

---

## 1. Ogólne Zasady Architektoniczne

### 1.1. Komunikacja i Odpowiedzialność Warstw

- **Handler:** Jego głównym zadaniem jest obsługa żądania/odpowiedzi HTTP.
  - Odpowiada za **walidację składniową** (formatu) danych wejściowych.
  - Wywołuje odpowiednią metodę w serwisie.
  - Tłumaczy wynik (dane lub błąd) z serwisu na odpowiedź HTTP (renderuje widok lub zwraca kod błędu).
  - **Nie zawiera logiki biznesowej.**

- **Service:** Jest sercem aplikacji i zawiera logikę biznesową.
  - Odpowiada za **walidację biznesową** (np. sprawdzanie unikalności, uprawnień).
  - Orkiestruje operacje, komunikując się z repozytoriami i innymi serwisami.
  - Zwraca dane lub ustrukturyzowany błąd biznesowy (`*ServiceError`).

- **Repository:** Odpowiada wyłącznie za dostęp do danych.
  - Wykonuje operacje CRUD na bazie danych.
  - Zwraca dane lub standardowe błędy `error` (np. błąd połączenia, `sql.ErrNoRows`).
  - **Nie zawiera logiki biznesowej** i nie wie nic o `*ServiceError`.

### 1.2. Cykl Życia Żądania w Handlerze

Każda funkcja handlera powinna realizować następujący, uporządkowany przepływ obsługi żądania, rozróżniając cztery główne ścieżki:

1.  **Ścieżka 1: Błąd Bindowania (`c.Bind`)**
    - **Kiedy:** Gdy dane przychodzące w żądaniu mają niepoprawny format (np. błędny JSON), co świadczy o błędzie programistycznym (niespójność frontend-backend).
    - **Akcja:** Handler opakowuje oryginalny błąd w `echo.NewHTTPError(http.StatusBadRequest, err)` i zwraca go. Pozwala to na przekazanie błędu do globalnego error handlera, który go zaloguje i wyświetli ogólny alert, ale z zachowaniem poprawnego kodu statusu `400 Bad Request` zamiast `500`.

2.  **Ścieżka 2: Błąd Walidacji Składniowej (`c.Validate`)**
    - **Kiedy:** Gdy dane są poprawnie sformatowane, ale nie spełniają podstawowych reguł (np. brak wymaganego pola, zły format e-maila).
    - **Akcja:** Handler przerywa przetwarzanie i zwraca odpowiedź `422 Unprocessable Entity`, najczęściej renderując formularz ponownie z listą błędów przy polach.

3.  **Ścieżka 3: Błąd Biznesowy z Serwisu (`*ServiceError`)**
    - **Kiedy:** Po pomyślnym bindowaniu i walidacji składniowej, serwis zwraca przewidywalny błąd biznesowy (np. zasób nie istnieje, nazwa jest już zajęta).
    - **Akcja:** Handler używa `errors.As()`, aby zidentyfikować błąd jako `*ServiceError`. Następnie wykorzystuje dane z obiektu błędu (kod HTTP, błędy pól), aby wyrenderować odpowiednią, kontekstową odpowiedź (np. formularz z błędem `409 Conflict`).

4.  **Ścieżka 4: Błąd Systemowy (standardowy `error`)**
    - **Kiedy:** Serwis zwraca nieoczekiwany błąd, który nie jest błędem biznesowym (np. brak połączenia z bazą danych).
    - **Akcja:** Handler nie próbuje interpretować błędu. Przekazuje go (`return err`) do globalnego middleware'u do obsługi błędów, który go loguje i zwraca użytkownikowi generyczną odpowiedź `500 Internal Server Error`.

### 1.3. Ustrukturyzowana Obsługa Błędów

- **Błędy Biznesowe:** Wszystkie przewidywalne błędy (np. "email zajęty", "nie znaleziono zasobu") są reprezentowane przez jeden, współdzielony typ `shared.ServiceError`. Błąd ten jest tworzony **wyłącznie w warstwie serwisu**.
- **Błędy Systemowe:** Wszystkie nieoczekiwane problemy (np. brak połączenia z bazą) są reprezentowane przez standardowy typ `error` i obsługiwane przez globalny error handler (prowadząc do odpowiedzi 500).
- **Rozróżnianie błędów:** W handlerze **zawsze** używamy `errors.As()` do sprawdzenia, czy zwrócony błąd jest błędem biznesowym (`*ServiceError`), co pozwala uniknąć skomplikowanych bloków `if/else` i `switch`.

### 1.4. Zależności i Abstrakcje

- **Izolacja od Zależności Zewnętrznych:** Dostęp do zewnętrznych usług (jak Supabase `gotrue`) realizujemy przez **Adapter** – warstwę pośredniczącą, która implementuje nasz własny, mały interfejs. Chroni to naszą logikę przed zmianami w zewnętrznych bibliotekach.
- **Unikanie Nadmiernej Abstrakcji:** Świadomie rezygnujemy z wzorców, które mogłyby skomplikować kod, takich jak generyczne handlery czy `BaseRepository`. Preferujemy bardziej jawne i proste rozwiązania.

### 1.5. Dostęp do Danych (Repozytoria)

- **Wzorzec Query Object:** Metody repozytoriów, które przyjmują wiele parametrów do filtrowania lub paginacji (np. `ListFeeds`), powinny przyjmować je jako pojedynczą strukturę (wzorzec Query Object), a nie jako długą listę argumentów funkcji. Zwiększa to czytelność i ułatwia rozszerzanie zapytań w przyszłości.

---

## 2. Generalny Plan Refaktoryzacji Aplikacji

**Cel:** Wprowadzenie spójnego wzorca obsługi błędów, walidacji i komunikacji między warstwami w całej aplikacji.

### Krok 1: Stworzenie Fundamentów

1.  **Akcja:** Stworzenie centralnego, współdzielonego typu błędu `ServiceError` w pliku `internal/shared/errors/types.go`.
    - **Uzasadnienie:** Zapewni to jeden, spójny sposób komunikowania błędów biznesowych przez wszystkie serwisy.

### Krok 2: Refaktoryzacja Warstwy Serwisów (`internal/**/service.go`)

1.  **Akcja:** Zidentyfikowanie wszystkich plików `service.go` w projekcie.
2.  **Akcja:** Dla każdego serwisu (`auth`, `feed`, `summary` itd.):
    - Zmodyfikować metody, które mogą generować błędy biznesowe, aby zwracały `*shared.ServiceError`.
    - Usunąć lokalne, predefiniowane zmienne błędów (np. `ErrFeedAlreadyExists`) na rzecz nowej, ustrukturyzowanej formy.
    - Zaimplementować uzgodniony **Adapter dla `gotrue.Client`** i podmienić zależność w `auth.Service`.
    - Zaimplementować prostą metodę do logowania zdarzeń zamiast wzorca Dekorator.

### Krok 3: Refaktoryzacja Warstwy Handlerów (`internal/**/handler.go`)

1.  **Akcja:** Zidentyfikowanie wszystkich plików `handler.go` w projekcie.
2.  **Akcja:** Dla każdego handlera (`auth`, `feed`, `summary` itd.):
    - Uprościć logikę obsługi błędów, zastępując bloki `switch` i `if-else` na błędach z serwisu pojedynczym blokiem `if err != nil` z `errors.As()`.
    - Potwierdzić, że **walidacja składniowa** (`c.Validate()`) jest wykonywana w handlerze **przed** wywołaniem serwisu.

### Krok 4: Weryfikacja Warstwy Repozytoriów (`internal/**/repository.go`)

1.  **Akcja:** Zidentyfikowanie wszystkich plików `repository.go` w projekcie.
2.  **Akcja:** Dla każdego repozytorium:
    - Upewnić się, że repozytoria zwracają standardowe błędy `error`, a **nie** `*shared.ServiceError`. Warstwa serwisu jest odpowiedzialna za opakowanie tych błędów.
    - Sprawdzić spójność w użyciu helperów, np. `database.IsNotFoundError`.

### Krok 5: Uruchomienie Testów

1.  **Akcja:** Po zakończeniu refaktoryzacji, kluczowe będzie uruchomienie wszystkich testów w projekcie, aby upewnić się, że żadna funkcjonalność nie została naruszona.
2.  **Narzędzie:** `run_shell_command` z odpowiednią komendą testującą z `Taskfile.yml`.
