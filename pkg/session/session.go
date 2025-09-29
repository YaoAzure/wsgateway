package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

const (
	// keyFormat 定义了Session在Redis中的存储键格式，设为常量以方便管理和复用。
	keyFormat = "gateway:session:bizId:%d:userId:%d"
)

var (
	_ Session = &redisSession{}

	// ErrSessionExisted 表示尝试创建的Session已经存在。
	ErrSessionExisted = errors.New("session已存在")

	// ErrCreateSessionFailed 表示一个通用的创建失败，通常由底层Redis错误引起。
	ErrCreateSessionFailed = errors.New("创建session失败")

	// ErrDestroySessionFailed 表示销毁Session时发生错误。
	ErrDestroySessionFailed = errors.New("销毁session失败")

	// luaSetSessionIfNotExist 脚本用于原子性地创建Session。
	// 只有当Key不存在时，才会执行HSET操作。
	// 返回1表示创建成功，返回0表示Key已存在。
	// 使用 unpack(ARGV) 需要 Redis 4.0.0+，性能优于循环HSET。
	luaSetSessionIfNotExist = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
    redis.call('HSET', KEYS[1], unpack(ARGV))
    return 1
else
    return 0
end
`)
)

type Session interface {
	// UserInfo 返回当前Session关联的用户身份信息。
	UserInfo() UserInfo
	// Get 从Session中获取一个字段值。
	// 可以通过 判断 errors.Is(err, redis.Nil) ，
	// 来判断是否是 Key 不存在的情况。
	Get(ctx context.Context,key string) (string, error)
	// Set 向Session中设置一个字段键值对。
	Set(ctx context.Context, key, value string) error
	// Destroy 销毁整个Session。
	Destroy(ctx context.Context) error
}

// UserInfo 结构体定义了用户会话信息。
type UserInfo struct {
	BizID int64 `json:"bizId"` // 业务域或者是租户ID
	UserID int64 `json:"userId"` // 用户ID
	AutoClose bool  `json:"autoClose"` // 是否允许空闲时自动关闭连接
}

// redisSession 是 Session 接口的Redis实现。
type redisSession struct {
	userInfo UserInfo
	rdb      redis.Cmdable // Redis客户端的抽象接口
	key      string
}

// newRedisSession 创建一个新的Redis会话实例。
func newRedisSession(userInfo UserInfo, rdb redis.Cmdable) *redisSession {
	return &redisSession{
		userInfo: userInfo,                                              // 保存用户信息
		rdb:      rdb,                                                   // 保存Redis客户端
		key:      fmt.Sprintf(keyFormat, userInfo.BizID, userInfo.UserID), // 根据业务ID和用户ID生成唯一的Redis键
	}
}


// initialize 负责在Redis中实际创建Session。这是一个内部方法。
func (s *redisSession) initialize(ctx context.Context) error {
		// 定义初始Session内容。
	// bizId和userId已在key中，这里不再冗余存储。
	// 使用RFC3339Nano格式存储时间，确保一致性。
	args := []any{
		"loginTime", time.Now().Format(time.RFC3339Nano),
	}
		// 执行Lua脚本
	res, err := luaSetSessionIfNotExist.Run(ctx, s.rdb, []string{s.key}, args...).Result()
	if err != nil {
		// 如果脚本执行出错，包装底层错误。
		return fmt.Errorf("%w: %w", ErrCreateSessionFailed, err)
	}
		
	created, ok := res.(int64)
	if !ok {
		// 正常情况下不会发生，但作为防御性编程，检查脚本返回类型。
		return fmt.Errorf("%w: 未知的脚本结果类型: %T", ErrCreateSessionFailed, res)
	}

	if created != 1 {
		// 如果脚本返回0，说明Session已存在。
		return ErrSessionExisted
	}
	return nil
}

func (s *redisSession) UserInfo() UserInfo { return s.userInfo }

func (s *redisSession) Get(ctx context.Context, key string) (string, error) {
	// 如果没有对应的 key，返回 Redis Nil 错误
	return s.rdb.HGet(ctx, s.key, key).Result()
}

func (s *redisSession) Set(ctx context.Context, key, value string) error {
	// HSet 的 value 其实可以是 any 类型，go-redis会自动处理
	// 但传入结构体时它会被 go-redis 序列化成一种默认的字符串格式，这可能不是你期望的。反序列化时会遇到麻烦
	// 因此这里明确使用string类型，确保数据的可预测性
	// 返回HSet的原始错误，让调用方处理具体的错误情况
	return s.rdb.HSet(ctx, s.key, key, value).Err()
}

func (s *redisSession) Destroy(ctx context.Context) error {
	err := s.rdb.Del(ctx, s.key).Err()
	if err != nil {
		// 包装底层错误，提供更清晰的错误链，便于上层调用者识别错误类型
		return fmt.Errorf("%w: %w", ErrDestroySessionFailed, err)
	}
	return nil
}

type Builder interface {
	// Build 获取或创建一个Session。
	// 无论Session是新创建的还是已存在的，都会返回一个可用的Session实例。
	// 返回的bool值表示Session是否为本次调用新创建的。
	Build(ctx context.Context, info UserInfo) (session Session, isNew bool, err error)
}

// RedisSessionBuilder 是 Builder 接口的Redis实现。
// 负责创建和管理Redis会话实例
type RedisSessionBuilder struct {
	rdb redis.Cmdable // Redis客户端接口，用于执行Redis命令
}

func NewRedisSessionBuilder(i do.Injector) (Builder, error) {
	rdb, err := do.Invoke[redis.Cmdable](i)
	if err != nil {
		return nil, err
	}
	return &RedisSessionBuilder{
		rdb: rdb,
	}, nil
}

// Build 实现 "GetOrCreate" 语义，获取或创建一个会话。
// 如果会话不存在则创建新会话，如果已存在则返回现有会话。
func (r *RedisSessionBuilder) Build(ctx context.Context, userInfo UserInfo) (session Session, isNew bool, err error) {
	s := newRedisSession(userInfo, r.rdb)
	err = s.initialize(ctx)
	switch {
	case err == nil:
		// 没有错误，表示会话是新创建的
		return s, true, nil
	case errors.Is(err, ErrSessionExisted):
		// 如果错误是 ErrSessionExisted，这不是一个失败，返回现有的session实例
		return s, false, nil
	default:
		// 其他所有错误（如redis连接失败、权限错误等）都是真正的失败
		return nil, false, err
	}
}