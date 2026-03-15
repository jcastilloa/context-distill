package config

func NormalizeProviderName(provider string) string {
	return normalizeProviderName(provider)
}

func IsOpenAICompatibleProvider(provider string) bool {
	return isOpenAICompatibleProvider(provider)
}

func IsSupportedDistillProvider(provider string) bool {
	return isSupportedDistillProvider(provider)
}

func ProviderDefaultBaseURL(provider string) string {
	return providerDefaultBaseURL(provider)
}

func ProviderRequiresAPIKey(provider string) bool {
	return providerRequiresAPIKey(provider)
}
