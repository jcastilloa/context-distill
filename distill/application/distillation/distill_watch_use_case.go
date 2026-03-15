package distillation

import (
	"context"
	"strings"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
)

type DistillWatchUseCase struct {
	promptBuilder  distilldomain.PromptBuilder
	textPolicy     distilldomain.TextPolicy
	summarizer     distilldomain.SummarizerRepository
	changeDetector WatchChangeDetector
}

func NewDistillWatchUseCase(
	promptBuilder distilldomain.PromptBuilder,
	textPolicy distilldomain.TextPolicy,
	summarizer distilldomain.SummarizerRepository,
) *DistillWatchUseCase {
	return &DistillWatchUseCase{
		promptBuilder:  promptBuilder,
		textPolicy:     textPolicy,
		summarizer:     summarizer,
		changeDetector: NewWatchChangeDetector(),
	}
}

func (u *DistillWatchUseCase) Execute(ctx context.Context, request DistillWatchRequest) (DistillWatchResult, error) {
	if strings.TrimSpace(request.Question) == "" {
		return DistillWatchResult{}, ErrQuestionRequired
	}
	if strings.TrimSpace(request.PreviousCycle) == "" {
		return DistillWatchResult{}, ErrPreviousCycleRequired
	}
	if strings.TrimSpace(request.CurrentCycle) == "" {
		return DistillWatchResult{}, ErrCurrentCycleRequired
	}

	normalizedPrevious := u.textPolicy.NormalizeForModel(request.PreviousCycle)
	normalizedCurrent := u.textPolicy.NormalizeForModel(request.CurrentCycle)
	if !u.changeDetector.HasRelevantChange(normalizedPrevious, normalizedCurrent) {
		return DistillWatchResult{Output: u.textPolicy.EnsureTrailingNewline("No relevant change."), UsedFallback: false}, nil
	}

	prompt := u.promptBuilder.BuildWatchPrompt(strings.TrimSpace(request.Question), normalizedPrevious, normalizedCurrent)
	summary, err := u.summarizer.SummarizeWatch(ensureContext(ctx), prompt)
	if err != nil {
		return DistillWatchResult{Output: request.CurrentCycle, UsedFallback: true}, nil
	}

	summary = strings.TrimSpace(summary)
	if u.textPolicy.LooksLikeBadDistillation(request.CurrentCycle, summary) {
		return DistillWatchResult{Output: request.CurrentCycle, UsedFallback: true}, nil
	}

	return DistillWatchResult{
		Output:       u.textPolicy.EnsureTrailingNewline(summary),
		UsedFallback: false,
	}, nil
}
