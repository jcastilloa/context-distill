package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

func TestDistillSummarizerRepositorySummarizeBatch(t *testing.T) {
	var receivedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(`{"response":"  concise  "}`))
	}))
	defer server.Close()

	repo := NewDistillSummarizerRepository(aiDomain.ProviderConfig{
		BaseURL:  server.URL,
		Model:    "qwen3.5:2b",
		Timeout:  5 * time.Second,
		Thinking: true,
	}, nil)

	output, err := repo.SummarizeBatch(context.Background(), "prompt text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "concise" {
		t.Fatalf("unexpected output: %q", output)
	}
	if receivedBody["model"] != "qwen3.5:2b" {
		t.Fatalf("unexpected model: %v", receivedBody["model"])
	}
	if receivedBody["prompt"] != "prompt text" {
		t.Fatalf("unexpected prompt: %v", receivedBody["prompt"])
	}
	if receivedBody["stream"] != false {
		t.Fatalf("expected stream=false, got %v", receivedBody["stream"])
	}
	if receivedBody["think"] != true {
		t.Fatalf("expected think=true, got %v", receivedBody["think"])
	}
}

func TestDistillSummarizerRepositorySummarizeWatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":"watch delta"}`))
	}))
	defer server.Close()

	repo := NewDistillSummarizerRepository(aiDomain.ProviderConfig{BaseURL: server.URL, Timeout: 5 * time.Second}, nil)
	output, err := repo.SummarizeWatch(context.Background(), "prompt text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "watch delta" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestDistillSummarizerRepositoryReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	repo := NewDistillSummarizerRepository(aiDomain.ProviderConfig{BaseURL: server.URL, Timeout: 5 * time.Second}, nil)
	_, err := repo.SummarizeBatch(context.Background(), "prompt text")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "500") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestDistillSummarizerRepositoryReturnsErrorOnInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("no-json"))
	}))
	defer server.Close()

	repo := NewDistillSummarizerRepository(aiDomain.ProviderConfig{BaseURL: server.URL, Timeout: 5 * time.Second}, nil)
	_, err := repo.SummarizeBatch(context.Background(), "prompt text")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "json") {
		t.Fatalf("expected JSON error, got %v", err)
	}
}
