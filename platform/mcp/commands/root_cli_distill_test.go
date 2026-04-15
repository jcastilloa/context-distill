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

type fakeDistillBatchUseCase struct {
	request distillapp.DistillBatchRequest
	result  distillapp.DistillBatchResult
	err     error
}

func (u *fakeDistillBatchUseCase) Execute(_ context.Context, request distillapp.DistillBatchRequest) (distillapp.DistillBatchResult, error) {
	u.request = request
	if u.err != nil {
		return distillapp.DistillBatchResult{}, u.err
	}
	return u.result, nil
}

type fakeDistillWatchUseCase struct {
	request distillapp.DistillWatchRequest
	result  distillapp.DistillWatchResult
	err     error
}

func (u *fakeDistillWatchUseCase) Execute(_ context.Context, request distillapp.DistillWatchRequest) (distillapp.DistillWatchResult, error) {
	u.request = request
	if u.err != nil {
		return distillapp.DistillWatchResult{}, u.err
	}
	return u.result, nil
}

func TestRootCommandDistillBatchRunsUseCaseAndPrintsOutput(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{result: distillapp.DistillBatchResult{Output: "PASS\n"}}
	watchUseCase := &fakeDistillWatchUseCase{}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		watchUseCase,
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"distill_batch", "--question", "Return only PASS or FAIL.", "--input", "ok"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.String() != "PASS\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if batchUseCase.request.Question != "Return only PASS or FAIL." {
		t.Fatalf("unexpected question: %q", batchUseCase.request.Question)
	}
	if batchUseCase.request.Input != "ok" {
		t.Fatalf("unexpected input: %q", batchUseCase.request.Input)
	}
}

func TestRootCommandDistillBatchReturnsUseCaseError(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{err: errors.New("boom")}
	watchUseCase := &fakeDistillWatchUseCase{}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		watchUseCase,
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"distill_batch", "--question", "q", "--input", "raw"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestRootCommandDistillBatchRequiresQuestion(t *testing.T) {
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		&fakeDistillBatchUseCase{},
		&fakeDistillWatchUseCase{},
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"distill_batch"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected required flags error")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("expected required flag error, got %v", err)
	}
}

func TestRootCommandDistillBatchReadsInputFromStdin(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{result: distillapp.DistillBatchResult{Output: "FAIL\n"}}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		&fakeDistillWatchUseCase{},
	)

	stdinContent := "PASS: TestA, FAIL: TestB - expected 4 got 5"
	stdout := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader(stdinContent))
	cmd.SetArgs([]string{"distill_batch", "--question", "Did tests pass?"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if batchUseCase.request.Input != stdinContent {
		t.Fatalf("expected input from stdin %q, got %q", stdinContent, batchUseCase.request.Input)
	}
	if stdout.String() != "FAIL\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRootCommandDistillBatchInputDashReadsFromStdin(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{result: distillapp.DistillBatchResult{Output: "ok\n"}}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		&fakeDistillWatchUseCase{},
	)

	stdinContent := "piped data from stdin"
	stdout := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader(stdinContent))
	cmd.SetArgs([]string{"distill_batch", "--question", "summarize", "--input", "-"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if batchUseCase.request.Input != stdinContent {
		t.Fatalf("expected input from stdin %q, got %q", stdinContent, batchUseCase.request.Input)
	}
}

func TestRootCommandDistillBatchFlagInputTakesPrecedenceOverStdin(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{result: distillapp.DistillBatchResult{Output: "done\n"}}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		&fakeDistillWatchUseCase{},
	)

	stdout := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader("stdin data that should be ignored"))
	cmd.SetArgs([]string{"distill_batch", "--question", "q", "--input", "flag data"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if batchUseCase.request.Input != "flag data" {
		t.Fatalf("expected flag input %q, got %q", "flag data", batchUseCase.request.Input)
	}
}

func TestRootCommandDistillWatchRunsUseCaseAndPrintsOutput(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{}
	watchUseCase := &fakeDistillWatchUseCase{result: distillapp.DistillWatchResult{Output: "changed\n"}}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		watchUseCase,
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := runner.newRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"distill_watch",
		"--question", "What changed? Return one line.",
		"--previous-cycle", "a",
		"--current-cycle", "b",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.String() != "changed\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if watchUseCase.request.Question != "What changed? Return one line." {
		t.Fatalf("unexpected question: %q", watchUseCase.request.Question)
	}
	if watchUseCase.request.PreviousCycle != "a" {
		t.Fatalf("unexpected previous cycle: %q", watchUseCase.request.PreviousCycle)
	}
	if watchUseCase.request.CurrentCycle != "b" {
		t.Fatalf("unexpected current cycle: %q", watchUseCase.request.CurrentCycle)
	}
}

func TestRootCommandDistillWatchReturnsUseCaseError(t *testing.T) {
	batchUseCase := &fakeDistillBatchUseCase{}
	watchUseCase := &fakeDistillWatchUseCase{err: errors.New("watch failed")}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		batchUseCase,
		watchUseCase,
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{
		"distill_watch",
		"--question", "q",
		"--previous-cycle", "a",
		"--current-cycle", "b",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command error")
	}
	if !strings.Contains(err.Error(), "watch failed") {
		t.Fatalf("expected watch failed error, got %v", err)
	}
}

func TestRootCommandDistillWatchRequiresAllFlags(t *testing.T) {
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		&fakeConfigUIRunner{},
		tools.DistillBatch{},
		tools.DistillWatch{},
		&fakeDistillBatchUseCase{},
		&fakeDistillWatchUseCase{},
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"distill_watch", "--question", "q"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected required flags error")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("expected required flag error, got %v", err)
	}
}
