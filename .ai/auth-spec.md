# Specyfikacja Techniczna: Moduł Autentykacji Użytkowników

## 1. Wprowadzenie

Niniejszy dokument opisuje architekturę i projekt techniczny modułu autentykacji, rejestracji i odzyskiwania hasła dla aplikacji VibeFeeder. Specyfikacja jest zgodna z wymaganiami zawartymi w PRD (US-001, US-002, US-003, US-012) oraz opiera się na zdefiniowanym stacku technologicznym (Go, Echo, Templ, htmx, Supabase).

Celem jest stworzenie bezpiecznego i spójnego systemu zarządzania tożsamością użytkownika, który integruje się z istniejącą architekturą aplikacji, wykorzystując renderowanie po stronie serwera (SSR) z dynamicznymi interakcjami po stronie klienta.

## 2. Architektura Interfejsu Użytkownika

### 2.1. Struktura i Zmiany w Komponentach

#### 2.1.1. Nowy Moduł `auth`

Wprowadzony zostanie nowy moduł w `internal/auth`, zawierający logikę i widoki związane z autentykacją.

- `internal/auth/view/`
  - `login.templ`: Strona z formularzem logowania.
  - `register.templ`: Strona z formularzem rejestracji.
  - `forgot_password.templ`: Strona z formularzem do zainicjowania resetu hasła.
  - `reset_password.templ`: Strona z formularzem do ustawienia nowego hasła.
  - `form_fields.templ`: Komponenty współdzielone dla pól formularzy (e-mail, hasło) z obsługą walidacji.

#### 2.1.2. Modyfikacja Layoutów

Wprowadzone zostaną dwa osobne layouty, aby oddzielić widoki publiczne (autentykacja) od widoków prywatnych (aplikacja).

1.  **Layout Aplikacji (`internal/shared/view/layout.templ`)**
    - Ten layout będzie używany dla **wszystkich widoków po zalogowaniu** (np. dashboard, zarządzanie feedami).
    - **Nie będzie posiadał wariantu `non-auth`**. Zawsze będzie zakładał, że użytkownik jest uwierzytelniony.
    - Navbar (`internal/shared/view/components/navbar.templ`) będzie w nim na stałe wyświetlał dane zalogowanego użytkownika (np. e-mail) oraz przycisk "Log out".

2.  **Layout Autentykacji (`internal/auth/view/layout.templ`)**
    - Zostanie stworzony nowy, minimalistyczny layout dedykowany dla stron logowania, rejestracji i odzyskiwania hasła.
    - Nie będzie on zawierał nawigacji aplikacji, a jedynie centralnie umieszczony kontener na formularze.

Dzięki temu podejściu, główny layout aplikacji jest uproszczony, a middleware autentykacji jednoznacznie kieruje niezalogowanych użytkowników do stron korzystających z layoutu autentykacji.

### 2.2. Przepływy Użytkownika i Integracja htmx/Templ

Interakcje w formularzach będą realizowane z użyciem htmx, co pozwoli na dynamiczną walidację i komunikację z backendem bez przeładowywania całej strony.

#### 2.2.1. Rejestracja z Weryfikacją E-mail (US-001)

Proces został podzielony na trzy główne kroki.

1.  **Krok 1: Przesłanie formularza rejestracyjnego**
    - **Widok:** `auth/view/register.templ` renderuje formularz z polami: e-mail, hasło, potwierdzenie hasła i opcjonalnie kod rejestracyjny (jeśli jest wymagany przez konfigurację `AUTH_REGISTRATION_CODE`).
    - **Kod rejestracyjny:** Jeśli zmienna środowiskowa `AUTH_REGISTRATION_CODE` jest ustawiona, pole kodu rejestracyjnego jest wyświetlane i wymagane. W przeciwnym razie pole nie jest wyświetlane i rejestracja jest otwarta dla wszystkich. Walidacja kodu używa `subtle.ConstantTimeCompare` aby zapobiec timing attacks.
    - **Interakcja (htmx):** Formularz wysyła żądanie `POST /auth/register`. Atrybut `hx-target` wskazuje na kontener formularza, a `hx-swap="outerHTML"` powoduje jego podmianę w odpowiedzi.
    - **Odpowiedź serwera:**
      - **Błąd:** W przypadku błędu walidacji (np. hasła niezgodne, e-mail zajęty, nieprawidłowy kod rejestracyjny), serwer zwraca kod `422 Unprocessable Entity` i ponownie renderuje komponent formularza z odpowiednimi komunikatami błędów.
      - **Sukces:** Serwer zwraca kod `200 OK`. Formularz zostaje zastąpiony komunikatem dla użytkownika (np. z widoku `auth/view/registration_pending.templ`) o konieczzości sprawdzenia skrzynki mailowej w celu dokończenia procesu. Użytkownik nie jest na tym etapie zalogowany.

