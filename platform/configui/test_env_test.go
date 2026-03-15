package configui

import (
	"os"
	"path/filepath"
	"testing"
)

func configureUserConfigEnv(t *testing.T, workspace string) {
	t.Helper()

	t.Setenv("HOME", workspace)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("APPDATA", filepath.Join(workspace, "AppData", "Roaming"))
}

func serviceConfigDirForTest(t *testing.T, serviceName string) string {
	t.Helper()

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("resolve user config directory: %v", err)
	}

	return filepath.Join(userConfigDir, serviceName)
}

func serviceConfigFileForTest(t *testing.T, serviceName string) string {
	t.Helper()

	return filepath.Join(serviceConfigDirForTest(t, serviceName), "config.yaml")
}
