package distillation

import (
	"context"
	"errors"
	"testing"
)

func TestDistillBatchUseCaseExecuteReturnsErrorWhenQuestionIsEmpty(t *testing.T) {
	useCase := NewDistillBatchUseCase(&fakePromptBuilder{}, &fakeTextPolicy{}, &fakeSummarizerRepository{})

	_, err := useCase.Execute(context.Background(), DistillBatchRequest{Input: "payload"})
	if !errors.Is(err, ErrQuestionRequired) {
		t.Fatalf("expected ErrQuestionRequired, got %v", err)
	}
}

func TestDistillBatchUseCaseExecuteReturnsErrorWhenInputIsEmpty(t *testing.T) {
	useCase := NewDistillBatchUseCase(&fakePromptBuilder{}, &fakeTextPolicy{}, &fakeSummarizerRepository{})

	_, err := useCase.Execute(context.Background(), DistillBatchRequest{Question: "summarize"})
	if !errors.Is(err, ErrInputRequired) {
		t.Fatalf("expected ErrInputRequired, got %v", err)
	}
}

func TestDistillBatchUseCaseExecuteReturnsSummarizedOutput(t *testing.T) {
	builder := &fakePromptBuilder{batchPrompt: "batch-prompt"}
	policy := &fakeTextPolicy{normalizedValue: "normalized input"}
	repo := &fakeSummarizerRepository{batchSummary: "summary"}
	useCase := NewDistillBatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillBatchRequest{
		Question: "what changed?",
		Input:    "raw input",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "summary\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if result.UsedFallback {
		t.Fatalf("expected non-fallback output")
	}
	if builder.batchQuestion != "what changed?" {
		t.Fatalf("unexpected question passed to prompt builder: %q", builder.batchQuestion)
	}
	if builder.batchInput != "normalized input" {
		t.Fatalf("expected normalized input, got %q", builder.batchInput)
	}
	if repo.batchPrompt != "batch-prompt" {
		t.Fatalf("unexpected prompt passed to repository: %q", repo.batchPrompt)
	}
	if policy.ensuredValue != "summary" {
		t.Fatalf("expected ensure newline to receive raw summary, got %q", policy.ensuredValue)
	}
}

func TestDistillBatchUseCaseExecuteFallsBackToRawInputWhenSummarizerFails(t *testing.T) {
	builder := &fakePromptBuilder{batchPrompt: "batch-prompt"}
	policy := &fakeTextPolicy{normalizedValue: "normalized input"}
	repo := &fakeSummarizerRepository{batchErr: errors.New("boom")}
	useCase := NewDistillBatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillBatchRequest{
		Question: "what changed?",
		Input:    "raw input",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "raw input" {
		t.Fatalf("expected raw input fallback, got %q", result.Output)
	}
	if !result.UsedFallback {
		t.Fatalf("expected fallback output")
	}
}

func TestDistillBatchUseCaseExecuteFallsBackToRawInputWhenSummaryLooksBad(t *testing.T) {
	builder := &fakePromptBuilder{batchPrompt: "batch-prompt"}
	policy := &fakeTextPolicy{normalizedValue: "normalized input", badDistillation: true}
	repo := &fakeSummarizerRepository{batchSummary: "bad summary"}
	useCase := NewDistillBatchUseCase(builder, policy, repo)

	result, err := useCase.Execute(context.Background(), DistillBatchRequest{
		Question: "what changed?",
		Input:    "raw input",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "raw input" {
		t.Fatalf("expected raw input fallback, got %q", result.Output)
	}
	if !result.UsedFallback {
		t.Fatalf("expected fallback output")
	}
	if policy.badSource != "raw input" {
		t.Fatalf("unexpected source passed to bad-distillation check: %q", policy.badSource)
	}
	if policy.badCandidate != "bad summary" {
		t.Fatalf("unexpected candidate passed to bad-distillation check: %q", policy.badCandidate)
	}
}
