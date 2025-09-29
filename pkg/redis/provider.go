package redis

import (
	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

// Package 定义 Redis 包的服务包，使用 Package Loading 模式
var Package = do.Package(
	// Redis 客户端使用懒加载
	do.Lazy(NewRedisClient),
)

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(i do.Injector) (redis.Cmdable, error) {
	// 从依赖注入容器中获取 Redis 配置
	redisConfig, err := do.Invoke[config.RedisConfig](i)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,
	})

	return rdb, nil
}
