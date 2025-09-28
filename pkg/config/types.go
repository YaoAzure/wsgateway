package config

import (
	"github.com/samber/do/v2"
)

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

func NewAppConfig(i do.Injector) (AppConfig, error) {
	conf := do.MustInvoke[Config](i)
	return conf.GetAppConfig(), nil
}

type JWTConfig struct {
	Key    string `yaml:"key" mapstructure:"key"`
	Issuer string `yaml:"issuer" mapstructure:"issuer"`
}

func NewJWTConfig(i do.Injector) (JWTConfig, error) {
	conf := do.MustInvoke[Config](i)
	return conf.GetJWTConfig(), nil
}
