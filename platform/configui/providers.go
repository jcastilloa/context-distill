package configui

import (
	"fmt"

	configrepo "github.com/jcastilloa/context-distill/platform/config"
)

var providerNames = []string{
	"ollama",
	"openrouter",
	"openai",
	"openai-compatible",
	"lmstudio",
	"jan",
	"localai",
	"vllm",
	"sglang",
	"llama.cpp",
	"mlx-lm",
	"docker-model-runner",
}

func ProviderOptions() []ProviderOption {
	options := make([]ProviderOption, 0, len(providerNames))
	for _, provider := range providerNames {
		options = append(options, ProviderOption{
			Name:           provider,
			Label:          providerLabel(provider),
			DefaultBaseURL: configrepo.ProviderDefaultBaseURL(provider),
			RequiresAPIKey: configrepo.ProviderRequiresAPIKey(provider),
		})
	}

	return options
}

func FindProviderOption(provider string) (ProviderOption, bool) {
	normalized := configrepo.NormalizeProviderName(provider)
	for _, option := range ProviderOptions() {
		if option.Name == normalized {
			return option, true
		}
	}

	return ProviderOption{}, false
}

func providerLabel(provider string) string {
	defaultBaseURL := configrepo.ProviderDefaultBaseURL(provider)
	if defaultBaseURL == "" {
		return provider
	}

	return fmt.Sprintf("%s  (%s)", provider, defaultBaseURL)
}
