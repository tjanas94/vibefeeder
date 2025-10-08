# Dokument wymagań produktu (PRD) - VibeFeeder
## 1. Przegląd produktu
VibeFeeder to aplikacja webowa, której celem jest uproszczenie konsumpcji treści w internecie. Aplikacja agreguje artykuły z różnych źródeł RSS dostarczonych przez użytkownika, a następnie wykorzystuje sztuczną inteligencję do generowania zwięzłych, codziennych podsumowań. VibeFeeder jest przeznaczony dla osób, które chcą być na bieżąco z informacjami z wielu dziedzin, ale nie mają czasu na śledzenie wszystkich źródeł indywidualnie. Wersja MVP (Minimum Viable Product) skupia się na podstawowej funkcjonalności zarządzania feedami, uwierzytelnianiu użytkowników oraz generowaniu jednego, zbiorczego podsumowania na żądanie.

## 2. Problem użytkownika
Współczesny użytkownik internetu jest codziennie konfrontowany z ogromną ilością informacji pochodzących z portali newsowych, blogów, serwisów branżowych i innych źródeł. Próba śledzenia wszystkich interesujących tematów jest czasochłonna i często prowadzi do przeciążenia informacyjnego (tzw. "information overload"). Użytkownicy potrzebują narzędzia, które pozwoli im efektywnie przyswajać kluczowe informacje z wybranych przez siebie źródeł, bez konieczności manualnego przeglądania każdej strony z osobna. VibeFeeder rozwiązuje ten problem, automatyzując proces zbierania i syntezy treści, dostarczając najważniejsze informacje w skondensowanej, łatwej do przyswojenia formie.

## 3. Wymagania funkcjonalne
Wymagania funkcjonalne dla wersji MVP produktu VibeFeeder są następujące:

1.  Zarządzanie kontem użytkownika:
    *   Użytkownicy muszą mieć możliwość założenia konta za pomocą adresu e-mail i hasła.
    *   Proces rejestracji musi wymagać akceptacji polityki prywatności.
    *   Użytkownicy muszą mieć możliwość zalogowania się do aplikacji i wylogowania się z niej.

2.  Zarządzanie feedami RSS:
    *   Użytkownicy muszą mieć możliwość dodawania nowych źródeł RSS poprzez podanie ich nazwy i adresu URL.
    *   System musi walidować poprawność adresu URL dodawanego feedu.
    *   Użytkownicy muszą mieć możliwość edytowania nazwy i adresu URL istniejących feedów.
    *   Użytkownicy muszą mieć możliwość usuwania dodanych wcześniej feedów.
    *   Interfejs musi wyświetlać listę dodanych feedów wraz z wizualnym wskaźnikiem ich statusu (np. informacja o błędzie pobierania).

3.  Pobieranie i przetwarzanie treści:
    *   System musi automatycznie i cyklicznie pobierać nowe artykuły ze wszystkich aktywnych feedów RSS w systemie.

4.  Generowanie i wyświetlanie podsumowania:
    *   Użytkownik musi mieć możliwość ręcznego zainicjowania procesu generowania podsumowania za pomocą dedykowanego przycisku.
    *   Generowane jest jedno, zbiorcze podsumowanie na podstawie artykułów opublikowanych w ciągu ostatnich 24 godzin.
    *   Podsumowanie ma formę kilku zwięzłych akapitów.
    *   Wygenerowane podsumowanie jest wyświetlane w interfejsie aplikacji wraz z datą i godziną jego utworzenia.

## 4. Granice produktu
Następujące funkcje i cechy celowo nie wchodzą w zakres wersji MVP, aby zapewnić szybkie dostarczenie podstawowej wartości produktu:

*   Kategoryzacja i filtrowanie artykułów.
*   Zaawansowane opcje personalizacji podsumowań (np. wybór tonu, długości, tematów).
*   Obsługa innych formatów niż RSS (np. Atom, JSON Feed).
*   Integracje z zewnętrznymi aplikacjami i czytnikami RSS (np. Feedly, Inoreader).
*   Wysyłanie podsumowań na zewnątrz aplikacji (np. poprzez e-mail, powiadomienia push).
*   Systemy uwierzytelniania oparte na dostawcach zewnętrznych (np. Google, Facebook).
*   Integracja z zewnętrznymi narzędziami analitycznymi do śledzenia zachowań użytkowników.

## 5. Historyjki użytkowników

### Zarządzanie kontem

