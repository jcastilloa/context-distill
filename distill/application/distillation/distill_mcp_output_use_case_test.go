package distillation

import (
	"context"
	"testing"
)

type fakeDistillMCPBatchUseCase struct {
	request DistillBatchRequest
	result  DistillBatchResult
	err     error
}

func (f *fakeDistillMCPBatchUseCase) Execute(_ context.Context, request DistillBatchRequest) (DistillBatchResult, error) {
	f.request = request
	if f.err != nil {
		return DistillBatchResult{}, f.err
	}
	return f.result, nil
}

type fakeMCPOutputInvoker struct {
	request DistillMCPToolCallRequest
	result  DistillMCPToolCallResult
	err     error
}

func (f *fakeMCPOutputInvoker) CallTool(_ context.Context, request DistillMCPToolCallRequest) (DistillMCPToolCallResult, error) {
	f.request = request
	if f.err != nil {
		return DistillMCPToolCallResult{}, f.err
	}
	return f.result, nil
}

func TestDistillMCPOutputUseCaseExtractsTextContentAndDelegatesToBatch(t *testing.T) {
	batch := &fakeDistillMCPBatchUseCase{
		result: DistillBatchResult{Output: "Edit Service\n"},
	}
	useCase := NewDistillMCPOutputUseCase(batch, &fakeMCPOutputInvoker{})

	result, err := useCase.Execute(context.Background(), DistillMCPOutputRequest{
		Question: "Return only job names, one per line.",
		ToolName: "jenkins_job_list",
		Output:   `[{"type":"text","text":"[{\"name\":\"Edit Service\"}]"}]`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "Edit Service\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if batch.request.Question != "Return only job names, one per line." {
		t.Fatalf("unexpected question: %q", batch.request.Question)
	}

	expectedInput := "MCP tool: jenkins_job_list\n\n[{\"name\":\"Edit Service\"}]"
	if batch.request.Input != expectedInput {
		t.Fatalf("unexpected normalized input: %q", batch.request.Input)
	}
}

func TestDistillMCPOutputUseCaseUsesRawOutputWhenPayloadIsPlainText(t *testing.T) {
	batch := &fakeDistillMCPBatchUseCase{
		result: DistillBatchResult{Output: "summary\n"},
	}
	useCase := NewDistillMCPOutputUseCase(batch, &fakeMCPOutputInvoker{})

	_, err := useCase.Execute(context.Background(), DistillMCPOutputRequest{
		Question: "Summarize.",
		Output:   "plain output",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batch.request.Input != "plain output" {
		t.Fatalf("expected raw output passthrough, got %q", batch.request.Input)
	}
}

func TestDistillMCPOutputUseCaseCallsMCPToolWhenRawOutputMissing(t *testing.T) {
	batch := &fakeDistillMCPBatchUseCase{
		result: DistillBatchResult{Output: "29\n"},
	}
	invoker := &fakeMCPOutputInvoker{
		result: DistillMCPToolCallResult{
			Output: `{"content":[{"type":"text","text":"{\"job_count\":29,\"node_count\":0}"}]}`,
		},
	}
	useCase := NewDistillMCPOutputUseCase(batch, invoker)

	result, err := useCase.Execute(context.Background(), DistillMCPOutputRequest{
		Question:      "Return only job_count.",
		ToolName:      "jenkins_info",
		ServerCommand: "gaz-mcp",
		ServerArgs:    []string{"--transport", "stdio"},
		ToolArguments: map[string]any{"environment": "development"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "29\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if invoker.request.ServerCommand != "gaz-mcp" {
		t.Fatalf("unexpected server command: %q", invoker.request.ServerCommand)
	}
	if invoker.request.ToolName != "jenkins_info" {
		t.Fatalf("unexpected tool name: %q", invoker.request.ToolName)
	}

	expectedInput := "MCP tool: jenkins_info\n\n{\"job_count\":29,\"node_count\":0}"
	if batch.request.Input != expectedInput {
		t.Fatalf("unexpected normalized input: %q", batch.request.Input)
	}
}

func TestDistillMCPOutputUseCaseRequiresQuestion(t *testing.T) {
	useCase := NewDistillMCPOutputUseCase(&fakeDistillMCPBatchUseCase{}, &fakeMCPOutputInvoker{})

	_, err := useCase.Execute(context.Background(), DistillMCPOutputRequest{
		Output: `[{"type":"text","text":"ok"}]`,
	})
	if err != ErrQuestionRequired {
		t.Fatalf("expected ErrQuestionRequired, got %v", err)
	}
}

func TestDistillMCPOutputUseCaseRequiresOutput(t *testing.T) {
	useCase := NewDistillMCPOutputUseCase(&fakeDistillMCPBatchUseCase{}, &fakeMCPOutputInvoker{})

	_, err := useCase.Execute(context.Background(), DistillMCPOutputRequest{
		Question: "q",
	})
	if err != ErrMCPInvocationRequired {
		t.Fatalf("expected ErrMCPInvocationRequired, got %v", err)
	}
}
