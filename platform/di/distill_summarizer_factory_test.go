package di

import (
	"testing"

	"github.com/jcastilloa/context-distill/platform/ollama"
	"github.com/jcastilloa/context-distill/platform/openai"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

func TestNewDistillSummarizerRepositorySelectsOllama(t *testing.T) {
	repo, err := newDistillSummarizerRepository(aiDomain.ProviderConfig{ProviderName: "ollama"}, fakeAIRepository{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.(*ollama.DistillSummarizerRepository); !ok {
		t.Fatalf("expected ollama repository, got %T", repo)
	}
}

func TestNewDistillSummarizerRepositorySelectsOpenAICompatible(t *testing.T) {
	repo, err := newDistillSummarizerRepository(aiDomain.ProviderConfig{ProviderName: "openai-compatible"}, fakeAIRepository{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.(*openai.DistillSummarizerRepository); !ok {
		t.Fatalf("expected openai-compatible repository, got %T", repo)
	}
}

func TestNewDistillSummarizerRepositoryReturnsErrorOnUnknownProvider(t *testing.T) {
	_, err := newDistillSummarizerRepository(aiDomain.ProviderConfig{ProviderName: "unknown-provider"}, fakeAIRepository{})
	if err == nil {
		t.Fatalf("expected error for unsupported provider")
	}
}
