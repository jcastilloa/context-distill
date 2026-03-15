package di

import (
	"fmt"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
	"github.com/jcastilloa/context-distill/platform/ollama"
	"github.com/jcastilloa/context-distill/platform/openai"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

func newDistillSummarizerRepository(cfg aiDomain.ProviderConfig, aiRepository aiDomain.AIRepository) (distilldomain.SummarizerRepository, error) {
	provider := NormalizeProviderName(cfg.ProviderName)
	if provider == "" {
		provider = "ollama"
	}

	if provider == "ollama" {
		return ollama.NewDistillSummarizerRepository(cfg, nil), nil
	}

	if IsOpenAICompatibleProvider(provider) {
		return openai.NewDistillSummarizerRepository(aiRepository), nil
	}

	return nil, fmt.Errorf("unsupported distill provider: %s", provider)
}
