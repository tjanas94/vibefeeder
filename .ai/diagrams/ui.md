<architecture_analysis>

### 1. Lista Komponentów

Na podstawie dostarczonych plików referencyjnych, zidentyfikowano następujące kluczowe komponenty UI, które zostaną utworzone lub zaktualizowane:

**Nowe Layouty:**

- `internal/auth/view/layout.templ`: Minimalistyczny layout dla stron publicznych (logowanie, rejestracja, reset hasła), centrujący formularze na ekranie.

**Nowe Strony (Komponenty `templ`):**

- `internal/auth/view/login.templ`: Strona z formularzem logowania.
- `internal/auth/view/register.templ`: Strona z formularzem rejestracji.
- `internal/auth/view/forgot_password.templ`: Strona z formularzem inicjującym reset hasła.
- `internal/auth/view/reset_password.templ`: Strona z formularzem do ustawienia nowego hasła.
- `internal/auth/view/registration_pending.templ`: Komponent wyświetlany po wysłaniu formularza rejestracji, informujący o konieczności weryfikacji e-mail.

**Nowe lub Zmodyfikowane Komponenty Współdzielone:**

- `internal/auth/view/form_fields.templ`: Współdzielone pola formularzy (e-mail, hasło) z logiką walidacji.
- `internal/shared/view/layout.templ` (Aktualizacja): Główny layout aplikacji, używany wyłącznie dla zalogowanych użytkowników. Usunięty zostanie wariant `non-auth`.
- `internal/shared/view/components/navbar.templ` (Aktualizacja): Pasek nawigacji, który teraz zawsze będzie wyświetlał dane użytkownika i przycisk wylogowania, ponieważ będzie używany tylko w layoucie dla zalogowanych.
- `internal/shared/view/components/toast.templ`: Komponent do wyświetlania powiadomień (np. "Rejestracja pomyślna").

### 2. Główne Strony i Ich Komponenty

- **Strona Logowania (`/auth/login`):**
  - Używa `auth/view/layout.templ`.
  - Renderuje `auth/view/login.templ`, który z kolei używa `auth/view/form_fields.templ`.
- **Strona Rejestracji (`/auth/register`):**
  - Używa `auth/view/layout.templ`.
  - Renderuje `auth/view/register.templ`, który używa `auth/view/form_fields.templ`.
- **Strona "Zapomniałem Hasła" (`/auth/forgot-password`):**
  - Używa `auth/view/layout.templ`.
  - Renderuje `auth/view/forgot_password.templ`.
- **Strona Resetowania Hasła (`/auth/reset-password`):**
  - Używa `auth/view/layout.templ`.
  - Renderuje `auth/view/reset_password.templ`.
- **Panel Główny / Dashboard (`/dashboard`):**
  - Używa zaktualizowanego `shared/view/layout.templ`.
  * Zawiera `shared/view/components/navbar.templ`.
  - Renderuje widok panelu głównego (np. `dashboard/view/index.templ`).

### 3. Przepływ Danych

Przepływ danych opiera się na interakcjach htmx, które wywołują endpointy backendowe i zamieniają fragmenty HTML bez przeładowywania strony.

1.  **Użytkownik -> Formularz (np. Logowania)**: Użytkownik wypełnia dane.
2.  **Formularz -> Backend (`POST /auth/login`)**: htmx wysyła żądanie POST z danymi formularza.
3.  **Backend -> Formularz (Błąd Walidacji)**: W przypadku błędu, backend zwraca kod 4xx i ponownie renderuje komponent formularza z dołączonymi komunikatami o błędach. htmx podmienia istniejący formularz nową wersją.
4.  **Backend -> Przeglądarka (Sukces)**: W przypadku sukcesu, backend zwraca kod 200 z nagłówkiem `HX-Redirect`, który instruuje htmx po stronie klienta, aby przekierować na nową stronę (np. `/dashboard`).
5.  **Dostęp do strony chronionej (`/dashboard`)**: Middleware autentykacji (`shared/auth/middleware.go`) przechwytuje żądanie, weryfikuje token z ciasteczka, a następnie albo zezwala na dostęp, renderując stronę z `shared/view/layout.templ`, albo przekierowuje na `/auth/login`.

### 4. Opis Funkcjonalności Komponentów

- **`auth/view/layout.templ`**: Zapewnia spójny, minimalistyczny wygląd dla wszystkich stron związanych z procesem autentykacji.
- **`shared/view/layout.templ`**: Główny szablon aplikacji dla zalogowanych użytkowników, zawierający nawigację i główną treść strony.
- **`login.templ` / `register.templ`**: Renderują formularze i obsługują logikę interakcji (wysyłanie danych, wyświetlanie błędów) za pomocą atrybutów htmx.
- **`navbar.templ`**: Wyświetla nawigację, e-mail użytkownika i przycisk "Wyloguj", który wysyła żądanie `POST /auth/logout`.
- **`toast.templ`**: Służy do wyświetlania krótkich, globalnych powiadomień o sukcesie lub błędzie (np. po pomyślnej zmianie hasła).
- **Middleware Autentykacji**: Nie jest komponentem UI, ale kluczowym elementem logiki, który decyduje, który layout i stronę wyświetlić w zależności od stanu sesji użytkownika.

