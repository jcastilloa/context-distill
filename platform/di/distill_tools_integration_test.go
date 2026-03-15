package di

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	githubdi "github.com/sarulabs/di"

	"github.com/jcastilloa/context-distill/platform/config"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

type scriptedAIRepository struct {
	responseText string
}

func (r scriptedAIRepository) GetAIResponse(context.Context, *aiDomain.Request) (*aiDomain.Response, error) {
	return aiDomain.NewResponse(r.responseText, nil, "", "", "", nil, false, "")
}

func TestDistillBatchToolIntegrationWithOllamaProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}

		defer req.Body.Close()
		payload := map[string]any{}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		if payload["prompt"] == "" {
			t.Fatalf("expected non-empty prompt")
		}

		_, _ = w.Write([]byte(`{"response":"All tests passed."}`))
	}))
	defer server.Close()

	container := newContainerFromViperConfig(
		t,
		`service:
  version: test
distill:
  provider_name: ollama
  base_url: `+server.URL+`
  model: qwen3.5:2b
  timeout: 3s
`,
		scriptedAIRepository{responseText: "unused"},
	)

	tool := (*container).Get(DistillBatchToolLabel).(interface {
		Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"question": "did tests pass?",
			"input":    "Ran 12 tests\\n12 passed",
		}},
	})
	if err != nil {
		t.Fatalf("unexpected tool handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful tool result")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content")
	}
	if text.Text != "All tests passed.\n" {
		t.Fatalf("unexpected distilled output: %q", text.Text)
	}
}

func TestDistillWatchToolIntegrationWithOpenAICompatibleProvider(t *testing.T) {
	container := newContainerFromViperConfig(
		t,
		`service:
  version: test
distill:
  provider_name: openai-compatible
  base_url: http://127.0.0.1:9000/v1
`,
		scriptedAIRepository{responseText: "Failure count changed from 0 to 1."},
	)

	tool := (*container).Get(DistillWatchToolLabel).(interface {
		Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"question":       "what changed?",
			"previous_cycle": "failures: 0",
			"current_cycle":  "failures: 1",
		}},
	})
	if err != nil {
		t.Fatalf("unexpected tool handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful tool result")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content")
	}
	if text.Text != "Failure count changed from 0 to 1.\n" {
		t.Fatalf("unexpected distilled output: %q", text.Text)
	}
}

func TestDistillBatchToolIntegrationReturnsToolErrorForValidation(t *testing.T) {
	container := newContainerFromViperConfig(
		t,
		`service:
  version: test
distill:
  provider_name: openai-compatible
  base_url: http://127.0.0.1:9000/v1
`,
		scriptedAIRepository{responseText: "unused"},
	)

	tool := (*container).Get(DistillBatchToolLabel).(interface {
		Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"question": "",
			"input":    "payload",
		}},
	})
	if err != nil {
		t.Fatalf("unexpected tool handler error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected tool-level validation error")
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content")
	}
	if text.Text != "question is required" {
		t.Fatalf("unexpected validation error text: %q", text.Text)
	}
}

func newContainerFromViperConfig(t *testing.T, configYAML string, aiRepository aiDomain.AIRepository) *githubdi.Container {
	t.Helper()

	workspace := t.TempDir()
	t.Setenv("HOME", workspace)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("APPDATA", filepath.Join(workspace, "AppData", "Roaming"))
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err = os.Chdir(workspace); err != nil {
		t.Fatalf("change directory: %v", err)
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("resolve user config directory: %v", err)
	}
	configDir := filepath.Join(userConfigDir, "context-distill")
	if err = os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if err = os.WriteFile(configPath, []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfgRepo, err := config.New("context-distill")
	if err != nil {
		t.Fatalf("load viper config: %v", err)
	}

	builder := New(aiRepository, cfgRepo.DistillProviderConfig(), "context-distill", cfgRepo.ServiceConfig())
	container, err := builder.Build()
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	return container
}
