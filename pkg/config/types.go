package config

// Config represents the application configuration
type Config struct {
	App    AppConfig    `yaml:"app" mapstructure:"app"`
	JWT    JWTConfig    `yaml:"jwt" mapstructure:"jwt"`
	Redis  RedisConfig  `yaml:"redis" mapstructure:"redis"`
	Log    LogConfig    `yaml:"log" mapstructure:"log"`
	Server ServerConfig `yaml:"server" mapstructure:"server"`
	Link   LinkConfig   `yaml:"link" mapstructure:"link"`
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

type ServerConfig struct {
	Websocket WebsocketConfig `yaml:"websocket" mapstructure:"websocket"`
}

type LinkConfig struct {
	Timeout     TimeoutConfig     `yaml:"timeout" mapstructure:"timeout"`
	Buffer      BufferConfig      `yaml:"buffer" mapstructure:"buffer"`
	RetryStrategy RetryStrategyConfig `yaml:"retryStrategy" mapstructure:"retryStrategy"`
	Limit       LimitConfig       `yaml:"limit" mapstructure:"limit"`
	EventHandler EventHandlerConfig `yaml:"eventHandler" mapstructure:"eventHandler"`
}

type WebsocketConfig struct {
	Host        string            `yaml:"host" mapstructure:"host"`
	Port        int               `yaml:"port" mapstructure:"port"`
	Compression CompressionConfig `yaml:"compression" mapstructure:"compression"`
	TokenLimiter TokenLimiterConfig `yaml:"tokenLimiter" mapstructure:"tokenLimiter"`
}

type CompressionConfig struct {
	Enabled         bool `yaml:"enabled" mapstructure:"enabled"`
	ServerMaxWindow int  `yaml:"serverMaxWindow" mapstructure:"serverMaxWindow"`
	ClientMaxWindow int  `yaml:"clientMaxWindow" mapstructure:"clientMaxWindow"`
	ServerNoContext bool `yaml:"serverNoContext" mapstructure:"serverNoContext"`
	ClientNoContext bool `yaml:"clientNoContext" mapstructure:"clientNoContext"`
	Level           int  `yaml:"level" mapstructure:"level"`
}

type TokenLimiterConfig struct {
	InitialCapacity  int64 `yaml:"initialCapacity" mapstructure:"initialCapacity"`
	MaxCapacity      int64 `yaml:"maxCapacity" mapstructure:"maxCapacity"`
	IncreaseStep     int64 `yaml:"increaseStep" mapstructure:"increaseStep"`
	IncreaseInterval int64 `yaml:"increaseInterval" mapstructure:"increaseInterval"`
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



type TimeoutConfig struct {
	Read  int64 `yaml:"read" mapstructure:"read"`
	Write int64 `yaml:"write" mapstructure:"write"`
}

type BufferConfig struct {
	ReceiveBufferSize int `yaml:"receiveBufferSize" mapstructure:"receiveBufferSize"`
	SendBufferSize    int `yaml:"sendBufferSize" mapstructure:"sendBufferSize"`
}

type RetryStrategyConfig struct {
	InitInterval int64 `yaml:"initInterval" mapstructure:"initInterval"`
	MaxInterval  int64 `yaml:"maxInterval" mapstructure:"maxInterval"`
	MaxRetries   int   `yaml:"maxRetries" mapstructure:"maxRetries"`
}

type LimitConfig struct {
	Rate int `yaml:"rate" mapstructure:"rate"`
}

type EventHandlerConfig struct {
	RequestTimeout int64             `yaml:"requestTimeout" mapstructure:"requestTimeout"`
	RetryStrategy  RetryStrategyConfig `yaml:"retryStrategy" mapstructure:"retryStrategy"`
	PushMessage    PushMessageConfig `yaml:"pushMessage" mapstructure:"pushMessage"`
}

type PushMessageConfig struct {
	RetryInterval int64 `yaml:"retryInterval" mapstructure:"retryInterval"`
	MaxRetries    int   `yaml:"maxRetries" mapstructure:"maxRetries"`
}
