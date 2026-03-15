package domain

import "time"

type ProviderConfig struct {
	APIKey  string
	BaseURL string
	Model   string

	ProviderName string

	Timeout    time.Duration
	MaxRetries int
	Thinking   bool

	SupportsSystemRole bool
	SupportsJSONMode   bool

	DefaultTemperature *float64
}
