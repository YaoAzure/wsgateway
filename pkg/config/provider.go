package config


// Provider interface for configuration access
type Provider interface {
	GetConfig() Config
	GetAppConfig() AppConfig
}

// ConfigProvider implements the Provider interface
type ConfigProvider struct {
	config Config
}

// NewConfigProvider creates a new configuration provider
func NewConfigProvider(config Config) Provider {
	return &ConfigProvider{
		config: config,
	}
}

// GetConfig returns the full configuration
func (p *ConfigProvider) GetConfig() Config {
	return p.config
}

// GetAppConfig returns the application configuration
func (p *ConfigProvider) GetAppConfig() AppConfig {
	return p.config.App
}
