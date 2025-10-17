# Architektura UI dla VibeFeeder

## 1. Przegląd struktury UI

Architektura interfejsu użytkownika (UI) dla VibeFeeder opiera się na paradygmacie **HTML over the wire**. Backend (Go, Echo, Templ) jest jedynym źródłem prawdy, renderującym HTML po stronie serwera. Aplikacja nie jest aplikacją jednostronicową (SPA).

Dynamiczne interakcje i częściowe aktualizacje strony są realizowane za pomocą **htmx**, co minimalizuje potrzebę pisania złożonego kodu JavaScript. Minimalny, efemeryczny stan po stronie klienta (np. widoczność okna modalnego) jest zarządzany przez **Alpine.js**. Warstwa wizualna bazuje na **Tailwind 4** i bibliotece komponentów **DaisyUI**, co zapewnia spójność i szybkość rozwoju.

Główny widok po zalogowaniu to "powłoka aplikacji" (app shell), która zawiera nawigację i statyczne elementy. Treść, taka jak lista feedów, jest dynamicznie ładowana do tej powłoki za pomocą zapytań htmx. Operacje CRUD (tworzenie, edycja, usuwanie) odbywają się w oknach modalnych, których zawartość jest również dynamicznie pobierana z serwera.

## 2. Lista widoków

### Widok: Logowanie

- **Ścieżka:** `/login` (dla niezalogowanych użytkowników)
- **Główny cel:** Umożliwienie użytkownikowi zalogowania się do aplikacji.
- **Kluczowe informacje do wyświetlenia:** Formularz z polami na e-mail i hasło.
- **Kluczowe komponenty widoku:**
  - Formularz logowania.
  - Pola tekstowe: "Email", "Password".
  - Przycisk "Sign In".
  - Link do widoku rejestracji: "Don't have an account? Sign up".
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Prosty, wyśrodkowany formularz. W przypadku błędu wyświetlany jest jeden, ogólny komunikat ("Invalid email or password"), aby nie ujawniać, czy konto o danym e-mailu istnieje.
  - **Dostępność:** Poprawne etykiety dla pól formularza. Obsługa nawigacji klawiaturą.
  - **Bezpieczeństwo:** Komunikacja przez HTTPS. Ochrona przed atakami CSRF.

### Widok: Rejestracja

- **Ścieżka:** `/register` (dla niezalogowanych użytkowników)
- **Główny cel:** Umożliwienie nowemu użytkownikowi założenia konta.
- **Kluczowe informacje do wyświetlenia:** Formularz rejestracyjny.
- **Kluczowe komponenty widoku:**
  - Formularz rejestracji.
  - Pola tekstowe: "Email", "Password", "Confirm Password".
  - Pole wyboru (checkbox) do akceptacji polityki prywatności: "I agree to the Privacy Policy".
  - Przycisk "Sign Up".
  - Link do widoku logowania: "Already have an account? Sign in".
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Walidacja pól w czasie rzeczywistym (np. czy hasła są zgodne). Jasne komunikaty o błędach przy konkretnych polach ("Passwords do not match", "Email is already taken").
  - **Dostępność:** Poprawne etykiety i powiązania `aria` dla pól i błędów.
  - **Bezpieczeństwo:** Wymagania dotyczące siły hasła implementowane po stronie serwera.

### Widok: Panel Główny (Dashboard)

- **Ścieżka:** `/dashboard` (dla zalogowanych użytkowników)
- **Główny cel:** Stanowienie "powłoki aplikacji" (app shell) i głównego punktu nawigacyjnego dla zalogowanego użytkownika.
- **Kluczowe informacje do wyświetlenia:** Nawigacja, obszar na dynamiczną treść.
- **Kluczowe komponenty widoku:**
  - Główny pasek nawigacyjny.
  - Kontener (`<div id="feed-list-container">`) na dynamicznie ładowaną listę feedów.
  - Okna modalne (`<dialog>`) dla formularza feedu, potwierdzenia usunięcia i podsumowania.
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Po załadowaniu strony, lista feedów jest pobierana automatycznie, co daje wrażenie szybkości.
  - **Dostępność:** Struktura strony oparta o semantyczne znaczniki HTML5 (`<header>`, `<main>`, `<nav>`).
  - **Bezpieczeństwo:** Dostęp do tej ścieżki jest chroniony przez middleware uwierzytelniający.

### Fragment Widoku: Lista Feedów

