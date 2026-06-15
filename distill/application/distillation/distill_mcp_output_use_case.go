package distillation

import (
	"context"
	"strings"
)

type DistillMCPOutputBatchUseCase interface {
	Execute(ctx context.Context, request DistillBatchRequest) (DistillBatchResult, error)
}

type DistillMCPOutputUseCase struct {
	distillBatch DistillMCPOutputBatchUseCase
	invoker      DistillMCPToolInvoker
}

type DistillMCPToolInvoker interface {
	CallTool(ctx context.Context, request DistillMCPToolCallRequest) (DistillMCPToolCallResult, error)
}

func NewDistillMCPOutputUseCase(
	distillBatch DistillMCPOutputBatchUseCase,
	invoker DistillMCPToolInvoker,
) *DistillMCPOutputUseCase {
	return &DistillMCPOutputUseCase{
		distillBatch: distillBatch,
		invoker:      invoker,
	}
}

func (u *DistillMCPOutputUseCase) Execute(ctx context.Context, request DistillMCPOutputRequest) (DistillMCPOutputResult, error) {
	if strings.TrimSpace(request.Question) == "" {
		return DistillMCPOutputResult{}, ErrQuestionRequired
	}

	output, err := u.resolveOutput(ctx, request)
	if err != nil {
		return DistillMCPOutputResult{}, err
	}

	normalizedInput := buildMCPDistillInput(request.ToolName, output)
	result, err := u.distillBatch.Execute(ctx, DistillBatchRequest{
		Question: request.Question,
		Input:    normalizedInput,
	})
	if err != nil {
		return DistillMCPOutputResult{}, err
	}

	return DistillMCPOutputResult{
		Output:       result.Output,
		UsedFallback: result.UsedFallback,
	}, nil
}

func (u *DistillMCPOutputUseCase) resolveOutput(ctx context.Context, request DistillMCPOutputRequest) (string, error) {
	if strings.TrimSpace(request.Output) != "" {
		return request.Output, nil
	}
	if strings.TrimSpace(request.ServerCommand) == "" && strings.TrimSpace(request.ToolName) == "" {
		return "", ErrMCPInvocationRequired
	}
	if strings.TrimSpace(request.ServerCommand) == "" {
		return "", ErrMCPServerCommandReq
	}
	if strings.TrimSpace(request.ToolName) == "" {
		return "", ErrMCPToolNameRequired
	}

	result, err := u.invoker.CallTool(ctx, DistillMCPToolCallRequest{
		ServerCommand: request.ServerCommand,
		ServerArgs:    request.ServerArgs,
		ToolName:      request.ToolName,
		ToolArguments: request.ToolArguments,
	})
	if err != nil {
		return "", err
	}

	return result.Output, nil
}

func buildMCPDistillInput(toolName, output string) string {
	body := normalizeMCPOutput(output)
	if strings.TrimSpace(toolName) == "" {
		return body
	}

	return "MCP tool: " + strings.TrimSpace(toolName) + "\n\n" + body
}
