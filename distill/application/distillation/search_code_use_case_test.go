package distillation

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeSearchCodeSearcher struct {
	request SearchCodeRequest
	matches []SearchMatch
	err     error
}

func (f *fakeSearchCodeSearcher) Search(_ context.Context, request SearchCodeRequest) ([]SearchMatch, error) {
	f.request = request
	if f.err != nil {
		return nil, f.err
	}
	return f.matches, nil
}

type fakeSearchCodeDistiller struct {
	request DistillBatchRequest
	result  DistillBatchResult
	err     error
}

func (f *fakeSearchCodeDistiller) Execute(_ context.Context, request DistillBatchRequest) (DistillBatchResult, error) {
	f.request = request
	if f.err != nil {
		return DistillBatchResult{}, f.err
	}
	return f.result, nil
}

func TestSearchCodeRequestValidate(t *testing.T) {
	testCases := []struct {
		name    string
		request SearchCodeRequest
		err     error
	}{
		{
			name: "fails when query empty",
			request: SearchCodeRequest{
				Mode:     SearchModeText,
				Question: "Return only file paths.",
			},
			err: ErrQueryRequired,
		},
		{
			name: "fails when question empty",
			request: SearchCodeRequest{
				Query: "provider_name",
				Mode:  SearchModeText,
			},
			err: ErrQuestionRequired,
		},
		{
			name: "fails when mode invalid",
			request: SearchCodeRequest{
				Query:    "provider_name",
				Mode:     "invalid",
				Question: "Return only file paths.",
			},
			err: ErrUnsupportedSearchMode,
		},
		{
			name: "fails when max results negative",
			request: SearchCodeRequest{
				Query:      "provider_name",
				Mode:       SearchModeText,
				Question:   "Return only file paths.",
				MaxResults: -1,
			},
			err: ErrInvalidMaxResults,
		},
		{
			name: "fails when context lines negative",
			request: SearchCodeRequest{
				Query:        "provider_name",
				Mode:         SearchModeText,
				Question:     "Return only file paths.",
				ContextLines: -1,
			},
			err: ErrInvalidContextLines,
		},
		{
			name: "passes when valid",
			request: SearchCodeRequest{
				Query:        "provider_name",
				Mode:         SearchModeText,
				Question:     "Return only file paths.",
				MaxResults:   10,
				ContextLines: 2,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.request.Validate()
			if testCase.err == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if !errors.Is(err, testCase.err) {
				t.Fatalf("expected error %v, got %v", testCase.err, err)
			}
		})
	}
}

func TestSearchCodeUseCaseExecuteReturnsDistilledOutput(t *testing.T) {
	searcher := &fakeSearchCodeSearcher{matches: []SearchMatch{
		{File: "a.go", Line: 10, Snippet: strings.Repeat("x", 300), Kind: SearchMatchKindUsage},
		{File: "a.go", Line: 10, Snippet: "duplicated", Kind: SearchMatchKindUsage},
		{File: "b.go", Line: 5, Snippet: "func Find()", Kind: SearchMatchKindDefinition},
		{File: "c.go", Line: 1, Snippet: "provider_name", Kind: SearchMatchKindUsage},
	}}
	distiller := &fakeSearchCodeDistiller{result: DistillBatchResult{Output: "b.go:5\na.go:10\n"}}
	useCase := NewSearchCodeUseCase(searcher, distiller)

	request := SearchCodeRequest{
		Query:        "Find",
		Mode:         SearchModeSymbol,
		Question:     "Return definitions first as file:line, one per line.",
		MaxResults:   2,
		ContextLines: 2,
		Scope:        []string{"  cmd/**/*.go  ", "", "platform/**/*.go"},
	}

	result, err := useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "b.go:5\na.go:10\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}

	if len(result.RawMatches) != 2 {
		t.Fatalf("expected 2 matches after compacting, got %d", len(result.RawMatches))
	}
	if result.RawMatches[0].File != "b.go" || result.RawMatches[0].Kind != SearchMatchKindDefinition {
		t.Fatalf("expected definition first, got %#v", result.RawMatches[0])
	}
	if !strings.HasSuffix(result.RawMatches[1].Snippet, "...") {
		t.Fatalf("expected long snippet to be trimmed, got %q", result.RawMatches[1].Snippet)
	}

	if searcher.request.MaxResults != 2 {
		t.Fatalf("expected max results passed to searcher, got %d", searcher.request.MaxResults)
	}
	if len(searcher.request.Scope) != 2 || searcher.request.Scope[0] != "cmd/**/*.go" || searcher.request.Scope[1] != "platform/**/*.go" {
		t.Fatalf("unexpected normalized scope: %#v", searcher.request.Scope)
	}

	if distiller.request.Question != request.Question {
		t.Fatalf("unexpected distiller question: %q", distiller.request.Question)
	}
	if !strings.Contains(distiller.request.Input, "\"file\": \"b.go\"") {
		t.Fatalf("expected serialized matches in distiller input, got %q", distiller.request.Input)
	}
}

func TestSearchCodeUseCaseExecuteUsesDefaultsWhenZeroValues(t *testing.T) {
	searcher := &fakeSearchCodeSearcher{}
	distiller := &fakeSearchCodeDistiller{result: DistillBatchResult{Output: "[]\n"}}
	useCase := NewSearchCodeUseCase(searcher, distiller)

	_, err := useCase.Execute(context.Background(), SearchCodeRequest{
		Query:    "provider_name",
		Mode:     SearchModeText,
		Question: "Return matches.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if searcher.request.MaxResults != DefaultSearchCodeMaxResults {
		t.Fatalf("expected default max results %d, got %d", DefaultSearchCodeMaxResults, searcher.request.MaxResults)
	}
	if searcher.request.ContextLines != 0 {
		t.Fatalf("expected explicit context lines value to be preserved, got %d", searcher.request.ContextLines)
	}
}

func TestSearchCodeUseCaseExecuteReturnsSearcherError(t *testing.T) {
	searcher := &fakeSearchCodeSearcher{err: errors.New("search failed")}
	useCase := NewSearchCodeUseCase(searcher, &fakeSearchCodeDistiller{})

	_, err := useCase.Execute(context.Background(), SearchCodeRequest{
		Query:    "provider_name",
		Mode:     SearchModeText,
		Question: "Return matches.",
	})
	if err == nil || !strings.Contains(err.Error(), "search failed") {
		t.Fatalf("expected search failed error, got %v", err)
	}
}

func TestSearchCodeUseCaseExecuteReturnsDistillerError(t *testing.T) {
	searcher := &fakeSearchCodeSearcher{}
	distiller := &fakeSearchCodeDistiller{err: errors.New("distill failed")}
	useCase := NewSearchCodeUseCase(searcher, distiller)

	_, err := useCase.Execute(context.Background(), SearchCodeRequest{
		Query:    "provider_name",
		Mode:     SearchModeText,
		Question: "Return matches.",
	})
	if err == nil || !strings.Contains(err.Error(), "distill failed") {
		t.Fatalf("expected distill failed error, got %v", err)
	}
}