---
*   ID: US-001
*   Tytuł: Rejestracja nowego użytkownika
*   Opis: Jako nowy użytkownik, chcę móc założyć konto w aplikacji przy użyciu mojego adresu e-mail i hasła, abym mógł zacząć korzystać z usługi.
*   Kryteria akceptacji:
    1.  Formularz rejestracji zawiera pola na adres e-mail, hasło oraz potwierdzenie hasła.
    2.  Formularz zawiera pole wyboru (checkbox) do akceptacji polityki prywatności, które jest wymagane.
    3.  Walidacja po stronie klienta i serwera sprawdza, czy podany e-mail ma poprawny format.
    4.  Walidacja sprawdza, czy hasła w obu polach są identyczne.
    5.  Po pomyślnej rejestracji użytkownik jest automatycznie zalogowany i przekierowany do głównego widoku aplikacji.
    6.  W przypadku, gdy konto o podanym adresie e-mail już istnieje, wyświetlany jest stosowny komunikat o błędzie.
    7.  Zdarzenie `user_registered` jest zapisywane w wewnętrznej bazie danych.

---
*   ID: US-002
*   Tytuł: Logowanie do aplikacji
*   Opis: Jako zarejestrowany użytkownik, chcę móc zalogować się do aplikacji przy użyciu mojego adresu e-mail i hasła, aby uzyskać dostęp do moich feedów i podsumowań.
*   Kryteria akceptacji:
    1.  Formularz logowania zawiera pola na adres e-mail i hasło.
    2.  Po poprawnym uwierzytelnieniu użytkownik jest przekierowany do głównego widoku aplikacji.
    3.  W przypadku podania błędnego e-maila lub hasła, wyświetlany jest jeden, ogólny komunikat o błędzie ("Nieprawidłowy e-mail lub hasło").
    4.  Zdarzenie `user_login` jest zapisywane w wewnętrznej bazie danych.

---
*   ID: US-003
*   Tytuł: Wylogowywanie z aplikacji
*   Opis: Jako zalogowany użytkownik, chcę móc się wylogować z aplikacji, aby zabezpieczyć swoje konto na współdzielonym urządzeniu.
*   Kryteria akceptacji:
    1.  W interfejsie aplikacji znajduje się przycisk lub link "Wyloguj".
    2.  Po kliknięciu przycisku sesja użytkownika jest kończona.
    3.  Użytkownik jest przekierowywany do strony logowania.

### Zarządzanie feedami RSS

---
*   ID: US-004
*   Tytuł: Widok początkowy dla nowego użytkownika
*   Opis: Jako nowy użytkownik, który zalogował się po raz pierwszy i nie ma żadnych feedów, chcę zobaczyć jasną instrukcję, jak dodać swoje pierwsze źródło, aby szybko rozpocząć korzystanie z aplikacji.
*   Kryteria akceptacji:
    1.  Gdy lista feedów użytkownika jest pusta, na ekranie głównym wyświetlany jest specjalny widok ("empty state").
    2.  Widok ten zawiera czytelny tekst zachęcający do dodania pierwszego feedu.
    3.  Widok zawiera wyraźny przycisk lub link "Dodaj feed", który otwiera formularz dodawania nowego źródła.

---
*   ID: US-005
*   Tytuł: Dodawanie nowego feedu RSS
*   Opis: Jako użytkownik, chcę móc dodać nowy feed RSS, podając jego nazwę i adres URL, aby uwzględnić go w moich przyszłych podsumowaniach.
*   Kryteria akceptacji:
    1.  Formularz dodawania feedu zawiera dwa pola tekstowe: "Nazwa" i "Adres URL".
    2.  Po przesłaniu formularza, serwer próbuje zweryfikować poprawność podanego adresu URL feedu.
    3.  Jeśli URL jest poprawny i feed został pomyślnie dodany, użytkownik widzi nowy feed na swojej liście, a formularz jest zamykany/czyszczony.
    4.  Jeśli podany URL jest nieprawidłowy lub serwer nie może pobrać z niego danych, użytkownik widzi komunikat o błędzie przy odpowiednim polu formularza.
    5.  Zdarzenie `feed_added` jest zapisywane w wewnętrznej bazie danych po pomyślnym dodaniu.

---
*   ID: US-006
*   Tytuł: Wyświetlanie listy feedów i ich statusu
*   Opis: Jako użytkownik, chcę widzieć listę wszystkich moich dodanych feedów RSS wraz z informacją o ich statusie, abym mógł zarządzać moimi źródłami i wiedzieć, czy działają poprawnie.
*   Kryteria akceptacji:
    1.  Wszystkie dodane przez użytkownika feedy są wyświetlane w formie listy.
    2.  Każdy element listy pokazuje nazwę feedu.
    3.  Jeśli system napotkał problem podczas ostatniej próby pobrania danych z feedu (np. błąd 404, błąd serwera), obok nazwy feedu wyświetlana jest ikona błędu.
    4.  Po najechaniu kursorem myszy na ikonę błędu, wyświetla się dymek (tooltip) z krótkim, zrozumiałym opisem problemu (np. "Nie udało się pobrać danych z tego adresu URL").

