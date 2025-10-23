Frontend - htmx + Alpine.js dla komponentów interaktywnych:

- htmx do obsługi dynamicznych interakcji i komunikacji z backendem bez konieczności przeładowywania strony
- Alpine.js do zarządzania stanem i logiką komponentów w prosty sposób
- Tailwind 4 pozwala na wygodne stylowanie aplikacji
- DaisyUI jako biblioteka komponentów UI oparta na Tailwindzie

Backend - Golang + Templ do generowania HTML po stronie serwera:

- Echo do obsługi routing'u i middleware
- Templ jako lekki silnik szablonów do generowania HTML
- Statycznie typowany język, co pomaga w unikaniu błędów w run-time
- Zapewnia wysoką wydajność i niskie zużycie zasobów
- Posiada bogatą bibliotekę standardową oraz wiele dostępnych bibliotek zewnętrznych
- Wspiera background jobs do okresowego pobierania feedów RSS i generowania podsumowań

Baza danych - Supabase jako kompleksowe rozwiązanie backendowe:

- Zapewnia bazę danych PostgreSQL
- Zapewnia SDK w wielu językach, które posłużą jako Backend-as-a-Service
- Jest rozwiązaniem open source, które można hostować lokalnie lub na własnym serwerze
- Posiada wbudowaną autentykację użytkowników

AI - Komunikacja z modelami przez usługę Openrouter.ai:

- Dostęp do szerokiej gamy modeli (OpenAI, Anthropic, Google i wiele innych), które pozwolą nam znaleźć rozwiązanie zapewniające wysoką efektywność i niskie koszta
- Pozwala na ustawianie limitów finansowych na klucze API

CI/CD i Hosting:

- Github Actions do tworzenia pipeline'ów CI/CD
- Fly.io jako platforma hostingowa z globalnym zasięgiem i prostą konfiguracją dla aplikacji Dockerowych
- Cloudflare jako reverse proxy z CDN i ochroną DDoS

Testowanie:

Testy jednostkowe:

- Standardowa biblioteka `testing` w Go jako podstawa wszystkich testów
- Testify (stretchr/testify) do asercji i tworzenia mocków

Testy E2E (End-to-End):

- Playwright jako framework do automatyzacji przeglądarki i testowania interakcji htmx
- Doskonałe wsparcie dla dynamicznych treści i oczekiwania na żądania sieciowe
- Testy uruchamiane na najnowszych wersjach Chrome
