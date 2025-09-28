package config

// Config represents the application configuration
type Config struct {
	App AppConfig `yaml:"app" mapstructure:"app"`
	JWT JWTConfig `yaml:"jwt" mapstructure:"jwt"`
}

func (c Config) GetAppConfig() AppConfig {
	return c.App
}

func (c Config) GetJWTConfig() JWTConfig {
	return c.JWT
}

// AppConfig represents the application-specific configuration
type AppConfig struct {
	Name string `yaml:"name" mapstructure:"name"`
	Addr string `yaml:"addr" mapstructure:"addr"`
}

type JWTConfig struct {
	Key    string `yaml:"key" mapstructure:"key"`
	Issuer string `yaml:"issuer" mapstructure:"issuer"`
}
