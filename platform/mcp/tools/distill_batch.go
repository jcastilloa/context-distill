package tools

import (
	"context"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type DistillBatchUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillBatchRequest) (distillapp.DistillBatchResult, error)
}

type DistillBatch struct {
	useCase DistillBatchUseCase
}

func NewDistillBatch(useCase DistillBatchUseCase) DistillBatch {
	return DistillBatch{useCase: useCase}
}

func (t DistillBatch) Definition() mcp.Tool {
	return mcp.NewTool("distill_batch",
		mcp.WithDescription("Distill command output into a compact answer for a specific question"),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("Exact question to answer from the command output"),
		),
		mcp.WithString("input",
			mcp.Required(),
			mcp.Description("Raw command output to distill"),
		),
	)
}

func (t DistillBatch) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := t.useCase.Execute(ctx, distillapp.DistillBatchRequest{
		Question: mcp.ParseString(request, "question", ""),
		Input:    mcp.ParseString(request, "input", ""),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.Output), nil
}
