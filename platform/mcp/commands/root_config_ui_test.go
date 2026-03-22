package commands

import (
	"errors"
	"testing"

	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"
)

type fakeConfigUIRunner struct {
	calledWith string
	err        error
}

func (r *fakeConfigUIRunner) Run(serviceName string) error {
	r.calledWith = serviceName
	return r.err
}

func TestRootCommandRunsConfigUIWhenFlagEnabled(t *testing.T) {
	uiRunner := &fakeConfigUIRunner{}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		uiRunner,
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"--config-ui"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uiRunner.calledWith != "context-distill" {
		t.Fatalf("expected config UI to run for context-distill, got %q", uiRunner.calledWith)
	}
}

func TestRootCommandReturnsConfigUIError(t *testing.T) {
	uiRunner := &fakeConfigUIRunner{err: errors.New("ui failed")}
	runner := NewRunner(
		"context-distill",
		configDomain.ServiceConfig{Version: "test", Transport: "stdio"},
		nil,
		uiRunner,
		tools.DistillBatch{},
		tools.DistillWatch{},
		nil,
		nil,
	)

	cmd := runner.newRootCommand()
	cmd.SetArgs([]string{"--config-ui"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected config UI error")
	}
}
