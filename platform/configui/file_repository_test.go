package configui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFileRepositorySaveWritesConfigInServiceDirectory(t *testing.T) {
	repo := NewFileRepository()
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

	err = repo.Save("context-distill", DistillSettings{
		ProviderName: "openrouter",
		BaseURL:      "https://openrouter.ai/api/v1",
		APIKey:       "secret-key",
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	configPath := filepath.Join(workspace, ".config", "context-distill", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	cfg := map[string]any{}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("parse saved config: %v", err)
	}

	distill, ok := cfg["distill"].(map[string]any)
	if !ok {
		t.Fatalf("expected distill section in config")
	}
	service, ok := cfg["service"].(map[string]any)
	if !ok {
		t.Fatalf("expected service section in config")
	}

	if distill["provider_name"] != "openrouter" {
		t.Fatalf("unexpected provider: %v", distill["provider_name"])
	}
	if distill["base_url"] != "https://openrouter.ai/api/v1" {
		t.Fatalf("unexpected base_url: %v", distill["base_url"])
	}
	if distill["api_key"] != "secret-key" {
		t.Fatalf("unexpected api_key: %v", distill["api_key"])
	}
	if service["transport"] != "stdio" {
		t.Fatalf("unexpected service.transport: %v", service["transport"])
	}
	if service["version"] != "0.1.0" {
		t.Fatalf("unexpected service.version: %v", service["version"])
	}
}

func TestFileRepositorySavePreservesExistingConfigSections(t *testing.T) {
	repo := NewFileRepository()
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

	targetDir := filepath.Join(workspace, ".config", "context-distill")
	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("create target directory: %v", err)
	}

	existing := `service:
  version: 0.1.0
openai:
  provider_name: openai
distill:
  provider_name: ollama
  model: qwen3.5:2b
`
	if err = os.WriteFile(filepath.Join(targetDir, "config.yaml"), []byte(existing), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	err = repo.Save("context-distill", DistillSettings{
		ProviderName: "ollama",
		BaseURL:      "http://127.0.0.1:11434",
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	serialized := string(content)
	if !strings.Contains(serialized, "service:") {
		t.Fatalf("expected service section to be preserved")
	}
	if !strings.Contains(serialized, "openai:") {
		t.Fatalf("expected openai section to be preserved")
	}
	if !strings.Contains(serialized, "model: qwen3.5:2b") {
		t.Fatalf("expected distill.model to be preserved")
	}
}
