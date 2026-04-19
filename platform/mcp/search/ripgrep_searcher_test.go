package search

import (
	"context"
	"errors"
	"strings"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
)

type fakeCommandRunner struct {
	name    string
	args    []string
	workDir string
	output  []byte
	err     error
}

func (f *fakeCommandRunner) Run(_ context.Context, workDir, name string, args []string) ([]byte, error) {
	f.workDir = workDir
	f.name = name
	f.args = append([]string(nil), args...)
	return f.output, f.err
}

type fakeFileReader struct {
	files map[string]string
}

func (f *fakeFileReader) ReadFile(path string) ([]byte, error) {
	content, ok := f.files[path]
	if !ok {
		return nil, errors.New("file not found")
	}
	return []byte(content), nil
}

func TestRipgrepSearcherSearchTextReturnsNormalizedMatches(t *testing.T) {
	runner := &fakeCommandRunner{output: []byte(strings.Join([]string{
		`{"type":"match","data":{"path":{"text":"platform/config/load.go"},"lines":{"text":"provider_name := cfg.ProviderName\\n"},"line_number":27}}`,
		`{"type":"match","data":{"path":{"text":"platform/config/load.go"},"lines":{"text":"provider_name := cfg.ProviderName\\n"},"line_number":27}}`,
	}, "\n"))}

	searcher := NewRipgrepSearcher("/repo")
	searcher.Runner = runner
	searcher.Reader = &fakeFileReader{files: map[string]string{}}

	matches, err := searcher.Search(context.Background(), distillapp.SearchCodeRequest{
		Query:        "provider_name",
		Mode:         distillapp.SearchModeText,
		Question:     "q",
		ContextLines: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 raw matches from parser, got %d", len(matches))
	}
	if matches[0].File != "platform/config/load.go" || matches[0].Line != 27 {
		t.Fatalf("unexpected match: %#v", matches[0])
	}
	if matches[0].Kind != distillapp.SearchMatchKindMatch {
		t.Fatalf("unexpected kind: %q", matches[0].Kind)
	}
	if runner.name != "rg" {
		t.Fatalf("expected rg command, got %q", runner.name)
	}
	if !containsArg(runner.args, "--fixed-strings") {
		t.Fatalf("expected fixed strings mode args, got %v", runner.args)
	}
}

func TestRipgrepSearcherSearchPathReturnsFilteredFiles(t *testing.T) {
	runner := &fakeCommandRunner{output: []byte("cmd/server/main.go\nplatform/config/load.go\nREADME.md\n")}

	searcher := NewRipgrepSearcher("/repo")
	searcher.Runner = runner
	searcher.Reader = &fakeFileReader{files: map[string]string{}}

	matches, err := searcher.Search(context.Background(), distillapp.SearchCodeRequest{
		Query:    "config",
		Mode:     distillapp.SearchModePath,
		Question: "q",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 path match, got %d", len(matches))
	}
	if matches[0].File != "platform/config/load.go" {
		t.Fatalf("unexpected file match: %#v", matches[0])
	}
	if matches[0].Kind != distillapp.SearchMatchKindPath {
		t.Fatalf("expected path kind, got %q", matches[0].Kind)
	}
}

func TestRipgrepSearcherSearchSymbolClassifiesDefinitionAndUsage(t *testing.T) {
	runner := &fakeCommandRunner{output: []byte(strings.Join([]string{
		`{"type":"match","data":{"path":{"text":"a.go"},"lines":{"text":"func LoadDistillConfig() error {\\n"},"line_number":10}}`,
		`{"type":"match","data":{"path":{"text":"b.go"},"lines":{"text":"_ = LoadDistillConfig()\\n"},"line_number":20}}`,
	}, "\n"))}

	searcher := NewRipgrepSearcher("/repo")
	searcher.Runner = runner
	searcher.Reader = &fakeFileReader{files: map[string]string{}}

	matches, err := searcher.Search(context.Background(), distillapp.SearchCodeRequest{
		Query:    "LoadDistillConfig",
		Mode:     distillapp.SearchModeSymbol,
		Question: "q",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 symbol matches, got %d", len(matches))
	}
	if matches[0].Kind != distillapp.SearchMatchKindDefinition {
		t.Fatalf("expected definition, got %q", matches[0].Kind)
	}
	if matches[1].Kind != distillapp.SearchMatchKindUsage {
		t.Fatalf("expected usage, got %q", matches[1].Kind)
	}
}

func TestRipgrepSearcherSearchReturnsNoMatchesWhenRunnerReturnsExitCodeOne(t *testing.T) {
	runner := &fakeCommandRunner{err: &fakeExitError{exitCode: 1}}

	searcher := NewRipgrepSearcher("/repo")
	searcher.Runner = runner
	searcher.Reader = &fakeFileReader{files: map[string]string{}}

	matches, err := searcher.Search(context.Background(), distillapp.SearchCodeRequest{
		Query:    "none",
		Mode:     distillapp.SearchModeText,
		Question: "q",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected empty matches, got %d", len(matches))
	}
}

func containsArg(args []string, expected string) bool {
	for _, arg := range args {
		if arg == expected {
			return true
		}
	}
	return false
}

// fakeExitError emulates exec.ExitError behavior needed by no-match detection.
type fakeExitError struct {
	exitCode int
}

func (f *fakeExitError) Error() string {
	return "exit"
}

func (f *fakeExitError) ExitCode() int {
	return f.exitCode
}