- **Ścieżka:** Dynamicznie ładowany z `GET /feeds` do widoku Panelu Głównego.
- **Główny cel:** Wyświetlanie, filtrowanie i zarządzanie listą feedów RSS użytkownika.
- **Kluczowe informacje do wyświetlenia:** Lista feedów, ich nazwy, statusy (działający/błąd), paginacja.
- **Kluczowe komponenty widoku:**
  - Pole wyszukiwania po nazwie z placeholderem "Search feeds...".
  - Grupa przycisków do filtrowania po statusie: "All", "Active", "Pending", "Error".
  - Przycisk "Add Feed".
  - Lista/tabela z elementami feedów.
  - Każdy element zawiera: nazwę, ikonę błędu z tooltipem (jeśli występuje), przyciski "Edit" i "Delete".
  - Kontrolki paginacji: "Previous", "Next", "Page X of Y".
  - **Stany specjalne:**
    - **Empty State:** "No feeds yet. Add your first RSS feed to get started!" z przyciskiem "Add Your First Feed".
    - **No Results State:** "No feeds found matching your search criteria."
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Wyszukiwanie i filtrowanie odbywa się bez przeładowania strony. Adres URL jest aktualizowany, aby odzwierciedlić stan filtrów, co umożliwia udostępnianie linków. Wyszukiwanie ma 500ms debounce.
  - **Dostępność:** Ikony mają tekst alternatywny. Tooltipy są dostępne z klawiatury.
  - **Bezpieczeństwo:** Wszystkie dane wejściowe od użytkownika (wyszukiwanie) są odpowiednio escapowane po stronie serwera.

### Fragment Widoku: Formularz Feedu (w Modalu)

- **Ścieżka:** Dynamicznie ładowany z `GET /feeds/new` lub `GET /feeds/{id}/edit` do modala w Panelu Głównym.
- **Główny cel:** Umożliwienie dodawania lub edycji feedu RSS.
- **Kluczowe informacje do wyświetlenia:** Formularz z polami na nazwę i URL.
- **Kluczowe komponenty widoku:**
  - Tytuł modala: "Add New Feed" / "Edit Feed".
  - Pola tekstowe: "Name", "Feed URL".
  - Przyciski "Save" i "Cancel".
  - Obszary na komunikaty o błędach walidacji (pod polami i ogólny nad formularzem).
  - Przykładowe błędy: "This field is required", "Invalid URL format", "Failed to fetch feed".
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Po pomyślnym zapisie modal jest zamykany, a lista feedów odświeżana. W przypadku błędu walidacji, modal pozostaje otwarty, a formularz jest ponownie wyświetlany z błędami i wprowadzonymi przez użytkownika danymi.
  - **Dostępność:** Focus jest automatycznie przenoszony na pierwsze pole formularza po otwarciu modala i wraca na element wywołujący po zamknięciu.
  - **Bezpieczeństwo:** Walidacja URL po stronie serwera zapobiega próbom ataków SSRF.

### Fragment Widoku: Potwierdzenie Usunięcia (w Modalu)

- **Ścieżka:** Dynamicznie ładowany z `GET /feeds/{id}/delete` do modala w Panelu Głównym.
- **Główny cel:** Zapobieganie przypadkowemu usunięciu feedu.
- **Kluczowe informacje do wyświetlenia:** Pytanie potwierdzające.
- **Kluczowe komponenty widoku:**
  - Tytuł modala: "Delete Feed".
  - Tekst: "Are you sure you want to delete this feed? This action cannot be undone."
  - Przyciski "Delete" (czerwony, destrukcyjny) i "Cancel".
  - Wskaźnik ładowania wyświetlany po kliknięciu "Delete" z tekstem "Deleting...".
  - Obszar na komunikat o błędzie wewnątrz modala: "Failed to delete feed. Please try again."
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Pesymistyczne UI – modal pozostaje otwarty ze wskaźnikiem ładowania do czasu otrzymania odpowiedzi od serwera. W przypadku błędu, jest on wyświetlany w modalu.
  - **Dostępność:** Modal jest w pełni zarządzany klawiaturą.
  - **Bezpieczeństwo:** Operacja usunięcia jest wywoływana metodą `DELETE`, zgodnie z semantyką REST.

### Fragment Widoku: Podsumowanie (w Modalu)

- **Ścieżka:** Dynamicznie ładowany z `GET /summaries/latest` lub `POST /summaries` do modala w Panelu Głównym.
- **Główny cel:** Wyświetlanie ostatniego podsumowania i inicjowanie generowania nowego.
- **Kluczowe informacje do wyświetlenia:** Treść podsumowania, data jego utworzenia.
- **Kluczowe komponenty widoku:**
  - Tytuł modala: "Daily Summary".
  - Przycisk "Generate New Summary".
  - Obszar na treść podsumowania.
  - Data i godzina wygenerowania: "Generated on December 15, 2024 at 10:30 AM".
  - **Stany specjalne:**
    - **Initial State:** "No summary generated yet. Click the button below to generate your first daily summary." z przyciskiem "Generate Summary".
    - **Loading State:** Wskaźnik ładowania w miejscu treści podsumowania z tekstem "Generating your summary...".
    - **Error State:** "Failed to generate summary. Please try again." z przyciskiem "Try Again".
