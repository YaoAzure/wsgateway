package config

import (
	"github.com/samber/do/v2"
)

// NewPackage 创建配置包的服务包，需要传入配置实例
// 由于配置需要在启动时加载，所以使用 Eager Loading
func NewPackage(config Config) func(do.Injector) {
	return do.Package(
		do.Eager(config),       // 主配置对象
		do.Eager(config.App),   // App 配置
		do.Eager(config.JWT),   // JWT 配置
		do.Eager(config.Redis), // Redis 配置
		do.Eager(config.Log),   // Log 配置
	)
}
