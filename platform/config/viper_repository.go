package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ViperRepository struct {
	v *viper.Viper
}

const (
	defaultDistillProvider = "ollama"
	defaultDistillBaseURL  = "http://127.0.0.1:11434"
	defaultDistillTimeout  = 90 * time.Second
)

func New(serviceName string) (configDomain.Repository, error) {
	return newRepository(serviceName, true)
}

func NewForSetup(serviceName string) (configDomain.Repository, error) {
	return newRepository(serviceName, false)
}

func newRepository(serviceName string, validate bool) (configDomain.Repository, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if homeConfigDir, ok := serviceHomeConfigDir(serviceName); ok {
		v.AddConfigPath(homeConfigDir)
	}
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	if validate {
		if err := validateDistillProviderConfig(v); err != nil {
			return nil, fmt.Errorf("invalid distill config: %w", err)
		}
	}

	return &ViperRepository{v: v}, nil
}

func serviceHomeConfigDir(serviceName string) (string, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	return filepath.Join(homeDir, ".config", serviceName), true
}

func (r *ViperRepository) OpenAIProviderConfig() aiDomain.ProviderConfig {
	return aiDomain.ProviderConfig{
		APIKey:             r.v.GetString("openai.api_key"),
		BaseURL:            r.v.GetString("openai.base_url"),
		Model:              r.v.GetString("openai.model"),
		ProviderName:       r.v.GetString("openai.provider_name"),
		Timeout:            r.v.GetDuration("openai.timeout"),
		MaxRetries:         r.v.GetInt("openai.max_retries"),
		SupportsSystemRole: r.v.GetBool("openai.supports_system_role"),
		SupportsJSONMode:   r.v.GetBool("openai.supports_json_mode"),
	}
}

func (r *ViperRepository) DistillProviderConfig() aiDomain.ProviderConfig {
	provider := normalizeProviderName(r.v.GetString("distill.provider_name"))
	if provider == "" {
		provider = defaultDistillProvider
	}

	baseURL := strings.TrimSpace(r.v.GetString("distill.base_url"))
	if baseURL == "" {
		baseURL = providerDefaultBaseURL(provider)
	}

	model := strings.TrimSpace(r.v.GetString("distill.model"))
	if model == "" {
		model = defaultModelForProvider(provider)
	}

	timeout := r.v.GetDuration("distill.timeout")
	if timeout <= 0 {
		timeout = defaultDistillTimeout
	}

	return aiDomain.ProviderConfig{
		ProviderName:       provider,
		APIKey:             r.v.GetString("distill.api_key"),
		BaseURL:            baseURL,
		Model:              model,
		Timeout:            timeout,
		MaxRetries:         r.v.GetInt("distill.max_retries"),
		Thinking:           r.v.GetBool("distill.thinking"),
		SupportsSystemRole: r.v.GetBool("distill.supports_system_role"),
		SupportsJSONMode:   r.v.GetBool("distill.supports_json_mode"),
	}
}

func defaultModelForProvider(provider string) string {
	return providerDefaultModel(provider)
}

func (r *ViperRepository) ServiceConfig() configDomain.ServiceConfig {
	version := r.v.GetString("service.version")
	if version == "" {
		version = "0.1.0"
	}

	return configDomain.ServiceConfig{
		Host:      r.v.GetString("service.host"),
		Port:      r.v.GetInt("service.port"),
		APIPrefix: r.v.GetString("service.api_prefix"),
		Version:   version,
		Transport: r.v.GetString("service.transport"),
	}
}
