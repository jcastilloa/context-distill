package config

import (
	"strings"
	"unicode"
)

const (
	defaultOpenAIBaseURL            = "https://api.openai.com/v1"
	defaultOpenRouterBaseURL        = "https://openrouter.ai/api/v1"
	defaultLMStudioBaseURL          = "http://127.0.0.1:1234/v1"
	defaultJanBaseURL               = "http://127.0.0.1:1337/v1"
	defaultLocalAIBaseURL           = "http://127.0.0.1:8080/v1"
	defaultVLLMBaseURL              = "http://127.0.0.1:8000/v1"
	defaultDockerModelRunnerBaseURL = "http://127.0.0.1:12434/engines/v1"
)

var configProviderAliases = map[string]string{
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

func normalizeProviderName(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return ""
	}

	aliasKey := providerAliasKey(normalized)
	if canonical, ok := configProviderAliases[aliasKey]; ok {
		return canonical
	}

	return normalized
}

func isOpenAICompatibleProvider(provider string) bool {
	switch normalizeProviderName(provider) {
	case "openai", "openai-compatible", "openrouter", "lmstudio", "jan", "localai", "vllm", "sglang", "llama.cpp", "mlx-lm", "docker-model-runner":
		return true
	default:
		return false
	}
}

func isSupportedDistillProvider(provider string) bool {
	normalized := normalizeProviderName(provider)
	return normalized == "ollama" || isOpenAICompatibleProvider(normalized)
}

func providerDefaultBaseURL(provider string) string {
	switch normalizeProviderName(provider) {
	case "ollama":
		return defaultDistillBaseURL
	case "openai":
		return defaultOpenAIBaseURL
	case "openrouter":
		return defaultOpenRouterBaseURL
	case "lmstudio":
		return defaultLMStudioBaseURL
	case "jan":
		return defaultJanBaseURL
	case "localai":
		return defaultLocalAIBaseURL
	case "vllm":
		return defaultVLLMBaseURL
	case "docker-model-runner":
		return defaultDockerModelRunnerBaseURL
	default:
		return ""
	}
}

func providerRequiresAPIKey(provider string) bool {
	switch normalizeProviderName(provider) {
	case "openai", "openrouter", "jan":
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