---
*   ID: US-007
*   Tytuł: Edycja istniejącego feedu RSS
*   Opis: Jako użytkownik, chcę móc edytować nazwę i adres URL istniejącego feedu, aby poprawić błędy lub zaktualizować źródło.
*   Kryteria akceptacji:
    1.  Każdy element na liście feedów ma opcję "Edytuj".
    2.  Kliknięcie "Edytuj" otwiera formularz wypełniony aktualnymi danymi (nazwą i adresem URL) danego feedu.
    3.  Użytkownik może zmienić dane i zapisać je.
    4.  Po zapisaniu, walidacja adresu URL jest wykonywana ponownie, tak jak przy dodawaniu nowego feedu.
    5.  Zaktualizowane dane są widoczne na liście feedów.

---
*   ID: US-008
*   Tytuł: Usuwanie feedu RSS
*   Opis: Jako użytkownik, chcę móc usunąć feed RSS z mojej listy, gdy przestaje mnie on interesować.
*   Kryteria akceptacji:
    1.  Każdy element na liście feedów ma opcję "Usuń".
    2.  Po kliknięciu "Usuń", system prosi o potwierdzenie operacji (np. za pomocą okna modalnego "Czy na pewno chcesz usunąć ten feed?").
    3.  Po potwierdzeniu, feed jest trwale usuwany z listy użytkownika i z systemu.

### Generowanie podsumowania

---
*   ID: US-009
*   Tytuł: Generowanie podsumowania na żądanie
*   Opis: Jako użytkownik, chcę mieć przycisk "Generuj podsumowanie", aby w dowolnym momencie otrzymać streszczenie artykułów z ostatniej doby.
*   Kryteria akceptacji:
    1.  W głównym widoku aplikacji znajduje się widoczny przycisk "Generuj podsumowanie".
    2.  Przycisk jest aktywny tylko wtedy, gdy użytkownik ma dodany co najmniej jeden poprawnie działający feed.
    3.  Kliknięcie przycisku rozpoczyna proces generowania podsumowania AI z artykułów opublikowanych w ciągu ostatnich 24 godzin.
    4.  Podczas generowania podsumowania, interfejs informuje użytkownika o trwającym procesie (np. poprzez animację ładowania).
    5.  Zdarzenie `summary_generated` jest zapisywane w wewnętrznej bazie danych po pomyślnym wygenerowaniu.

---
*   ID: US-010
*   Tytuł: Wyświetlanie wygenerowanego podsumowania
*   Opis: Jako użytkownik, chcę widzieć ostatnio wygenerowane podsumowanie wraz z datą jego utworzenia, abym mógł je przeczytać i wiedzieć, jak aktualne są zawarte w nim informacje.
*   Kryteria akceptacji:
    1.  W głównym widoku aplikacji znajduje się dedykowany obszar do wyświetlania podsumowania.
    2.  W tym obszarze wyświetlany jest tekst ostatnio wygenerowanego podsumowania.
    3.  Nad lub pod tekstem podsumowania widoczna jest data i godzina jego wygenerowania (np. "Wygenerowano: 7.10.2025, 15:30").
    4.  Podsumowanie pozostaje widoczne do momentu, aż użytkownik wygeneruje nowe.

## 6. Metryki sukcesu
Sukces wersji MVP produktu VibeFeeder będzie mierzony za pomocą następujących kluczowych wskaźników wydajności (KPI):

1.  Adopcja podstawowej funkcji (User Adoption):
    *   Metryka: Procent aktywnych użytkowników, którzy posiadają co najmniej jeden dodany i aktywny feed RSS.
    *   Cel: 90%
    *   Opis: Mierzy, czy użytkownicy rozumieją i wykonują podstawową akcję wymaganą do korzystania z produktu. "Aktywny użytkownik" jest definiowany jako osoba, która zalogowała się co najmniej raz w ciągu ostatnich 30 dni.

2.  Zaangażowanie użytkowników (User Engagement):
    *   Metryka: Procent aktywnych użytkowników, którzy generują co najmniej jedno podsumowanie w ciągu tygodnia.
    *   Cel: 75%
    *   Opis: Mierzy, jak często użytkownicy wracają do aplikacji, aby skorzystać z jej głównej propozycji wartości, czyli generowania podsumowań.

Sposób pomiaru:
Oba wskaźniki będą mierzone na podstawie analizy zdarzeń (`user_login`, `feed_added`, `summary_generated`) zapisywanych i agregowanych w wewnętrznej bazie danych aplikacji. Nie będą wykorzystywane zewnętrzne narzędzia analityczne.
