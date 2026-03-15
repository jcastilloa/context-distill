package tools

import (
	"context"
	"errors"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeBatchUseCase struct {
	result  distillapp.DistillBatchResult
	err     error
	request distillapp.DistillBatchRequest
}

func (f *fakeBatchUseCase) Execute(_ context.Context, request distillapp.DistillBatchRequest) (distillapp.DistillBatchResult, error) {
	f.request = request
	return f.result, f.err
}

func TestDistillBatchDefinitionName(t *testing.T) {
	tool := NewDistillBatch(&fakeBatchUseCase{})
	if tool.Definition().Name != "distill_batch" {
		t.Fatalf("unexpected tool name: %s", tool.Definition().Name)
	}
}

func TestDistillBatchHandlerReturnsOutput(t *testing.T) {
	useCase := &fakeBatchUseCase{result: distillapp.DistillBatchResult{Output: "summary"}}
	tool := NewDistillBatch(useCase)

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"question": "what changed?",
				"input":    "raw",
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
	if text.Text != "summary" {
		t.Fatalf("unexpected output text: %q", text.Text)
	}
	if useCase.request.Question != "what changed?" || useCase.request.Input != "raw" {
		t.Fatalf("unexpected request passed to use case: %#v", useCase.request)
	}
}

func TestDistillBatchHandlerReturnsToolErrorOnValidationFailure(t *testing.T) {
	useCase := &fakeBatchUseCase{err: errors.New("question is required")}
	tool := NewDistillBatch(useCase)

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
	if text.Text != "question is required" {
		t.Fatalf("unexpected error message: %q", text.Text)
	}
}
