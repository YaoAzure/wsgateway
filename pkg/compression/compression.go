package compression

import (
	"github.com/gobwas/ws/wsflate"
)

// Config 压缩配置
type Config struct {
	// Enabled 是否启用压缩功能
	Enabled bool `yaml:"enabled"`
	// ServerMaxWindow 服务端最大滑动窗口大小，范围8-15，值越大压缩率越高但内存消耗越大
	ServerMaxWindow int `yaml:"serverMaxWindow"`
	// ClientMaxWindow 客户端最大滑动窗口大小，范围8-15，值越大压缩率越高但内存消耗越大
	ClientMaxWindow int `yaml:"clientMaxWindow"`
	// ServerNoContext 服务端是否禁用上下文接管，true表示每个消息独立压缩
	ServerNoContext bool `yaml:"serverNoContext"`
	// ClientNoContext 客户端是否禁用上下文接管，true表示每个消息独立压缩
	ClientNoContext bool `yaml:"clientNoContext"`
	// Level 压缩级别，范围1-9，1为最快速度，9为最高压缩率
	Level int `yaml:"level"`
}

// ToParameters 将配置转换为wsflate参数
func (c *Config) ToParameters() wsflate.Parameters {
	return wsflate.Parameters{
		ServerMaxWindowBits:     wsflate.WindowBits(c.ServerMaxWindow),
		ClientMaxWindowBits:     wsflate.WindowBits(c.ClientMaxWindow),
		ServerNoContextTakeover: c.ServerNoContext,
		ClientNoContextTakeover: c.ClientNoContext,
	}
}

// State 压缩状态，包含协商后的扩展信息
type State struct {
	// Enabled 压缩是否已启用
	Enabled bool
	// Extension WebSocket压缩扩展实例，用于实际的压缩和解压缩操作
	Extension *wsflate.Extension
	// Parameters 协商后的压缩参数，包含窗口大小和上下文接管设置
	Parameters wsflate.Parameters
}
