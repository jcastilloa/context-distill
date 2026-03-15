package distillation

import "testing"

func TestTextPolicyNormalizeForModel(t *testing.T) {
	policy := NewTextPolicy()
	got := policy.NormalizeForModel("\x1b[31merror\x1b[0m\r\n\r\n\r\nok")

	if got != "error\n\nok" {
		t.Fatalf("unexpected normalized text: %q", got)
	}
}

func TestTextPolicyHasPromptLikeTail(t *testing.T) {
	policy := NewTextPolicy()

	if !policy.HasPromptLikeTail("Continue? [y/N]") {
		t.Fatalf("expected prompt-like tail to be detected")
	}

	if policy.HasPromptLikeTail("plain log line") {
		t.Fatalf("did not expect plain log line to be detected as prompt")
	}
}

func TestTextPolicyStructuralSimilarity(t *testing.T) {
	policy := NewTextPolicy()
	similarity := policy.StructuralSimilarity("watch run\nfailed: 0\n", "watch run\nfailed: 1\n")

	if similarity <= 0.5 {
		t.Fatalf("expected similarity > 0.5, got %f", similarity)
	}
}

func TestTextPolicyLooksLikeBadDistillation(t *testing.T) {
	policy := NewTextPolicy()
	source := makeRepeated("x", 1500)

	if !policy.LooksLikeBadDistillation(source, source) {
		t.Fatalf("expected echoed source to be considered bad distillation")
	}

	if !policy.LooksLikeBadDistillation("input", "Please provide more command output") {
		t.Fatalf("expected prompt-for-more-input answer to be considered bad distillation")
	}
}

func TestTextPolicyEnsureTrailingNewline(t *testing.T) {
	policy := NewTextPolicy()

	if got := policy.EnsureTrailingNewline("ok"); got != "ok\n" {
		t.Fatalf("unexpected output: %q", got)
	}

	if got := policy.EnsureTrailingNewline("ok\n"); got != "ok\n" {
		t.Fatalf("unexpected output with existing newline: %q", got)
	}
}

func makeRepeated(value string, times int) string {
	result := ""
	for i := 0; i < times; i++ {
		result += value
	}
	return result
}
