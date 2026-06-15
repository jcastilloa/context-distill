package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

const distillMCPClientName = "context-distill"

type StdioInvoker struct{}

func NewStdioInvoker() StdioInvoker {
	return StdioInvoker{}
}

func (StdioInvoker) CallTool(ctx context.Context, request distillapp.DistillMCPToolCallRequest) (distillapp.DistillMCPToolCallResult, error) {
	client, err := mcpclient.NewStdioMCPClientWithOptions(
		request.ServerCommand,
		os.Environ(),
		request.ServerArgs,
	)
	if err != nil {
		return distillapp.DistillMCPToolCallResult{}, fmt.Errorf("start mcp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if _, err = client.Initialize(ctx, buildInitializeRequest()); err != nil {
		return distillapp.DistillMCPToolCallResult{}, fmt.Errorf("initialize mcp client: %w", err)
	}

	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      request.ToolName,
			Arguments: request.ToolArguments,
		},
	})
	if err != nil {
		return distillapp.DistillMCPToolCallResult{}, fmt.Errorf("call mcp tool %s: %w", request.ToolName, err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return distillapp.DistillMCPToolCallResult{}, fmt.Errorf("marshal mcp tool result: %w", err)
	}

	return distillapp.DistillMCPToolCallResult{Output: string(payload)}, nil
}

func buildInitializeRequest() mcp.InitializeRequest {
	return mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    distillMCPClientName,
				Version: "dev",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}
}
