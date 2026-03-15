package configui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configrepo "github.com/jcastilloa/context-distill/platform/config"

	"gopkg.in/yaml.v3"
)

type FileRepository struct{}

func NewFileRepository() Repository {
	return FileRepository{}
}

func (FileRepository) Load(serviceName string) (DistillSettings, error) {
	repo, err := configrepo.NewForSetup(serviceName)
	if err != nil {
		return DistillSettings{}, err
	}

	cfg := repo.DistillProviderConfig()
	settings := NormalizeSettings(DistillSettings{
		ProviderName: cfg.ProviderName,
		BaseURL:      cfg.BaseURL,
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
	})

	return settings, nil
}

func (FileRepository) Save(serviceName string, settings DistillSettings) error {
	normalized := NormalizeSettings(settings)
	if err := ValidateSettings(normalized); err != nil {
		return err
	}

	targetPath, err := configPath(serviceName)
	if err != nil {
		return err
	}
	cfg, err := readConfigMap(targetPath)
	if err != nil {
		return err
	}

	distillMap := readSection(cfg, "distill")
	distillMap["provider_name"] = normalized.ProviderName
	distillMap["base_url"] = normalized.BaseURL
	if normalized.Model == "" {
		delete(distillMap, "model")
	} else {
		distillMap["model"] = normalized.Model
	}

	if normalized.APIKey == "" {
		delete(distillMap, "api_key")
	} else {
		distillMap["api_key"] = normalized.APIKey
	}

	cfg["distill"] = distillMap
	ensureDefaultServiceSection(cfg)

	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err = os.WriteFile(targetPath, content, 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func ensureDefaultServiceSection(cfg map[string]any) {
	serviceMap := readSection(cfg, "service")
	if _, exists := serviceMap["transport"]; !exists {
		serviceMap["transport"] = "stdio"
	}
	cfg["service"] = serviceMap
}

func configPath(serviceName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", serviceName, "config.yaml"), nil
}

func readConfigMap(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return map[string]any{}, nil
	}

	cfg := map[string]any{}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

func readSection(cfg map[string]any, section string) map[string]any {
	rawSection, exists := cfg[section]
	if !exists {
		return map[string]any{}
	}

	if cast, ok := rawSection.(map[string]any); ok {
		return cast
	}

	// yaml.v3 can unmarshal nested maps as map[interface{}]interface{} in some paths.
	if cast, ok := rawSection.(map[any]any); ok {
		result := make(map[string]any, len(cast))
		for key, value := range cast {
			stringKey, ok := key.(string)
			if ok {
				result[stringKey] = value
			}
		}
		return result
	}

	return map[string]any{}
}
