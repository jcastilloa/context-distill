package config

import "testing"

func TestProviderDefaultBaseURLForOpenRouter(t *testing.T) {
	baseURL := ProviderDefaultBaseURL("openrouter")
	if baseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("expected openrouter default URL, got %q", baseURL)
	}
}
