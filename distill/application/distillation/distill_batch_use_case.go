package distillation

import (
	"context"
	"strings"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
)

type DistillBatchUseCase struct {
	promptBuilder distilldomain.PromptBuilder
	textPolicy    distilldomain.TextPolicy
	summarizer    distilldomain.SummarizerRepository
}

func NewDistillBatchUseCase(
	promptBuilder distilldomain.PromptBuilder,
	textPolicy distilldomain.TextPolicy,
	summarizer distilldomain.SummarizerRepository,
) *DistillBatchUseCase {
	return &DistillBatchUseCase{
		promptBuilder: promptBuilder,
		textPolicy:    textPolicy,
		summarizer:    summarizer,
	}
}

func (u *DistillBatchUseCase) Execute(ctx context.Context, request DistillBatchRequest) (DistillBatchResult, error) {
	if strings.TrimSpace(request.Question) == "" {
		return DistillBatchResult{}, ErrQuestionRequired
	}
	if strings.TrimSpace(request.Input) == "" {
		return DistillBatchResult{}, ErrInputRequired
	}

	normalizedInput := u.textPolicy.NormalizeForModel(request.Input)
	prompt := u.promptBuilder.BuildBatchPrompt(strings.TrimSpace(request.Question), normalizedInput)
	summary, err := u.summarizer.SummarizeBatch(ensureContext(ctx), prompt)
	if err != nil {
		return DistillBatchResult{Output: request.Input, UsedFallback: true}, nil
	}

	summary = strings.TrimSpace(summary)
	if u.textPolicy.LooksLikeBadDistillation(request.Input, summary) {
		return DistillBatchResult{Output: request.Input, UsedFallback: true}, nil
	}

	return DistillBatchResult{
		Output:       u.textPolicy.EnsureTrailingNewline(summary),
		UsedFallback: false,
	}, nil
}
