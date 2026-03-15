package configui

import "testing"

func TestValidateSettingsRequiresAPIKeyForOpenRouter(t *testing.T) {
	err := ValidateSettings(DistillSettings{
		ProviderName: "openrouter",
		BaseURL:      "https://openrouter.ai/api/v1",
		APIKey:       "",
	})
	if err == nil {
		t.Fatalf("expected validation error when api key is missing")
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
		BaseURL:      "  http://127.0.0.1:9000/v1 ",
		APIKey:       " token ",
	})

	if settings.ProviderName != "openai-compatible" {
		t.Fatalf("unexpected provider name: %q", settings.ProviderName)
	}
	if settings.BaseURL != "http://127.0.0.1:9000/v1" {
		t.Fatalf("unexpected base url: %q", settings.BaseURL)
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
