package tools

import (
	"context"
	"errors"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeWatchUseCase struct {
	result  distillapp.DistillWatchResult
	err     error
	request distillapp.DistillWatchRequest
}

func (f *fakeWatchUseCase) Execute(_ context.Context, request distillapp.DistillWatchRequest) (distillapp.DistillWatchResult, error) {
	f.request = request
	return f.result, f.err
}

func TestDistillWatchDefinitionName(t *testing.T) {
	tool := NewDistillWatch(&fakeWatchUseCase{})
	if tool.Definition().Name != "distill_watch" {
		t.Fatalf("unexpected tool name: %s", tool.Definition().Name)
	}
}

func TestDistillWatchHandlerReturnsOutput(t *testing.T) {
	useCase := &fakeWatchUseCase{result: distillapp.DistillWatchResult{Output: "delta"}}
	tool := NewDistillWatch(useCase)

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"question":       "what changed?",
				"previous_cycle": "failed: 0",
				"current_cycle":  "failed: 1",
			},
		},
	})
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
	if text.Text != "delta" {
		t.Fatalf("unexpected output text: %q", text.Text)
	}
	if useCase.request.Question != "what changed?" || useCase.request.PreviousCycle != "failed: 0" || useCase.request.CurrentCycle != "failed: 1" {
		t.Fatalf("unexpected request passed to use case: %#v", useCase.request)
	}
}

func TestDistillWatchHandlerReturnsToolErrorOnValidationFailure(t *testing.T) {
	useCase := &fakeWatchUseCase{err: errors.New("current cycle is required")}
	tool := NewDistillWatch(useCase)

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
	if text.Text != "current cycle is required" {
		t.Fatalf("unexpected error message: %q", text.Text)
	}
}
