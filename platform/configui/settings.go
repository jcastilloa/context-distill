package configui

import (
	"fmt"
	"strings"

	configrepo "github.com/jcastilloa/context-distill/platform/config"
)

func NormalizeSettings(settings DistillSettings) DistillSettings {
	normalized := DistillSettings{
		ProviderName: configrepo.NormalizeProviderName(settings.ProviderName),
		BaseURL:      strings.TrimSpace(settings.BaseURL),
		APIKey:       strings.TrimSpace(settings.APIKey),
	}
	if normalized.ProviderName == "" {
		normalized.ProviderName = "ollama"
	}
	if normalized.BaseURL == "" {
		normalized.BaseURL = configrepo.ProviderDefaultBaseURL(normalized.ProviderName)
	}

	return normalized
}

func ValidateSettings(settings DistillSettings) error {
	normalized := NormalizeSettings(settings)

	if !configrepo.IsSupportedDistillProvider(normalized.ProviderName) {
		return fmt.Errorf("unsupported provider: %s", normalized.ProviderName)
	}

	if configrepo.IsOpenAICompatibleProvider(normalized.ProviderName) && normalized.BaseURL == "" {
		return fmt.Errorf("base URL is required for provider %q", normalized.ProviderName)
	}

	if configrepo.ProviderRequiresAPIKey(normalized.ProviderName) && normalized.APIKey == "" {
		return fmt.Errorf("API key is required for provider %q", normalized.ProviderName)
	}

	return nil
}
