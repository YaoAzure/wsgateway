package session

import (
	"github.com/samber/do/v2"
)

// Package 定义 Session 包的服务包，使用 Package Loading 模式
var Package = do.Package(
	// Session Builder 使用懒加载，因为它依赖于 Redis 客户端
	// 只有在需要创建 session 时才初始化
	do.Lazy(NewRedisSessionBuilder),
)