2.  **Krok 2: Weryfikacja adresu e-mail**
    - **Interakcja:** Użytkownik klika unikalny link weryfikacyjny otrzymany w wiadomości e-mail. Powoduje to wysłanie żądania `GET` na predefiniowany w systemie autentykacji adres URL, który finalnie przekierowuje użytkownika na `GET /auth/confirm`.

3.  **Krok 3: Finalizacja rejestracji**
    - **Interakcja:** Handler obsługujący ścieżkę `GET /auth/confirm` nie renderuje widoku.
    - **Odpowiedź serwera:** Serwer zwraca odpowiedź z nagłówkiem `HX-Redirect: /auth/login?confirmed=true`.
    - **Efekt końcowy:** Użytkownik zostaje przekierowany na stronę logowania, gdzie wyświetlony jest toast z komunikatem o pomyślnym potwierdzeniu konta.

#### 2.2.2. Logowanie (US-002)

1.  **Widok:** `auth/view/login.templ` renderuje formularz z polami: e-mail i hasło oraz linkiem "Forgot password?".
2.  **Interakcja (htmx):**
    - Formularz wysyła żądanie `POST /auth/login`.
    - `hx-target` i `hx-swap` działają analogicznie do rejestracji.
3.  **Odpowiedź serwera:**
    - **Błąd:** Serwer zwraca kod `401 Unauthorized` i renderuje formularz z jednym, ogólnym komunikatem: "Invalid email or password".
    - **Sukces:** Serwer zwraca kod `200 OK` z nagłówkiem `HX-Redirect: /dashboard`.

#### 2.2.3. Wylogowanie (US-003)

1.  **Widok:** Przycisk "Log out" w `navbar.templ`.
2.  **Interakcja (htmx):**
    - Przycisk wysyła żądanie `POST /auth/logout`.
3.  **Odpowiedź serwera:**
    - Serwer zwraca kod `200 OK` z nagłówkiem `HX-Redirect: /auth/login`, przekierowując użytkownika do strony logowania.

#### 2.2.4. Odzyskiwanie Hasła (US-012)

1.  **Krok 1: Wprowadzenie adresu e-mail**
    - **Widok:** `auth/view/forgot_password.templ`.
    - **Interakcja:** Formularz wysyła `POST /auth/forgot-password`. `hx-target` wskazuje na kontener formularza.
    - **Odpowiedź:** Serwer zawsze zwraca `200 OK` i renderuje komunikat sukcesu (np. w komponencie `toast.templ`): "If the account exists, we've sent a password reset link."

2.  **Krok 2: Ustawienie nowego hasła**
    - **Widok:** Użytkownik, klikając w link z e-maila, trafia na `GET /auth/reset-password?token=...`. Serwer renderuje widok `auth/view/reset_password.templ` z formularzem zawierającym ukryte pole z tokenem.
    - **Interakcja:** Formularz wysyła `POST /auth/reset-password` z nowym hasłem, jego potwierdzeniem i tokenem.
    - **Odpowiedź:**
      - **Błąd:** `422 Unprocessable Entity` z ponownie wyrenderowanym formularzem i błędami walidacji lub `400 Bad Request` jeśli token jest nieprawidłowy/wygasł.
      - **Sukces:** `200 OK` z nagłówkiem `HX-Redirect: /auth/login` i parametrem query (np. `?reset_success=true`), aby na stronie logowania wyświetlić komunikat o pomyślnej zmianie hasła.

## 3. Logika Backendowa

### 3.1. Struktura Modułu `auth`

