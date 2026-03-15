package di

import (
	"strings"
	"unicode"
)

var providerAliases = map[string]string{
	"ollama":            "ollama",
	"openai":            "openai",
	"openaicompatible":  "openai-compatible",
	"openrouter":        "openrouter",
	"lmstudio":          "lmstudio",
	"jan":               "jan",
	"localai":           "localai",
	"vllm":              "vllm",
	"sglang":            "sglang",
	"llamacpp":          "llama.cpp",
	"mlxlm":             "mlx-lm",
	"dockermodelrunner": "docker-model-runner",
	"dmr":               "docker-model-runner",
	"modelrunner":       "docker-model-runner",
}

func NormalizeProviderName(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return ""
	}

	aliasKey := providerAliasKey(normalized)
	if canonical, ok := providerAliases[aliasKey]; ok {
		return canonical
	}

	return normalized
}

func IsOpenAICompatibleProvider(provider string) bool {
	switch NormalizeProviderName(provider) {
	case "openai", "openai-compatible", "openrouter", "lmstudio", "jan", "localai", "vllm", "sglang", "llama.cpp", "mlx-lm", "docker-model-runner":
		return true
	default:
		return false
	}
}

func providerAliasKey(provider string) string {
	var key strings.Builder
	key.Grow(len(provider))
	for _, r := range provider {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			key.WriteRune(unicode.ToLower(r))
		}
	}
	return key.String()
}
