package configui

import "fmt"

func providerIndex(options []ProviderOption, provider string) int {
	normalized := NormalizeSettings(DistillSettings{ProviderName: provider}).ProviderName
	for i, option := range options {
		if option.Name == normalized {
			return i
		}
	}

	return 0
}

func providerHelperText(option ProviderOption) string {
	if option.RequiresAPIKey {
		return fmt.Sprintf("[#facc15]Provider %s requires an API key.", option.Name)
	}
	return fmt.Sprintf("[#86efac]Provider %s does not require an API key.", option.Name)
}
