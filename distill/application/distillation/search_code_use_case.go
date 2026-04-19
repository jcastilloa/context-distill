package distillation

import (
	"context"
	"encoding/json"
	"fmt"
)

type SearchCodeSearcher interface {
	Search(ctx context.Context, request SearchCodeRequest) ([]SearchMatch, error)
}

type SearchCodeDistiller interface {
	Execute(ctx context.Context, request DistillBatchRequest) (DistillBatchResult, error)
}

type SearchCodeUseCase struct {
	searcher  SearchCodeSearcher
	distiller SearchCodeDistiller
}

func NewSearchCodeUseCase(searcher SearchCodeSearcher, distiller SearchCodeDistiller) *SearchCodeUseCase {
	return &SearchCodeUseCase{searcher: searcher, distiller: distiller}
}

func (u *SearchCodeUseCase) Execute(ctx context.Context, request SearchCodeRequest) (SearchCodeResult, error) {
	normalized := request.WithDefaults()
	if err := normalized.Validate(); err != nil {
		return SearchCodeResult{}, err
	}

	rawMatches, err := u.searcher.Search(ctx, normalized)
	if err != nil {
		return SearchCodeResult{}, err
	}

	compacted := compactSearchMatches(normalized, rawMatches)
	payload, err := serializeSearchMatches(compacted)
	if err != nil {
		return SearchCodeResult{}, fmt.Errorf("serialize matches: %w", err)
	}

	distilled, err := u.distiller.Execute(ctx, DistillBatchRequest{
		Question: normalized.Question,
		Input:    payload,
	})
	if err != nil {
		return SearchCodeResult{}, err
	}

	return SearchCodeResult{
		RawMatches:   compacted,
		Output:       distilled.Output,
		UsedFallback: distilled.UsedFallback,
	}, nil
}

func serializeSearchMatches(matches []SearchMatch) (string, error) {
	payload, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
