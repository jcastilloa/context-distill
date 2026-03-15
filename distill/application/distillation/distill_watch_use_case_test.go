package distillation

import (
	"context"
	"errors"
	"testing"
)

func TestDistillWatchUseCaseExecuteReturnsErrorWhenQuestionIsEmpty(t *testing.T) {
	useCase := NewDistillWatchUseCase(&fakePromptBuilder{}, &fakeTextPolicy{}, &fakeSummarizerRepository{})

	_, err := useCase.Execute(context.Background(), DistillWatchRequest{PreviousCycle: "a", CurrentCycle: "b"})
	if !errors.Is(err, ErrQuestionRequired) {
		t.Fatalf("expected ErrQuestionRequired, got %v", err)
	}
}

func TestDistillWatchUseCaseExecuteReturnsErrorWhenPreviousCycleIsEmpty(t *testing.T) {
	useCase := NewDistillWatchUseCase(&fakePromptBuilder{}, &fakeTextPolicy{}, &fakeSummarizerRepository{})

	_, err := useCase.Execute(context.Background(), DistillWatchRequest{Question: "what changed?", CurrentCycle: "b"})
	if !errors.Is(err, ErrPreviousCycleRequired) {
		t.Fatalf("expected ErrPreviousCycleRequired, got %v", err)
	}
}

func TestDistillWatchUseCaseExecuteReturnsErrorWhenCurrentCycleIsEmpty(t *testing.T) {
	useCase := NewDistillWatchUseCase(&fakePromptBuilder{}, &fakeTextPolicy{}, &fakeSummarizerRepository{})

	_, err := useCase.Execute(context.Background(), DistillWatchRequest{Question: "what changed?", PreviousCycle: "a"})
	if !errors.Is(err, ErrCurrentCycleRequired) {
		t.Fatalf("expected ErrCurrentCycleRequired, got %v", err)
	}
}

func TestDistillWatchUseCaseExecuteReturnsSummarizedOutput(t *testing.T) {
	builder := &fakePromptBuilder{watchPrompt: "watch-prompt"}
	policy := &fakeTextPolicy{normalizedByInput: map[string]string{"previous raw": "previous normalized", "current raw": "current normalized"}}
	repo := &fakeSummarizerRepository{watchSummary: "delta summary"}
	useCase := NewDistillWatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillWatchRequest{
		Question:      "what changed?",
		PreviousCycle: "previous raw",
		CurrentCycle:  "current raw",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "delta summary\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if result.UsedFallback {
		t.Fatalf("expected non-fallback output")
	}
	if builder.watchQuestion != "what changed?" {
		t.Fatalf("unexpected question passed to prompt builder: %q", builder.watchQuestion)
	}
	if builder.watchPrevious != "previous normalized" || builder.watchCurrent != "current normalized" {
		t.Fatalf("expected normalized cycles in prompt builder, got previous=%q current=%q", builder.watchPrevious, builder.watchCurrent)
	}
	if repo.watchPrompt != "watch-prompt" {
		t.Fatalf("unexpected prompt passed to repository: %q", repo.watchPrompt)
	}
}

func TestDistillWatchUseCaseExecuteReturnsNoRelevantChangeWithoutCallingSummarizer(t *testing.T) {
	builder := &fakePromptBuilder{watchPrompt: "watch-prompt"}
	policy := &fakeTextPolicy{normalizedValue: "same normalized cycle"}
	repo := &fakeSummarizerRepository{watchSummary: "should not be used"}
	useCase := NewDistillWatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillWatchRequest{
		Question:      "what changed?",
		PreviousCycle: "same raw cycle",
		CurrentCycle:  "same raw cycle",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "No relevant change.\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if result.UsedFallback {
		t.Fatalf("expected non-fallback output")
	}
	if repo.watchCalls != 0 {
		t.Fatalf("expected no summarizer calls, got %d", repo.watchCalls)
	}
}
func TestDistillWatchUseCaseExecuteFallsBackToCurrentCycleWhenSummarizerFails(t *testing.T) {
	builder := &fakePromptBuilder{watchPrompt: "watch-prompt"}
	policy := &fakeTextPolicy{normalizedByInput: map[string]string{"previous raw": "previous normalized", "current raw": "current normalized"}}
	repo := &fakeSummarizerRepository{watchErr: errors.New("boom")}
	useCase := NewDistillWatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillWatchRequest{
		Question:      "what changed?",
		PreviousCycle: "previous raw",
		CurrentCycle:  "current raw",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "current raw" {
		t.Fatalf("expected current cycle fallback, got %q", result.Output)
	}
	if !result.UsedFallback {
		t.Fatalf("expected fallback output")
	}
}

func TestDistillWatchUseCaseExecuteFallsBackWhenSummaryLooksBad(t *testing.T) {
	builder := &fakePromptBuilder{watchPrompt: "watch-prompt"}
	policy := &fakeTextPolicy{normalizedByInput: map[string]string{"previous raw": "previous normalized", "current raw": "current normalized"}, badDistillation: true}
	repo := &fakeSummarizerRepository{watchSummary: "bad summary"}
	useCase := NewDistillWatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillWatchRequest{
		Question:      "what changed?",
		PreviousCycle: "previous raw",
		CurrentCycle:  "current raw",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "current raw" {
		t.Fatalf("expected current cycle fallback, got %q", result.Output)
	}
	if !result.UsedFallback {
		t.Fatalf("expected fallback output")
	}
	if policy.badSource != "current raw" {
		t.Fatalf("unexpected source passed to bad-distillation check: %q", policy.badSource)
	}
	if policy.badCandidate != "bad summary" {
		t.Fatalf("unexpected candidate passed to bad-distillation check: %q", policy.badCandidate)
	}
}
