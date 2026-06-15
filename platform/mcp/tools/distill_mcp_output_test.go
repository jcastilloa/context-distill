package tools

import (
	"context"
	"errors"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeDistillMCPOutputUseCase struct {
	request distillapp.DistillMCPOutputRequest
	result  distillapp.DistillMCPOutputResult
	err     error
}

func (f *fakeDistillMCPOutputUseCase) Execute(_ context.Context, request distillapp.DistillMCPOutputRequest) (distillapp.DistillMCPOutputResult, error) {
	f.request = request
	if f.err != nil {
		return distillapp.DistillMCPOutputResult{}, f.err
	}
	return f.result, nil
}

func TestDistillMCPOutputDefinitionName(t *testing.T) {
	tool := NewDistillMCPOutput(&fakeDistillMCPOutputUseCase{})
	if tool.Definition().Name != "distill_mcp_output" {
		t.Fatalf("unexpected tool name: %s", tool.Definition().Name)
	}
}

func TestDistillMCPOutputHandlerReturnsOutput(t *testing.T) {
	useCase := &fakeDistillMCPOutputUseCase{
		result: distillapp.DistillMCPOutputResult{Output: "Edit Service\n"},
	}
	tool := NewDistillMCPOutput(useCase)

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"question":       "Return only job names.",
				"tool_name":      "jenkins_job_list",
				"output":         `[{"type":"text","text":"[{\"name\":\"Edit Service\"}]"}]`,
				"server_command": "gaz-mcp",
				"server_args":    []any{"--transport", "stdio"},
				"tool_arguments": map[string]any{"environment": "development"},
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
	if text.Text != "Edit Service\n" {
		t.Fatalf("unexpected output text: %q", text.Text)
	}
	if useCase.request.ToolName != "jenkins_job_list" {
		t.Fatalf("unexpected tool name: %q", useCase.request.ToolName)
	}
	if useCase.request.ServerCommand != "gaz-mcp" {
		t.Fatalf("unexpected server command: %q", useCase.request.ServerCommand)
	}
	if len(useCase.request.ServerArgs) != 2 {
		t.Fatalf("unexpected server args: %#v", useCase.request.ServerArgs)
	}
}

func TestDistillMCPOutputHandlerReturnsToolErrorOnValidationFailure(t *testing.T) {
	tool := NewDistillMCPOutput(&fakeDistillMCPOutputUseCase{
		err: errors.New("mcp output is required"),
	})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
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
	if text.Text != "mcp output is required" {
		t.Fatalf("unexpected error message: %q", text.Text)
	}
}
