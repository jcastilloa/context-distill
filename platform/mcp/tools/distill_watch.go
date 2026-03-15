package tools

import (
	"context"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type DistillWatchUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillWatchRequest) (distillapp.DistillWatchResult, error)
}

type DistillWatch struct {
	useCase DistillWatchUseCase
}

func NewDistillWatch(useCase DistillWatchUseCase) DistillWatch {
	return DistillWatch{useCase: useCase}
}

func (t DistillWatch) Definition() mcp.Tool {
	return mcp.NewTool("distill_watch",
		mcp.WithDescription("Distill changes between two watch-mode cycles for a specific question"),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("Exact question to answer from cycle changes"),
		),
		mcp.WithString("previous_cycle",
			mcp.Required(),
			mcp.Description("Previous watch cycle output"),
		),
		mcp.WithString("current_cycle",
			mcp.Required(),
			mcp.Description("Current watch cycle output"),
		),
	)
}

func (t DistillWatch) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := t.useCase.Execute(ctx, distillapp.DistillWatchRequest{
		Question:      mcp.ParseString(request, "question", ""),
		PreviousCycle: mcp.ParseString(request, "previous_cycle", ""),
		CurrentCycle:  mcp.ParseString(request, "current_cycle", ""),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.Output), nil
}
