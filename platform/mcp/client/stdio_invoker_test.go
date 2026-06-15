package client

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func TestStdioInvokerCallTool(t *testing.T) {
	t.Setenv("GO_WANT_MCP_HELPER", "1")

	invoker := NewStdioInvoker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := invoker.CallTool(ctx, distillapp.DistillMCPToolCallRequest{
		ServerCommand: os.Args[0],
		ServerArgs: []string{
			"-test.run=TestStdioInvokerHelperProcess",
			"--",
		},
		ToolName: "echo_json",
		ToolArguments: map[string]any{
			"name": "Edit Service",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Output, "Edit Service") {
		t.Fatalf("expected tool output to contain tool payload, got %q", result.Output)
	}
}

func TestStdioInvokerHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_MCP_HELPER") != "1" {
		return
	}

	server := mcpserver.NewMCPServer("helper-server", "1.0.0", mcpserver.WithToolCapabilities(true))
	server.AddTool(
		mcp.NewTool("echo_json", mcp.WithString("name")),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			payload, err := json.Marshal(request.GetArguments())
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(payload)), nil
		},
	)

	if err := mcpserver.ServeStdio(server); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
