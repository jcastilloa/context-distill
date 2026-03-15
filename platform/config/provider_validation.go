package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func validateDistillProviderConfig(v *viper.Viper) error {
	provider := normalizeProviderName(v.GetString("distill.provider_name"))
	if provider == "" {
		provider = defaultDistillProvider
	}

	if !isSupportedDistillProvider(provider) {
		return fmt.Errorf("unsupported distill provider: %s", provider)
	}

	if !isOpenAICompatibleProvider(provider) {
		return nil
	}

	baseURL := firstNonEmpty(
		v.GetString("distill.base_url"),
		v.GetString("openai.base_url"),
		v.GetString("openai.api_base"),
		providerDefaultBaseURL(provider),
	)
	if baseURL == "" {
		return fmt.Errorf("distill provider %q requires base_url (set distill.base_url or openai.base_url)", provider)
	}

	if providerRequiresAPIKey(provider) {
		apiKey := firstNonEmpty(v.GetString("distill.api_key"), v.GetString("openai.api_key"))
		if apiKey == "" {
			return fmt.Errorf("distill provider %q requires api_key (set distill.api_key or openai.api_key)", provider)
		}
	}

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}
