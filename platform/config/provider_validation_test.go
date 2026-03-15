package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewReturnsErrorWhenOpenAICompatibleProviderHasNoBaseURL(t *testing.T) {
	withIsolatedProviderEnv(t)
	workspace := t.TempDir()
	configureUserConfigEnv(t, workspace)
	originalDir := mustGetwd(t)
	defer func() { _ = os.Chdir(originalDir) }()
	mustChdir(t, workspace)

	writeConfigFile(t, workspace, []byte(`distill:
  provider_name: openai-compatible
`))

	_, err := New("context-distill")
	if err == nil {
		t.Fatalf("expected validation error when base URL is missing")
	}
}

func TestNewReturnsErrorWhenOpenAIProviderHasNoAPIKey(t *testing.T) {
	withIsolatedProviderEnv(t)
	workspace := t.TempDir()
	configureUserConfigEnv(t, workspace)
	originalDir := mustGetwd(t)
	defer func() { _ = os.Chdir(originalDir) }()
	mustChdir(t, workspace)

	writeConfigFile(t, workspace, []byte(`distill:
  provider_name: openai
  base_url: https://api.openai.com/v1
`))

	_, err := New("context-distill")
	if err == nil {
		t.Fatalf("expected validation error when API key is missing")
	}
}

func TestNewAcceptsOpenAICompatibleProviderWithOpenAIFallbacks(t *testing.T) {
	withIsolatedProviderEnv(t)
	workspace := t.TempDir()
	configureUserConfigEnv(t, workspace)
	originalDir := mustGetwd(t)
	defer func() { _ = os.Chdir(originalDir) }()
	mustChdir(t, workspace)

	writeConfigFile(t, workspace, []byte(`openai:
  base_url: https://openrouter.ai/api/v1
distill:
  provider_name: openai-compatible
`))

	_, err := New("context-distill")
	if err != nil {
		t.Fatalf("expected valid config with openai fallback base URL, got: %v", err)
	}
}

func TestNewRejectsUnsupportedDistillProvider(t *testing.T) {
	withIsolatedProviderEnv(t)
	workspace := t.TempDir()
	configureUserConfigEnv(t, workspace)
	originalDir := mustGetwd(t)
	defer func() { _ = os.Chdir(originalDir) }()
	mustChdir(t, workspace)

	writeConfigFile(t, workspace, []byte(`distill:
  provider_name: custom-llm
`))

	_, err := New("context-distill")
	if err == nil {
		t.Fatalf("expected validation error for unsupported provider")
	}
}

func withIsolatedProviderEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISTILL_PROVIDER_NAME", "")
	t.Setenv("DISTILL_BASE_URL", "")
	t.Setenv("DISTILL_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_API_BASE", "")
	t.Setenv("OPENAI_API_KEY", "")
}

func writeConfigFile(t *testing.T, workspace string, content []byte) {
	t.Helper()
	configPath := filepath.Join(workspace, "config.yaml")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return dir
}

func mustChdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change directory: %v", err)
	}
}
