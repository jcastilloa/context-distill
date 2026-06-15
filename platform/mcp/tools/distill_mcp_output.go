package tools

import (
	"context"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type DistillMCPOutputUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillMCPOutputRequest) (distillapp.DistillMCPOutputResult, error)
}

type DistillMCPOutput struct {
	useCase DistillMCPOutputUseCase
}

func NewDistillMCPOutput(useCase DistillMCPOutputUseCase) DistillMCPOutput {
	return DistillMCPOutput{useCase: useCase}
}

func (t DistillMCPOutput) Definition() mcp.Tool {
	return mcp.NewTool("distill_mcp_output",
		mcp.WithDescription("Distill raw MCP tool output or invoke an MCP tool and distill its result for a specific question"),
		mcp.WithString("question",
			mcp.Required(),
			mcp.Description("Exact question to answer from the MCP output"),
		),
		mcp.WithString("output",
			mcp.Description("Raw MCP tool output payload to distill"),
		),
		mcp.WithString("tool_name",
			mcp.Description("Optional MCP tool name. Required when invoking an MCP server."),
		),
		mcp.WithString("server_command",
			mcp.Description("Optional MCP server command to invoke when output is omitted"),
		),
		mcp.WithArray("server_args",
			mcp.WithStringItems(),
			mcp.Description("Optional MCP server command arguments"),
		),
		mcp.WithObject("tool_arguments",
			mcp.Description("Optional arguments object passed to the target MCP tool"),
		),
	)
}

func (t DistillMCPOutput) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := t.useCase.Execute(ctx, distillapp.DistillMCPOutputRequest{
		Question:      mcp.ParseString(request, "question", ""),
		ToolName:      mcp.ParseString(request, "tool_name", ""),
		Output:        mcp.ParseString(request, "output", ""),
		ServerCommand: mcp.ParseString(request, "server_command", ""),
		ServerArgs:    parseStringArrayArgument(request, "server_args"),
		ToolArguments: parseObjectArgument(request, "tool_arguments"),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.Output), nil
}

func parseStringArrayArgument(request mcp.CallToolRequest, key string) []string {
	argument := mcp.ParseArgument(request, key, []any{})
	if argument == nil {
		return nil
	}

	rawList, ok := argument.([]any)
	if !ok {
		castList, castOK := argument.([]string)
		if castOK {
			return castList
		}
		return nil
	}

	result := make([]string, 0, len(rawList))
	for _, item := range rawList {
		value, valueOK := item.(string)
		if !valueOK {
			continue
		}
		result = append(result, value)
	}

	return result
}

func parseObjectArgument(request mcp.CallToolRequest, key string) any {
	argument := mcp.ParseArgument(request, key, nil)
	if argument == nil {
		return nil
	}

	if object, ok := argument.(map[string]any); ok {
		return object
	}

	return argument
}
