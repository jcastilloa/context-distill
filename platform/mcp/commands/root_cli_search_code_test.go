package commands

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"
)

type fakeSearchCodeCLIUseCase struct {
	request distillapp.SearchCodeRequest
	result  distillapp.SearchCodeResult
	err     error
}

func (f *fakeSearchCodeCLIUseCase) Execute(_ context.Context, request distillapp.SearchCodeRequest) (distillapp.SearchCodeResult, error) {
	f.request = request
	if f.err != nil {
		return distillapp.SearchCodeResult{}, f.err
	}
	return f.result, nil
}

func TestRootCommandSearchCodeRunsUseCaseAndPrintsOutput(t *testing.T) {
	searchUseCase := &fakeSearchCodeCLIUseCase{result: distillapp.SearchCodeResult{Output: "platform/config/load.go:27\n"}}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	).WithSearchCode(tools.SearchCode{}, searchUseCase)

	stdout := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"search_code",
		"--query", "provider_name",
		"--mode", "text",
		"--question", "Return only file:line, one per line.",
		"--scope", "platform/**/*.go,cmd/**/*.go",
		"--max-results", "7",
		"--context-lines", "1",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.String() != "platform/config/load.go:27\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if searchUseCase.request.Query != "provider_name" {
		t.Fatalf("unexpected query: %q", searchUseCase.request.Query)
	}
	if searchUseCase.request.Mode != "text" {
		t.Fatalf("unexpected mode: %q", searchUseCase.request.Mode)
	}
	if searchUseCase.request.MaxResults != 7 || searchUseCase.request.ContextLines != 1 {
		t.Fatalf("unexpected limits: %#v", searchUseCase.request)
	}
	if len(searchUseCase.request.Scope) != 2 {
		t.Fatalf("unexpected scope: %#v", searchUseCase.request.Scope)
	}
}

func TestRootCommandSearchCodeSupportsViperEnvOverrides(t *testing.T) {
	searchUseCase := &fakeSearchCodeCLIUseCase{result: distillapp.SearchCodeResult{Output: "ok\n"}}
	t.Setenv("MCP_SEARCH_CODE_QUERY", "LoadDistillConfig")
	t.Setenv("MCP_SEARCH_CODE_MODE", "symbol")
	t.Setenv("MCP_SEARCH_CODE_QUESTION", "Return definitions first.")
	t.Setenv("MCP_SEARCH_CODE_SCOPE", "cmd/**/*.go,platform/**/*.go")
	t.Setenv("MCP_SEARCH_CODE_MAX_RESULTS", "3")
	t.Setenv("MCP_SEARCH_CODE_CONTEXT_LINES", "0")

	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	).WithSearchCode(tools.SearchCode{}, searchUseCase)

	cmd := runner.newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"search_code"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchUseCase.request.Query != "LoadDistillConfig" || searchUseCase.request.Mode != "symbol" {
		t.Fatalf("expected env overrides, got %#v", searchUseCase.request)
	}
	if searchUseCase.request.MaxResults != 3 || searchUseCase.request.ContextLines != 0 {
		t.Fatalf("expected env numeric overrides, got %#v", searchUseCase.request)
	}
}

func TestRootCommandSearchCodeReturnsErrorWhenRequiredValuesMissing(t *testing.T) {
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	).WithSearchCode(tools.SearchCode{}, &fakeSearchCodeCLIUseCase{err: distillapp.ErrQueryRequired})

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"search_code"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Fatalf("expected query error, got %v", err)
	}
}

func TestRootCommandSearchCodeReturnsUseCaseError(t *testing.T) {
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	).WithSearchCode(tools.SearchCode{}, &fakeSearchCodeCLIUseCase{err: errors.New("search failed")})

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"search_code", "--query", "provider_name", "--mode", "text", "--question", "q"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command error")
	}
	if !strings.Contains(err.Error(), "search failed") {
		t.Fatalf("expected search failed error, got %v", err)
	}
}
