package tools

import (
	"context"
	"errors"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeSearchCodeUseCase struct {
	request distillapp.SearchCodeRequest
	result  distillapp.SearchCodeResult
	err     error
}

func (f *fakeSearchCodeUseCase) Execute(_ context.Context, request distillapp.SearchCodeRequest) (distillapp.SearchCodeResult, error) {
	f.request = request
	if f.err != nil {
		return distillapp.SearchCodeResult{}, f.err
	}
	return f.result, nil
}

func TestSearchCodeDefinitionName(t *testing.T) {
	tool := NewSearchCode(&fakeSearchCodeUseCase{})
	if tool.Definition().Name != "search_code" {
		t.Fatalf("unexpected tool name: %s", tool.Definition().Name)
	}
}

func TestSearchCodeHandlerReturnsOutput(t *testing.T) {
	useCase := &fakeSearchCodeUseCase{result: distillapp.SearchCodeResult{Output: "a.go:10\n"}}
	tool := NewSearchCode(useCase)

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"query":         "LoadDistillConfig",
		"mode":          "symbol",
		"question":      "Return definitions first.",
		"scope":         []any{"cmd/**/*.go", "platform/**/*.go"},
		"max_results":   5,
		"context_lines": 2,
	}}})
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error tool result")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content result")
	}
	if text.Text != "a.go:10\n" {
		t.Fatalf("unexpected output text: %q", text.Text)
	}

	if useCase.request.Query != "LoadDistillConfig" || useCase.request.Mode != "symbol" {
		t.Fatalf("unexpected request passed to use case: %#v", useCase.request)
	}
	if len(useCase.request.Scope) != 2 {
		t.Fatalf("expected parsed scope list, got %#v", useCase.request.Scope)
	}
}

func TestSearchCodeHandlerReturnsToolErrorOnValidationFailure(t *testing.T) {
	tool := NewSearchCode(&fakeSearchCodeUseCase{err: errors.New("query is required")})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}})
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected tool result with error flag")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content result")
	}
	if text.Text != "query is required" {
		t.Fatalf("unexpected error message: %q", text.Text)
	}
}
