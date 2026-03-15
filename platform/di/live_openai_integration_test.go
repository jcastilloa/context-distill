//go:build live

package di

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
	"github.com/jcastilloa/context-distill/platform/config"
	"github.com/jcastilloa/context-distill/platform/openai"
)

func TestLiveDistillBatchWithOpenAICompatibleProvider(t *testing.T) {
	if os.Getenv("DISTILL_LIVE_TEST") != "1" {
		t.Skip("set DISTILL_LIVE_TEST=1 to run live provider integration test")
	}

	projectRoot := locateProjectRoot(t)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err = os.Chdir(projectRoot); err != nil {
		t.Fatalf("chdir to project root: %v", err)
	}

	cfgRepo, err := config.New("context-distill")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	openAICfg := cfgRepo.OpenAIProviderConfig()
	if strings.TrimSpace(openAICfg.APIKey) == "" {
		t.Fatalf("openai.api_key is empty; configure OPENAI_API_KEY or openai.api_key before running live test")
	}

	distillCfg := cfgRepo.DistillProviderConfig()
	distillCfg.ProviderName = "openai-compatible"

	builder := New(
		openai.NewOpenAIRepository(openAICfg, nil),
		distillCfg,
		"context-distill",
		cfgRepo.ServiceConfig(),
	)

	container, err := builder.Build()
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	summarizer := (*container).Get(DistillSummarizerRepositoryLabel).(distilldomain.SummarizerRepository)
	rawSummary, err := summarizer.SummarizeBatch(context.Background(), "Did tests pass? Return only PASS or FAIL.")
	if err != nil {
		t.Fatalf("live provider summarization failed: %v", err)
	}
	if strings.TrimSpace(rawSummary) == "" {
		t.Fatalf("live provider returned empty summary")
	}
	t.Logf("live provider summary: %s", strings.TrimSpace(rawSummary))

	tool := (*container).Get(DistillBatchToolLabel).(interface {
		Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	})

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"question": "Did the tests pass? Return only PASS or FAIL.",
			"input":    "Ran 12 tests\n12 passed",
		}},
	})
	if err != nil {
		t.Fatalf("tool handler error: %v", err)
	}
	if result.IsError {
		text := mcp.GetTextFromContent(result.Content[0])
		t.Fatalf("tool returned error result: %s", text)
	}

	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text output content")
	}

	output := strings.TrimSpace(text.Text)
	if output == "" {
		t.Fatalf("empty live response")
	}

	t.Logf("live distill output: %s", output)
}

func locateProjectRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve caller path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
