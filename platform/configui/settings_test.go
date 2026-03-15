package configui

import "testing"

func TestValidateSettingsRequiresAPIKeyForOpenRouter(t *testing.T) {
	err := ValidateSettings(DistillSettings{
		ProviderName: "openrouter",
		Model:        "openai/gpt-4o-mini",
		BaseURL:      "https://openrouter.ai/api/v1",
		APIKey:       "",
	})
	if err == nil {
		t.Fatalf("expected validation error when api key is missing")
	}
}

func TestValidateSettingsRequiresModelForOpenRouter(t *testing.T) {
	err := ValidateSettings(DistillSettings{
		ProviderName: "openrouter",
		BaseURL:      "https://openrouter.ai/api/v1",
		APIKey:       "token",
	})
	if err == nil {
		t.Fatalf("expected validation error when model is missing")
	}
}

func TestValidateSettingsAcceptsOllamaWithDefaultBaseURL(t *testing.T) {
	err := ValidateSettings(DistillSettings{
		ProviderName: "ollama",
	})
	if err != nil {
		t.Fatalf("expected valid ollama settings, got: %v", err)
	}
}

func TestNormalizeSettingsCanonicalizesProviderAlias(t *testing.T) {
	settings := NormalizeSettings(DistillSettings{
		ProviderName: "OpenAI Compatible",
		Model:        "  gpt-4.1-mini ",
		BaseURL:      "  http://127.0.0.1:9000/v1 ",
		APIKey:       " token ",
	})

	if settings.ProviderName != "openai-compatible" {
		t.Fatalf("unexpected provider name: %q", settings.ProviderName)
	}
	if settings.BaseURL != "http://127.0.0.1:9000/v1" {
		t.Fatalf("unexpected base url: %q", settings.BaseURL)
	}
	if settings.Model != "gpt-4.1-mini" {
		t.Fatalf("unexpected model: %q", settings.Model)
	}
	if settings.APIKey != "token" {
		t.Fatalf("unexpected api key: %q", settings.APIKey)
	}
}

func TestNormalizeSettingsAddsOpenRouterDefaultBaseURL(t *testing.T) {
	settings := NormalizeSettings(DistillSettings{
		ProviderName: "openrouter",
		APIKey:       "token",
	})

	if settings.BaseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("unexpected base url: %q", settings.BaseURL)
	}
}

func TestNormalizeSettingsAddsOllamaDefaultModel(t *testing.T) {
	settings := NormalizeSettings(DistillSettings{
		ProviderName: "ollama",
	})

	if settings.Model != "qwen3.5:2b" {
		t.Fatalf("unexpected model: %q", settings.Model)
	}
}
