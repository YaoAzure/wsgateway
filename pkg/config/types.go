package config

// Config represents the application configuration
type Config struct {
	App AppConfig `yaml:"app" mapstructure:"app"`
	JWT JWTConfig `yaml:"jwt" mapstructure:"jwt"`
	Redis RedisConfig `yaml:"redis" mapstructure:"redis"`
}

func (c Config) GetAppConfig() AppConfig {
	return c.App
}

func (c Config) GetJWTConfig() JWTConfig {
	return c.JWT
}

func (c Config) GetRedisConfig() RedisConfig {
	return c.Redis
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

type RedisConfig struct {
	Addr     string `yaml:"addr" mapstructure:"addr"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
	PoolSize int    `yaml:"pool_size" mapstructure:"pool_size"`
}
