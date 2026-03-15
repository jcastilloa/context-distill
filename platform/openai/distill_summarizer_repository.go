package openai

import (
	"context"
	"errors"
	"strings"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

var errEmptyProviderOutput = errors.New("provider returned empty output")

type DistillSummarizerRepository struct {
	aiRepository aiDomain.AIRepository
}

func NewDistillSummarizerRepository(aiRepository aiDomain.AIRepository) distilldomain.SummarizerRepository {
	return &DistillSummarizerRepository{aiRepository: aiRepository}
}

func (r *DistillSummarizerRepository) SummarizeBatch(ctx context.Context, prompt string) (string, error) {
	return r.summarize(ctx, prompt)
}

func (r *DistillSummarizerRepository) SummarizeWatch(ctx context.Context, prompt string) (string, error) {
	return r.summarize(ctx, prompt)
}

func (r *DistillSummarizerRepository) summarize(ctx context.Context, prompt string) (string, error) {
	response, err := r.aiRepository.GetAIResponse(ctx, aiDomain.NewRequest("", prompt))
	if err != nil {
		return "", err
	}
	if response == nil {
		return "", errEmptyProviderOutput
	}
	if response.HasError {
		if strings.TrimSpace(response.ErrorMsg) != "" {
			return "", errors.New(response.ErrorMsg)
		}
		return "", errEmptyProviderOutput
	}

	output := strings.TrimSpace(response.OutputText)
	if output == "" {
		return "", errEmptyProviderOutput
	}

	return output, nil
}
