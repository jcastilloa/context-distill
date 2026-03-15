package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPrefersDotConfigServiceDirectory(t *testing.T) {
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

	dotConfigDir := filepath.Join(workspace, ".config", "context-distill")
	if err = os.MkdirAll(dotConfigDir, 0o755); err != nil {
		t.Fatalf("create dot config dir: %v", err)
	}

	dotConfigContent := []byte("distill:\n  provider_name: openai-compatible\n  base_url: http://127.0.0.1:9000/v1\n  model: from-dot-config\n")
	if err = os.WriteFile(filepath.Join(dotConfigDir, "config.yaml"), dotConfigContent, 0o644); err != nil {
		t.Fatalf("write dot config: %v", err)
	}

	rootConfigContent := []byte("distill:\n  provider_name: ollama\n  model: from-root\n")
	if err = os.WriteFile(filepath.Join(workspace, "config.yaml"), rootConfigContent, 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	repo, err := New("context-distill")
	if err != nil {
		t.Fatalf("create config repository: %v", err)
	}

	cfg := repo.DistillProviderConfig()
	if cfg.Model != "from-dot-config" {
		t.Fatalf("expected model from $HOME/.config/<service>, got %q", cfg.Model)
	}
	if cfg.ProviderName != "openai-compatible" {
		t.Fatalf("expected provider from $HOME/.config/<service>, got %q", cfg.ProviderName)
	}
}

func TestNewFallsBackToCurrentDirectoryConfig(t *testing.T) {
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

	rootConfigContent := []byte("distill:\n  provider_name: openai-compatible\n  base_url: http://127.0.0.1:9000/v1\n  model: from-root\n")
	if err = os.WriteFile(filepath.Join(workspace, "config.yaml"), rootConfigContent, 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	repo, err := New("context-distill")
	if err != nil {
		t.Fatalf("create config repository: %v", err)
	}

	cfg := repo.DistillProviderConfig()
	if cfg.Model != "from-root" {
		t.Fatalf("expected model from current directory config.yaml, got %q", cfg.Model)
	}
}
