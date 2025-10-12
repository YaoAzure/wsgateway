package limiter

import (
	"github.com/samber/do/v2"
)

// Package 定义 JWT 包的服务包，使用 Package Loading 模式
var Package = do.Package(
	do.Lazy(NewTokenLimiter),
)