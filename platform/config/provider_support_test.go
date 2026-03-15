package config

import "testing"

func TestProviderDefaultBaseURLForOpenRouter(t *testing.T) {
	baseURL := ProviderDefaultBaseURL("openrouter")
	if baseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("expected openrouter default URL, got %q", baseURL)
	}
}

func TestProviderDefaultModelForOllama(t *testing.T) {
	model := ProviderDefaultModel("ollama")
	if model != "qwen3.5:2b" {
		t.Fatalf("expected ollama default model, got %q", model)
	}
}
