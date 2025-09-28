package config

// Config represents the application configuration
type Config struct {
	App AppConfig `yaml:"app" mapstructure:"app"`
}

// AppConfig represents the application-specific configuration
type AppConfig struct {
	Name string `yaml:"name" mapstructure:"name"`
	Addr string `yaml:"addr" mapstructure:"addr"`
}