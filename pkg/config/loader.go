package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

const DefaultConfigPath = "./configs/config.yaml"

// Loader handles configuration loading
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	return &Loader{
		configPath: configPath,
	}
}

// Load loads the configuration from the specified file
func (l *Loader) Load() (Config, error) {
	v := viper.New()
	
	// Set config file path
	v.SetConfigFile(l.configPath)
	
	// Set config type based on file extension
	ext := filepath.Ext(l.configPath)
	switch ext {
	case ".yaml", ".yml":
		v.SetConfigType("yaml")
	case ".json":
		v.SetConfigType("json")
	case ".toml":
		v.SetConfigType("toml")
	default:
		v.SetConfigType("yaml") // default to yaml
	}
	
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return config, nil
}

// LoadFromPath is a convenience function to load config from a specific path
func LoadFromPath(configPath string) (Config, error) {
	loader := NewLoader(configPath)
	return loader.Load()
}