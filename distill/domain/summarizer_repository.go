package domain

import "context"

// SummarizerRepository abstracts LLM summarization execution.
type SummarizerRepository interface {
	SummarizeBatch(ctx context.Context, prompt string) (string, error)
	SummarizeWatch(ctx context.Context, prompt string) (string, error)
}
