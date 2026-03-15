package configui

type DistillSettings struct {
	ProviderName string
	BaseURL      string
	APIKey       string
	Model        string
}

type ProviderOption struct {
	Name           string
	Label          string
	DefaultBaseURL string
	DefaultModel   string
	RequiresAPIKey bool
}

type Repository interface {
	Load(serviceName string) (DistillSettings, error)
	Save(serviceName string, settings DistillSettings) error
}

type Runner interface {
	Run(serviceName string) error
}
