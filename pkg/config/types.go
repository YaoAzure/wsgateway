package config

// Config represents the application configuration
type Config struct {
	App   AppConfig   `yaml:"app" mapstructure:"app"`
	JWT   JWTConfig   `yaml:"jwt" mapstructure:"jwt"`
	Redis RedisConfig `yaml:"redis" mapstructure:"redis"`
	Log   LogConfig   `yaml:"log" mapstructure:"log"`
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

type LogConfig struct {
	Level      string         `yaml:"level" mapstructure:"level"`
	Format     string         `yaml:"format" mapstructure:"format"`
	ShowCaller bool           `yaml:"show_caller" mapstructure:"show_caller"`
	Output     OutputConfig   `yaml:"output" mapstructure:"output"`
	Rotation   RotationConfig `yaml:"rotation" mapstructure:"rotation"`
	Fields     []FieldConfig  `yaml:"fields" mapstructure:"fields"`
}

// FieldConfig represents a key-value pair for log fields
type FieldConfig struct {
	Key   string `yaml:"key" mapstructure:"key"`
	Value string `yaml:"value" mapstructure:"value"`
}

// GetFieldsMap converts the Fields slice to a map for easier access
func (lc *LogConfig) GetFieldsMap() map[string]string {
	fieldsMap := make(map[string]string)
	for _, field := range lc.Fields {
		fieldsMap[field.Key] = field.Value
	}
	return fieldsMap
}

type OutputConfig struct {
	Type string `yaml:"type" mapstructure:"type"`
	Path string `yaml:"path" mapstructure:"path"`
}

type RotationConfig struct {
	MaxSize    int  `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int  `yaml:"max_age" mapstructure:"max_age"`
	MaxBackups int  `yaml:"max_backups" mapstructure:"max_backups"`
	Compress   bool `yaml:"compress" mapstructure:"compress"`
}