- **UX, dostępność i względy bezpieczeństwa:**
  - **UX:** Generowanie odbywa się w modalu i nie blokuje reszty interfejsu. Przycisk "Summary" w nawigacji jest nieaktywny z tooltipem "Add feeds first to generate summaries", jeśli użytkownik nie ma żadnych feedów.
  - **Dostępność:** Treść podsumowania jest zawarta w elemencie z odpowiednią rolą ARIA, aby czytniki ekranu mogły ją poprawnie zinterpretować.
  - **Bezpieczeństwo:** Treść podsumowania generowana przez AI jest odpowiednio oczyszczana z potencjalnie niebezpiecznego HTML przed renderowaniem.

## 3. Mapa podróży użytkownika

1.  **Rejestracja i Logowanie:**
    - Nowy użytkownik trafia na `/login`, klika link "Don't have an account? Sign up".
    - Wypełnia formularz, zostaje zalogowany i przekierowany na `/dashboard`.
2.  **Pierwsze kroki:**
    - Na `/dashboard` widzi stan pusty listy feedów z komunikatem "No feeds yet. Add your first RSS feed to get started!".
    - Klika "Add Your First Feed", co otwiera modal formularza.
    - Wypełnia i zapisuje formularz. Modal się zamyka, a na liście pojawia się pierwszy feed. Toast wyświetla "Feed added successfully".
3.  **Zarządzanie Feedami:**
    - Użytkownik dodaje kolejne feedy.
    - Używa pola wyszukiwania "Search feeds..." i filtrów statusu ("All", "Active", "Pending", "Error"), aby znaleźć konkretny feed. Lista aktualizuje się dynamicznie.
    - Klika "Edit" przy jednym z feedów, poprawia jego URL w modalu i zapisuje zmiany. Toast wyświetla "Feed updated successfully".
    - Klika "Delete" przy innym feedzie, potwierdza akcję w modalu. Feed znika z listy, toast wyświetla "Feed deleted successfully".
4.  **Generowanie Podsumowania:**
    - Użytkownik klika przycisk "Summary" w nawigacji.
    - Otwiera się modal z informacją "No summary generated yet. Click the button below to generate your first daily summary."
    - Klika "Generate Summary". W modalu pojawia się wskaźnik ładowania z tekstem "Generating your summary...".
    - Po chwili wskaźnik jest zastępowany przez wygenerowany tekst podsumowania i datę jego utworzenia.
5.  **Wylogowanie:**
    - Użytkownik klika "Sign Out" w nawigacji i zostaje przekierowany na stronę logowania.

## 4. Układ i struktura nawigacji

- **Nawigacja publiczna (niezalogowany użytkownik):** Brak stałego paska nawigacyjnego. Widoki logowania i rejestracji zawierają jedynie linki do siebie nawzajem.
- **Nawigacja prywatna (zalogowany użytkownik):**
  - **Lokalizacja:** Stały, górny pasek nawigacyjny (DaisyUI `navbar`) widoczny na wszystkich ekranach po zalogowaniu.
  - **Struktura:**
    - **Lewa strona:**
      - Logo/nazwa aplikacji ("VibeFeeder").
    - **Prawa strona:**
      - Przycisk "Summary" (otwiera modal podsumowania).
      - E-mail zalogowanego użytkownika.
      - Przycisk/link "Sign Out".

## 5. Kluczowe komponenty

- **Modal (`<dialog>`):** Sterowany przez Alpine.js, używany do wszystkich operacji CRUD i wyświetlania podsumowań. Zapewnia spójny wzorzec interakcji i prawidłowe zarządzanie focusem dla dostępności. Treść jest dynamicznie ładowana przez htmx.
- **Toast/Powiadomienie:** Komponent DaisyUI `toast` używany do wyświetlania globalnych, nietrwałych komunikatów o sukcesie (np. "Feed added successfully", "Feed updated successfully", "Feed deleted successfully") lub błędzie ("An error occurred. Please try again."). Wyzwalany przez nagłówki `HX-Trigger` wysyłane z serwera.
- **Grupa przycisków (`btn-group`):** Używana do filtrowania statusu feedów ("All", "Active", "Pending", "Error"). Stan `btn-active` jest zarządzany po stronie klienta przez Alpine.js.
- **Tooltip:** Komponent DaisyUI `tooltip` używany do dostarczania dodatkowych informacji kontekstowych (np. treść błędu przy ikonie, wyjaśnienie nieaktywnego przycisku "Add feeds first to generate summaries") bez zaśmiecania interfejsu.
- **Wskaźnik ładowania:** Elementy z klasą `htmx-indicator` zapewniają wizualną informację zwrotną podczas komunikacji z serwerem (np. przy wyszukiwaniu, generowaniu podsumowania, usuwaniu). Mogą zawierać tekst typu "Loading...", "Generating...", "Deleting...".
