package main

import (
	"strings"

	"github.com/jcastilloa/context-distill/platform/di"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

func buildDistillAIProviderConfig(openaiCfg, distillCfg aiDomain.ProviderConfig) aiDomain.ProviderConfig {
	normalizedProvider := di.NormalizeProviderName(distillCfg.ProviderName)
	if !di.IsOpenAICompatibleProvider(normalizedProvider) {
		return openaiCfg
	}

	merged := openaiCfg
	merged.ProviderName = normalizedProvider

	overrideString(&merged.APIKey, distillCfg.APIKey)
	overrideString(&merged.BaseURL, distillCfg.BaseURL)
	overrideString(&merged.Model, distillCfg.Model)

	if distillCfg.Timeout > 0 {
		merged.Timeout = distillCfg.Timeout
	}
	if distillCfg.MaxRetries > 0 {
		merged.MaxRetries = distillCfg.MaxRetries
	}

	return merged
}

func overrideString(target *string, candidate string) {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return
	}
	*target = trimmed
}
