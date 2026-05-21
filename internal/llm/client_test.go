package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestClient_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Hello from LLM!",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	result, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if result != "Hello from LLM!" {
		t.Errorf("Send() = %q, want %q", result, "Hello from LLM!")
	}
}

func TestClient_Send_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "bad-key",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	_, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status 401: %v", err)
	}
}

func TestClient_Send_Retryable(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "success on retry"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	result, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if result != "success on retry" {
		t.Errorf("Send() = %q", result)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestClient_Send_RetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	_, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for exhausted retries")
	}
	if !strings.Contains(err.Error(), "retry exhausted") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Send_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	_, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestClient_Send_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	_, err := client.Send([]ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestClient_Stream_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunks := []string{
			`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"content":" World"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"content":"!"},"finish_reason":"stop"}]}`,
		}
		for _, c := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", c)
			flusher.Flush()
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	ch := make(chan StreamDelta, 256)
	go client.Stream([]ChatMessage{{Role: "user", Content: "hi"}}, ch)

	var contents []string
	done := false
	for d := range ch {
		if d.Err != nil {
			t.Fatalf("Stream error: %v", d.Err)
		}
		if d.Content != "" {
			contents = append(contents, d.Content)
		}
		if d.Done {
			done = true
		}
	}

	if !done {
		t.Error("stream did not send Done")
	}
	result := strings.Join(contents, "")
	if result != "Hello World!" {
		t.Errorf("stream content = %q, want %q", result, "Hello World!")
	}
}

func TestClient_Stream_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "bad-key",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	ch := make(chan StreamDelta, 256)
	go client.Stream([]ChatMessage{{Role: "user", Content: "hi"}}, ch)

	hasError := false
	for d := range ch {
		if d.Err != nil {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error for 401")
	}
}

func TestClient_Stream_RetryThenSuccess(t *testing.T) {
	var mu sync.Mutex
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		count := callCount
		mu.Unlock()

		if count == 1 {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"retry ok"},"finish_reason":"stop"}]}`)
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	ch := make(chan StreamDelta, 256)
	go client.Stream([]ChatMessage{{Role: "user", Content: "hi"}}, ch)

	var contents []string
	retrying := false
	for d := range ch {
		if d.Err != nil {
			t.Fatalf("Stream error: %v", d.Err)
		}
		if d.Retrying {
			retrying = true
		}
		if d.Content != "" && !d.Retrying {
			contents = append(contents, d.Content)
		}
	}

	mu.Lock()
	gotCount := callCount
	mu.Unlock()

	if gotCount != 2 {
		t.Errorf("expected 2 calls, got %d", gotCount)
	}
	if !retrying {
		t.Error("expected retrying signal")
	}
	result := strings.Join(contents, "")
	if result != "retry ok" {
		t.Errorf("stream content = %q", result)
	}
}

func TestClient_Stream_RetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := &Client{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "test-model",
	}

	ch := make(chan StreamDelta, 256)
	go client.Stream([]ChatMessage{{Role: "user", Content: "hi"}}, ch)

	hasError := false
	for d := range ch {
		if d.Err != nil {
			hasError = true
			if !strings.Contains(d.Err.Error(), "retry exhausted") {
				t.Errorf("unexpected error: %v", d.Err)
			}
		}
	}
	if !hasError {
		t.Error("expected retry exhausted error")
	}
}

func TestClient_ModelName(t *testing.T) {
	c := &Client{Model: "gpt-4"}
	if c.ModelName() != "gpt-4" {
		t.Errorf("ModelName() = %q", c.ModelName())
	}
}

func TestClient_ProviderName(t *testing.T) {
	c := &Client{Provider: "openai"}
	if c.ProviderName() != "openai" {
		t.Errorf("ProviderName() = %q", c.ProviderName())
	}
}
