package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewForSetupSkipsDistillValidation(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err = os.Chdir(workspace); err != nil {
		t.Fatalf("change directory: %v", err)
	}

	configContent := []byte(`distill:
  provider_name: openai
  base_url: https://api.openai.com/v1
`)
	if err = os.WriteFile(filepath.Join(workspace, "config.yaml"), configContent, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if _, err = New("context-distill"); err == nil {
		t.Fatalf("expected strict New to fail due missing api_key")
	}

	if _, err = NewForSetup("context-distill"); err != nil {
		t.Fatalf("expected NewForSetup to skip validation, got: %v", err)
	}
}
