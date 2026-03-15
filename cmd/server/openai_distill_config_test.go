package main

import (
	"testing"
	"time"

	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

func TestBuildDistillAIProviderConfigUsesDistillOverridesForOpenAICompatibleProviders(t *testing.T) {
	openaiCfg := aiDomain.ProviderConfig{
		ProviderName:       "openai",
		APIKey:             "openai-key",
		BaseURL:            "https://api.openai.com/v1",
		Model:              "gpt-4o-mini",
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		SupportsSystemRole: true,
		SupportsJSONMode:   true,
	}

	distillCfg := aiDomain.ProviderConfig{
		ProviderName: "openai-compatible",
		APIKey:       "distill-key",
		BaseURL:      "https://openrouter.ai/api/v1",
		Model:        "meta-llama/llama-4-maverick",
		Timeout:      90 * time.Second,
		MaxRetries:   7,
	}

	cfg := buildDistillAIProviderConfig(openaiCfg, distillCfg)

	if cfg.ProviderName != "openai-compatible" {
		t.Fatalf("expected provider override, got %q", cfg.ProviderName)
	}
	if cfg.APIKey != "distill-key" {
		t.Fatalf("expected API key override, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("expected base URL override, got %q", cfg.BaseURL)
	}
	if cfg.Model != "meta-llama/llama-4-maverick" {
		t.Fatalf("expected model override, got %q", cfg.Model)
	}
	if cfg.Timeout != 90*time.Second {
		t.Fatalf("expected timeout override, got %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 7 {
		t.Fatalf("expected max retries override, got %d", cfg.MaxRetries)
	}
	if !cfg.SupportsSystemRole {
		t.Fatalf("supports_system_role should be preserved from openai config")
	}
	if !cfg.SupportsJSONMode {
		t.Fatalf("supports_json_mode should be preserved from openai config")
	}
}

func TestBuildDistillAIProviderConfigLeavesOpenAIConfigForNonOpenAICompatibleProvider(t *testing.T) {
	openaiCfg := aiDomain.ProviderConfig{
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
	}
	distillCfg := aiDomain.ProviderConfig{
		ProviderName: "ollama",
		Model:        "qwen3.5:2b",
	}

	cfg := buildDistillAIProviderConfig(openaiCfg, distillCfg)

	if cfg.ProviderName != "openai" {
		t.Fatalf("expected openai provider untouched, got %q", cfg.ProviderName)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Fatalf("expected openai model untouched, got %q", cfg.Model)
	}
}

func TestBuildDistillAIProviderConfigSupportsProviderAliasNormalization(t *testing.T) {
	openaiCfg := aiDomain.ProviderConfig{
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
	}
	distillCfg := aiDomain.ProviderConfig{
		ProviderName: "OpenAI Compatible",
		Model:        "alias-model",
	}

	cfg := buildDistillAIProviderConfig(openaiCfg, distillCfg)

	if cfg.ProviderName != "openai-compatible" {
		t.Fatalf("expected normalized provider name, got %q", cfg.ProviderName)
	}
	if cfg.Model != "alias-model" {
		t.Fatalf("expected distill model override, got %q", cfg.Model)
	}
}