- `internal/auth/handler.go`: Definiuje `Echo` handlery dla ścieżek autentykacji. Handler jest odpowiedzialny za parsowanie żądań, wywoływanie serwisu i renderowanie odpowiedzi `templ`.
- `internal/auth/service.go`: Zawiera logikę biznesową. Komunikuje się z klientem Supabase Auth, waliduje dane i zarządza sesją (ciasteczkami).
- `internal/auth/models/dto.go`: Definiuje struktury DTO (Data Transfer Objects) dla żądań (np. `LoginRequest`, `RegisterRequest`) wraz z tagami walidacji.
- `internal/app/routes.go`: Tutaj zostaną zarejestrowane nowe ścieżki i grupy ścieżek (`/auth`).

### 3.2. Endpointy API

- `GET /auth/login`: Renderuje stronę logowania.
- `POST /auth/login`: Przetwarza logowanie.
- `GET /auth/register`: Renderuje stronę rejestracji.
- `POST /auth/register`: Przetwarza rejestrację.
- `GET /auth/confirm`: Obsługuje przekierowanie po weryfikacji e-mail i finalizuje proces rejestracji.
- `POST /auth/logout`: Przetwarza wylogowanie.
- `GET /auth/forgot-password`: Renderuje stronę do resetu hasła.
- `POST /auth/forgot-password`: Wysyła e-mail z linkiem do resetu.
- `GET /auth/reset-password`: Renderuje stronę do ustawienia nowego hasła (oczekuje tokenu w query params).
- `POST /auth/reset-password`: Przetwarza zmianę hasła.

### 3.3. Walidacja i Obsługa Błędów

- Walidacja DTO będzie realizowana w `service.go` przy użyciu istniejącego pakietu `internal/shared/validator`.
- Serwis będzie zwracał dedykowane błędy (np. `ErrUserAlreadyExists`, `ErrInvalidCredentials`), które handler będzie mapował na odpowiednie kody HTTP i widoki `templ` z komunikatami dla użytkownika.

## 4. System Autentykacji (Supabase Auth)

### 4.1. Integracja z Supabase

- Backend będzie wykorzystywał oficjalny lub społecznościowy klient Go dla Supabase Auth (np. `gotrue-go`).
- Klient Supabase zostanie zainicjowany w `internal/auth/service.go` i wstrzyknięty jako zależność.

### 4.2. Zarządzanie Sesją

- **Logowanie/Rejestracja:** Po pomyślnym uwierzytelnieniu przez Supabase, serwis `auth.Service` otrzyma token dostępowy (`AccessToken`).
- **Ciasteczko (Cookie):** Serwis umieści `AccessToken` w bezpiecznym ciasteczku `HttpOnly` z flagą `Secure` (dla HTTPS) i `SameSite=Lax`. Ciasteczko będzie głównym mechanizmem utrzymywania sesji.
- **Wylogowanie:** Serwis wywoła metodę `SignOut` w kliencie Supabase (unieważniając token po stronie Supabase), a następnie usunie ciasteczko sesji z przeglądarki użytkownika.

### 4.3. Middleware Autentykacji

- Zostanie stworzony nowy middleware w `internal/shared/auth/middleware.go`, który będzie stanowił główną bramę do aplikacji.
- Middleware będzie aplikowany do **wszystkich ścieżek aplikacji**, z wyjątkiem grupy `/auth` oraz serwowania plików statycznych. Handler dla ścieżki głównej (`/`) również będzie objęty tym mechanizmem.
- **Logika middleware:**
  1.  Odczytaj token z ciasteczka sesji.
  2.  Jeśli ciasteczko nie istnieje, **przerwij dalsze przetwarzanie i przekieruj na `/auth/login`**.
  3.  Jeśli ciasteczko istnieje, zweryfikuj token, wywołując metodę `GetUser(token)` klienta Supabase.
  4.  Jeśli token jest ważny, pobierz dane użytkownika, umieść je w kontekście `echo.Context` i przekaż sterowanie do właściwego handlera.
  5.  Jeśli token jest nieważny lub wygasł, usuń ciasteczko i **przekieruj na `/auth/login`**.

W ten sposób każdy handler chroniony (czyli cała aplikacja) ma gwarancję dostępu do tożsamości zalogowanego użytkownika, a próba dostępu bez sesji jest centralnie zablokowana.