</architecture_analysis>

<mermaid_diagram>

```mermaid
flowchart TD
    classDef newComponent fill:#c8e6c9,stroke:#388e3c,stroke-width:2px,color:#000000;
    classDef updatedComponent fill:#fff9c4,stroke:#fbc02d,stroke-width:2px,color:#000000;

    subgraph "Użytkownik Niezalogowany"
        U_GUEST["Użytkownik"]
    end

    subgraph "Strony Publiczne"
        subgraph "Layout Autentykacji"
            L_AUTH["layout.templ"]:::newComponent
        end

        P_LOGIN["Strona Logowania"]:::newComponent
        P_REGISTER["Strona Rejestracji"]:::newComponent
        P_FORGOT["Strona Resetu Hasła"]:::newComponent
        P_RESETPASS["Strona Ustawiania Nowego Hasła"]:::newComponent

        L_AUTH --> P_LOGIN
        L_AUTH --> P_REGISTER
        L_AUTH --> P_FORGOT
        L_AUTH --> P_RESETPASS

        subgraph "Komponenty Formularzy"
            C_LOGIN_FORM["Formularz Logowania"]:::newComponent
            C_REGISTER_FORM["Formularz Rejestracji"]:::newComponent
            C_FORGOT_FORM["Formularz Resetu Hasła"]:::newComponent
            C_RESETPASS_FORM["Formularz Nowego Hasła"]:::newComponent
            C_FORM_FIELDS["Pola Formularza (współdzielone)"]:::newComponent
        end

        P_LOGIN --> C_LOGIN_FORM
        P_REGISTER --> C_REGISTER_FORM
        P_FORGOT --> C_FORGOT_FORM
        P_RESETPASS --> C_RESETPASS_FORM

        C_LOGIN_FORM --> C_FORM_FIELDS
        C_REGISTER_FORM --> C_FORM_FIELDS
    end

    subgraph "Logika Backendowa"
        B_ROUTER["Router Aplikacji"]
        B_AUTH_HANDLER["Auth Handler"]:::newComponent
        B_AUTH_SERVICE["Auth Service"]:::newComponent
        B_VALIDATOR["Walidator"]
        B_DB["Baza Danych (Supabase)"]
    end

    subgraph "Strony Prywatne"
        subgraph "Layout Aplikacji"
            L_APP["layout.templ"]:::updatedComponent
        end

        P_DASHBOARD["Panel Główny"]
        P_FEEDS["Zarządzanie Feedami"]

        subgraph "Komponenty Współdzielone"
            C_NAVBAR["Navbar"]:::updatedComponent
            C_TOAST["Toast"]:::newComponent
        end

        L_APP --> C_NAVBAR
        L_APP --> P_DASHBOARD
        L_APP --> P_FEEDS
    end

    subgraph "Użytkownik Zalogowany"
        U_AUTH["Użytkownik"]
    end

    %% Przepływy
    U_GUEST -- "Otwiera stronę" --> P_LOGIN
    C_LOGIN_FORM -- "POST /auth/login (htmx)" --> B_ROUTER
    B_ROUTER -- "/auth/*" --> B_AUTH_HANDLER
    B_AUTH_HANDLER -- "Waliduj dane" --> B_VALIDATOR
    B_AUTH_HANDLER -- "Logika biznesowa" --> B_AUTH_SERVICE
    B_AUTH_SERVICE -- "Operacje na użytkowniku" --> B_DB

    B_AUTH_HANDLER -- "Błąd walidacji (422)" --> C_LOGIN_FORM
    B_AUTH_HANDLER -- "Sukces (HX-Redirect: /dashboard)" --> B_MIDDLEWARE

    B_MIDDLEWARE["Middleware Autentykacji"]:::newComponent
    U_GUEST -- "Próba dostępu do /dashboard" --> B_MIDDLEWARE
    B_MIDDLEWARE -- "Brak sesji -> Przekieruj" --> P_LOGIN

    B_MIDDLEWARE -- "Sesja aktywna" --> L_APP
    L_APP -- "Renderuje widok dla" --> U_AUTH

    C_NAVBAR -- "POST /auth/logout (htmx)" --> B_ROUTER
    B_AUTH_HANDLER -- "Wylogowano (HX-Redirect: /auth/login)" --> P_LOGIN

    %% Powiązania z Toastem
    P_LOGIN -- "Wyświetla po sukcesie (np. reset hasła)" --> C_TOAST
```

</mermaid_diagram>
