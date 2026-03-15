package distillation

import (
	"strings"
	"testing"
)

func TestPromptBuilderBuildBatchPrompt(t *testing.T) {
	builder := NewPromptBuilder()
	prompt := builder.BuildBatchPrompt("Did tests pass?", "12 passed")

	if !strings.Contains(prompt, "Question: Did tests pass?") {
		t.Fatalf("question section missing from batch prompt")
	}

	if !strings.Contains(prompt, "Command output:\n12 passed") {
		t.Fatalf("command output section missing from batch prompt")
	}
}

func TestPromptBuilderBuildWatchPrompt(t *testing.T) {
	builder := NewPromptBuilder()
	prompt := builder.BuildWatchPrompt("What changed?", "failed: 0", "failed: 1")

	if !strings.Contains(prompt, "Previous cycle:\nfailed: 0") {
		t.Fatalf("previous cycle section missing from watch prompt")
	}

	if !strings.Contains(prompt, "Current cycle:\nfailed: 1") {
		t.Fatalf("current cycle section missing from watch prompt")
	}
}
