package jwt

import (
	"github.com/samber/do/v2"
)

// Package 定义 JWT 包的服务包，使用 Package Loading 模式
var Package = do.Package(
	do.Lazy(NewToken),      // JWT Token 服务使用懒加载
	do.Lazy(NewUserToken),  // User JWT Token 服务使用懒加载
)