package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestDistillProviderConfigUsesDefaults(t *testing.T) {
	repo := &ViperRepository{v: viper.New()}
	cfg := repo.DistillProviderConfig()

	if cfg.ProviderName != "ollama" {
		t.Fatalf("unexpected default provider: %q", cfg.ProviderName)
	}
	if cfg.BaseURL != "http://127.0.0.1:11434" {
		t.Fatalf("unexpected default base URL: %q", cfg.BaseURL)
	}
	if cfg.Model != "qwen3.5:2b" {
		t.Fatalf("unexpected default model: %q", cfg.Model)
	}
	if cfg.Timeout != 90*time.Second {
		t.Fatalf("unexpected default timeout: %v", cfg.Timeout)
	}
	if cfg.Thinking {
		t.Fatalf("expected thinking disabled by default")
	}
}

func TestDistillProviderConfigReadsConfiguredValues(t *testing.T) {
	v := viper.New()
	v.Set("distill.provider_name", "openai-compatible")
	v.Set("distill.api_key", "key")
	v.Set("distill.base_url", "http://127.0.0.1:9000/v1")
	v.Set("distill.model", "model-x")
	v.Set("distill.timeout", "12s")
	v.Set("distill.max_retries", 4)
	v.Set("distill.supports_system_role", true)
	v.Set("distill.supports_json_mode", true)
	v.Set("distill.thinking", true)

	repo := &ViperRepository{v: v}
	cfg := repo.DistillProviderConfig()

	if cfg.ProviderName != "openai-compatible" {
		t.Fatalf("unexpected provider: %q", cfg.ProviderName)
	}
	if cfg.APIKey != "key" {
		t.Fatalf("unexpected API key: %q", cfg.APIKey)
	}
	if cfg.BaseURL != "http://127.0.0.1:9000/v1" {
		t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
	}
	if cfg.Model != "model-x" {
		t.Fatalf("unexpected model: %q", cfg.Model)
	}
	if cfg.Timeout != 12*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 4 {
		t.Fatalf("unexpected max retries: %d", cfg.MaxRetries)
	}
	if !cfg.SupportsSystemRole {
		t.Fatalf("expected supports system role true")
	}
	if !cfg.SupportsJSONMode {
		t.Fatalf("expected supports json mode true")
	}
	if !cfg.Thinking {
		t.Fatalf("expected thinking true")
	}
}

func TestDistillProviderConfigDoesNotForceOllamaModelForOpenAICompatibleProviders(t *testing.T) {
	v := viper.New()
	v.Set("distill.provider_name", "openrouter")
	v.Set("distill.api_key", "key")
	v.Set("distill.base_url", "https://openrouter.ai/api/v1")

	repo := &ViperRepository{v: v}
	cfg := repo.DistillProviderConfig()

	if cfg.ProviderName != "openrouter" {
		t.Fatalf("unexpected provider: %q", cfg.ProviderName)
	}
	if cfg.Model != "" {
		t.Fatalf("expected empty model for openai-compatible provider when distill.model is not configured, got %q", cfg.Model)
	}
}
