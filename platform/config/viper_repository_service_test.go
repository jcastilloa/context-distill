package config

import (
	"testing"

	"github.com/jcastilloa/context-distill/shared/buildinfo"
	"github.com/spf13/viper"
)

func TestServiceConfigUsesBuildVersion(t *testing.T) {
	previous := buildinfo.Version
	t.Cleanup(func() {
		buildinfo.Version = previous
	})

	buildinfo.Version = "v9.9.9"
	v := viper.New()
	v.Set("service.version", "0.1.0")
	v.Set("service.transport", "stdio")

	repo := &ViperRepository{v: v}
	cfg := repo.ServiceConfig()

	if cfg.Version != "v9.9.9" {
		t.Fatalf("unexpected service version: %q", cfg.Version)
	}
}

func TestServiceConfigFallsBackToDevVersion(t *testing.T) {
	previous := buildinfo.Version
	t.Cleanup(func() {
		buildinfo.Version = previous
	})

	buildinfo.Version = "   "
	repo := &ViperRepository{v: viper.New()}
	cfg := repo.ServiceConfig()

	if cfg.Version != "dev" {
		t.Fatalf("unexpected fallback version: %q", cfg.Version)
	}
}
