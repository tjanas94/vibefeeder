# Plan Wdrożenia Usługi OpenRouter

## 1. Opis Usługi

Usługa OpenRouter (`OpenRouterService`) będzie odpowiedzialna za komunikację z API OpenRouter w celu generowania uzupełnień czatów opartych na modelach LLM. Usługa ta będzie stanowić abstrakcję nad API, upraszczając proces wysyłania żądań i obsługi odpowiedzi. Zostanie zintegrowana z istniejącą architekturą aplikacji, wykorzystując pakiety `config` do zarządzania kluczami API, `logger` do logowania oraz `errors` do spójnej obsługi błędów.

Implementacja znajdzie się w pliku `internal/shared/ai/openrouter.go`.

## 2. Opis Konstruktora

Konstruktor `NewOpenRouterService` będzie tworzył nową instancję `OpenRouterService`.

```go
// NewOpenRouterService tworzy nową instancję usługi OpenRouter.
// Wymaga dostarczenia konfiguracji OpenRouter oraz klienta HTTP.
func NewOpenRouterService(cfg config.OpenRouterConfig, httpClient *http.Client) (*OpenRouterService, error)
```

**Argumenty:**

- `cfg config.OpenRouterConfig`: Konfiguracja OpenRouter, z której pobrany zostanie klucz API OpenRouter (`cfg.APIKey`).
- `httpClient *http.Client`: Klient HTTP do wykonywania żądań. Pozwala to na współdzielenie klienta i jego konfiguracji (np. timeoutów) w całej aplikacji.

**Logika:**

1. Inicjalizuje i zwraca nową instancję `OpenRouterService` z podanymi zależnościami.

## 3. Publiczne Metody i Pola

### `GenerateChatCompletion`

Główna metoda publiczna usługi, która wysyła żądanie do API OpenRouter i zwraca odpowiedź.

```go
// GenerateChatCompletion wysyła żądanie uzupełnienia czatu do API OpenRouter.
// Zwraca sparsowaną odpowiedź lub błąd.
func (s *OpenRouterService) GenerateChatCompletion(ctx context.Context, options GenerateChatCompletionOptions) (*ChatCompletionResponse, error)
```

**Argumenty:**

- `ctx context.Context`: Kontekst żądania, umożliwiający anulowanie operacji.
- `options GenerateChatCompletionOptions`: Struktura zawierająca wszystkie parametry żądania.

### Struktury Danych

Poniższe struktury będą publiczne, aby umożliwić klientom usługi budowanie żądań i odczytywanie odpowiedzi.

```go
// GenerateChatCompletionOptions definiuje parametry dla żądania uzupełnienia czatu.
type GenerateChatCompletionOptions struct {
    Model           string
    SystemPrompt    string
    UserPrompt      string
    ResponseFormat  *ResponseFormat
    Temperature     float64
    MaxTokens       int
}

// ResponseFormat definiuje format odpowiedzi, np. schemat JSON.
type ResponseFormat struct {
    Type       string      `json:"type"`
    JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema definiuje szczegóły schematu JSON.
type JSONSchema struct {
    Name   string      `json:"name"`
    Strict bool        `json:"strict"`
    Schema interface{} `json:"schema"` // Może być map[string]interface{} lub struct
}

// ChatCompletionResponse reprezentuje pełną odpowiedź z API.
type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Choices []Choice `json:"choices"`
    // ... inne pola odpowiedzi
}

// Choice reprezentuje pojedynczy wybór/odpowiedź w odpowiedzi API.
type Choice struct {
    Message ChatMessage `json:"message"`
}

// ChatMessage reprezentuje pojedynczą wiadomość w konwersacji.
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

## 4. Prywatne Metody i Pola

### Pola

```go
type OpenRouterService struct {
    apiKey     string
    httpClient *http.Client
    baseURL    string
}
```

- `apiKey`: Klucz API do autoryzacji żądań.
- `httpClient`: Klient HTTP używany do wysyłania żądań.
- `baseURL`: Podstawowy adres URL API OpenRouter (`https://openrouter.ai/api/v1`).

### Metody

- `buildRequest(ctx context.Context, options GenerateChatCompletionOptions) (*http.Request, error)`: Prywatna metoda do budowania obiektu `http.Request` na podstawie `GenerateChatCompletionOptions`. Będzie ona tworzyć ciało JSON, ustawiać nagłówki (`Authorization`, `Content-Type`) i dołączać kontekst.
- `parseResponse(resp *http.Response) (*ChatCompletionResponse, error)`: Metoda do parsowania odpowiedzi `http.Response` do struktury `ChatCompletionResponse`. Będzie również obsługiwać błędy HTTP.

## 5. Obsługa Błędów

Usługa będzie implementować spójną obsługę błędów, opakowując błędy z niższych warstw (np. `http`, `json`) i zwracając je z odpowiednim kontekstem.

- **Błędy walidacji**: `GenerateChatCompletion` zwróci błąd, jeśli `UserPrompt` jest pusty.
- **Błędy HTTP**:
  - `401 Unauthorized`: Zwrócony zostanie błąd wskazujący na problem z kluczem API.
  - `429 Too Many Requests`: Błąd zostanie zalogowany, a usługa może w przyszłości zaimplementować logikę ponawiania.
  - `5xx Server Error`: Podobnie jak wyżej, błąd zostanie zalogowany.
- **Błędy parsowania**: Jeśli odpowiedź API nie będzie mogła zostać sparsowana, surowa odpowiedź zostanie zalogowana, a usługa zwróci błąd.
- **Anulowanie kontekstu**: Jeśli `ctx` zostanie anulowany, żądanie HTTP zostanie przerwane, a metoda zwróci `context.Canceled`.

