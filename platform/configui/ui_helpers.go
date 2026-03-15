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
	if option.RequiresAPIKey && option.DefaultModel == "" {
		return fmt.Sprintf("[#facc15]Provider %s requires API key and explicit model.", option.Name)
	}
	if option.RequiresAPIKey {
		return fmt.Sprintf("[#facc15]Provider %s requires an API key.", option.Name)
	}
	if option.DefaultModel == "" {
		return fmt.Sprintf("[#facc15]Provider %s requires explicit model.", option.Name)
	}
	return fmt.Sprintf("[#86efac]Provider %s does not require an API key.", option.Name)
}