## 6. Kwestie Bezpieczeństwa

1.  **Klucz API**: Klucz API będzie wczytywany z konfiguracji (`.env` lub zmienne środowiskowe) i nigdy nie będzie hardkodowany w kodzie.
2.  **Walidacja Danych Wejściowych**: `UserPrompt` powinien być traktowany jako niezaufane dane wejściowe. Chociaż w tym przypadku jest on przekazywany do zewnętrznego API, należy unikać jego bezpośredniego użycia w logice aplikacji bez odpowiedniej walidacji lub sanitacji, jeśli byłby używany w innym kontekście.
3.  **Logowanie**: Należy unikać logowania pełnych treści promptów użytkownika, jeśli zawierają one dane wrażliwe. Logowanie metadanych i błędów jest preferowane.

## 7. Plan Wdrożenia Krok po Kroku

1.  **Krok 1: Zdefiniowanie Struktur Danych w `internal/shared/ai/openrouter.go`**
    - Zdefiniuj wszystkie publiczne i prywatne struktury (`OpenRouterService`, `GenerateChatCompletionOptions`, `ResponseFormat`, `JSONSchema`, `ChatCompletionResponse` itd.) zgodnie z sekcją 3 i 4.
    - Zdefiniuj strukturę dla ciała żądania (`chatCompletionRequest`), która będzie zawierać wszystkie pola, takie jak `model`, `messages`, `response_format`.

2.  **Krok 2: Implementacja Konstruktora `NewOpenRouterService`**
    - W `internal/shared/ai/openrouter.go` zaimplementuj funkcję `NewOpenRouterService`.
    - Zainicjalizuj `baseURL` na `https://openrouter.ai/api/v1`.

3.  **Krok 3: Implementacja Metody `buildRequest`**
    - Stwórz prywatną metodę `buildRequest`.
    - Zbuduj tablicę `messages`, dodając `SystemPrompt` (jeśli istnieje) i `UserPrompt`.
    - Stwórz instancję `chatCompletionRequest` i wypełnij ją danymi z `GenerateChatCompletionOptions`.
    - **Obsługa `response_format`**:
      - Jeśli `options.ResponseFormat` jest zdefiniowany i `Schema` jest strukturą Go, użyj `json.Marshal`, a następnie `json.Unmarshal` do `map[string]interface{}`, aby przekonwertować schemat Go na schemat JSON, który może być serializowany w ciele żądania.
      - **Przykład implementacji `response_format`**:

        ```go
        // Wewnątrz buildRequest
        var schemaPayload interface{}
        if options.ResponseFormat != nil && options.ResponseFormat.JSONSchema != nil && options.ResponseFormat.JSONSchema.Schema != nil {
            // Konwertuj schemat Go na mapę
            schemaBytes, err := json.Marshal(options.ResponseFormat.JSONSchema.Schema)
            if err != nil {
                return nil, fmt.Errorf("failed to marshal json schema: %w", err)
            }
            if err := json.Unmarshal(schemaBytes, &schemaPayload); err != nil {
                return nil, fmt.Errorf("failed to unmarshal json schema to map: %w", err)
            }
        }

        // ...

        requestBody := chatCompletionRequest{
            // ...
            ResponseFormat: &ResponseFormat{
                Type: "json_schema",
                JSONSchema: &JSONSchema{
                    Name:   options.ResponseFormat.JSONSchema.Name,
                    Strict: options.ResponseFormat.JSONSchema.Strict,
                    Schema: schemaPayload,
                },
            },
        }
        ```

    - Serializuj `requestBody` do formatu JSON.
    - Użyj `http.NewRequestWithContext` do stworzenia nowego żądania `POST`.
    - Ustaw nagłówki: `Authorization: Bearer <apiKey>` i `Content-Type: application/json`.

4.  **Krok 4: Implementacja Głównej Metody `GenerateChatCompletion`**
    - Wywołaj `buildRequest`, aby stworzyć żądanie.
    - Użyj `s.httpClient.Do(req)` do wysłania żądania.
    - Obsłuż błędy sieciowe i anulowanie kontekstu.
    - Sprawdź kod statusu odpowiedzi. Jeśli nie jest to `200 OK`, odczytaj ciało błędu, zaloguj je i zwróć odpowiedni błąd.
    - Jeśli odpowiedź jest pomyślna, wywołaj `parseResponse`.

5.  **Krok 5: Implementacja Metody `parseResponse`**
    - Odczytaj ciało odpowiedzi.
    - Użyj `json.NewDecoder(resp.Body).Decode()` do sparsowania odpowiedzi do struktury `ChatCompletionResponse`.
    - Zwróć sparsowaną odpowiedź lub błąd parsowania.

6.  **Krok 6: Integracja z Aplikacją**
    - W głównym pliku aplikacji (np. `cmd/vibefeeder/main.go` lub w miejscu, gdzie tworzone są serwisy) dodaj tworzenie instancji `OpenRouterService`.
    - Przekaż `cfg.OpenRouter` (konfigurację OpenRouter) i klienta HTTP do konstruktora.
    - Wstrzyknij `OpenRouterService` jako zależność do serwisów, które będą go używać (np. `SummaryService`).

7.  **Krok 7: Walidacja Konfiguracji**
    - W pliku `internal/shared/config/config.go` dodaj walidację klucza API OpenRouter w metodzie `validate()`.
    - Sprawdź, czy `c.OpenRouter.APIKey` nie jest pusty i zwróć błąd, jeśli nie jest ustawiony.
    - W pliku `.env.example` upewnij się, że zmienna `OPENROUTER_API_KEY=""` jest dodana jako przykład.
