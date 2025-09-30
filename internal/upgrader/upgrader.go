package upgrader

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"

	"github.com/YaoAzure/wsgateway/pkg/compression"
	"github.com/YaoAzure/wsgateway/pkg/session"
	"github.com/YaoAzure/wsgateway/pkg/jwt"
	"github.com/YaoAzure/wsgateway/pkg/log"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/gobwas/httphead"
)

var (
	ErrInvalidURI       = errors.New("无效的URI")       // URI格式错误或解析失败
	ErrInvalidUserToken = errors.New("无效的UserToken") // JWT token无效、过期或解析失败
	ErrExistedUser      = errors.New("用户已存在")       // 用户已经建立连接，可能是重连或多端登录
)

// Upgrader WebSocket连接升级器
// 负责将HTTP连接升级为WebSocket连接，并处理用户认证、压缩协商、会话管理等功能
type Upgrader struct {
	rdb               redis.Cmdable        // Redis客户端，用于存储和管理用户会话信息
	token             *jwt.UserToken       // JWT token处理器，用于验证和解析用户身份信息
	compressionConfig compression.Config   // 压缩配置，定义WebSocket压缩参数和策略
	sessionBuilder    session.Builder      // 会话构建器，用于创建和管理用户会话
	logger            *log.Logger      // 日志组件，用于记录升级过程中的操作和错误信息
}

func New(i do.Injector) (*Upgrader,error) {
	rdb,err := do.Invoke[redis.Cmdable](i)
	if err!= nil {
		return nil,err
	}	
	token,err := do.Invoke[*jwt.UserToken](i)
	if err!= nil {
		return nil,err
	}
	compressionConfig,err := do.Invoke[compression.Config](i)	
	if err!= nil {
		return nil,err
	}
	sessionBuilder,err := do.Invoke[session.Builder](i)
	if err!= nil {
		return nil,err
	}
	logger,err := do.Invoke[*log.Logger](i)
	if err!= nil {
		return nil,err
	}

	return &Upgrader{
		rdb:               rdb,
		token:             token,
		compressionConfig: compressionConfig,
		sessionBuilder:    sessionBuilder,
		logger:            logger,
	}, nil
}

func (u *Upgrader) Name() string {
	return "gateway.Upgrader"
}

// Upgrade 将HTTP连接升级为WebSocket连接并支持压缩协商
func (u *Upgrader) Upgrade(conn net.Conn) (session.Session, *compression.State, error) {
	var ss session.Session           // 用户会话对象
	var compressionState *compression.State  // 压缩状态对象
	var autoClose bool               // 是否自动关闭连接的标志
	var userInfo session.UserInfo    // 用户信息结构体

	// 只有配置启用时才创建压缩扩展
	// 压缩扩展用于与客户端协商WebSocket压缩参数
	var ext *wsflate.Extension
	if u.compressionConfig.Enabled {
		params := u.compressionConfig.ToParameters()
		ext = &wsflate.Extension{Parameters: params}
		u.logger.Info("压缩扩展已启用", slog.Any("params", params))	
	}
	// 创建WebSocket升级器，配置各种回调函数处理升级过程
	upgrader := ws.Upgrader{
		// Negotiate 压缩协商回调
		// 在WebSocket握手过程中与客户端协商压缩参数
		Negotiate: func(opt httphead.Option) (httphead.Option, error) {
			if ext != nil {
				return ext.Negotiate(opt)  // 执行压缩参数协商
			}
			return httphead.Option{}, nil  // 不启用压缩时返回空选项
		},

		// OnRequest 请求处理回调
		// 在接收到WebSocket升级请求时调用，主要用于用户认证
		OnRequest: func(uri []byte) error {
			var err error
			// 从请求URI中解析用户信息（包含JWT token）
			userInfo, err = u.getUserInfo(string(uri))
			if err != nil {
				u.logger.Error("获取用户信息失败",slog.String("uri", string(uri)),slog.Any("error", err),)
				return fmt.Errorf("%w", err)
			}
			return nil
		},

		// OnHeader HTTP头部处理回调
		// 解析自定义HTTP头部，如X-AutoClose等配置参数
		OnHeader: func(key, value []byte) error {
			// 解析 X-AutoClose header (大小写不敏感)
			// 该头部用于指示连接是否应该自动关闭
			if strings.EqualFold(string(key), "X-AutoClose") {
				autoClose = string(value) == "true"
				u.logger.Warn("解析到AutoClose header",slog.String("key", string(key)),slog.String("value", string(value)),slog.Any("autoClose", autoClose))
			}
			return nil
		},

		// OnBeforeUpgrade 升级前处理回调
		// 在实际升级连接前执行，主要用于创建用户会话
		OnBeforeUpgrade: func() (ws.HandshakeHeader, error) {
			// 在升级前设置autoClose并创建session
			userInfo.AutoClose = autoClose

			// 使用Redis会话构建器创建或获取用户会话
			builder := u.sessionBuilder
			s, isNew, err := builder.Build(context.Background(), userInfo)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			if !isNew {
				// 可能是重连，也可能是多次登录
				// 这种情况下会返回警告但不阻止连接建立
				err = ErrExistedUser
				u.logger.Warn("用户已存在",slog.Any("error", err))
			}
			ss = s
			return ws.HandshakeHeaderString(""), nil  // 返回空的握手头部
		},
	}

	// 执行WebSocket连接升级
	// 这里会触发上面定义的所有回调函数
	_, err := upgrader.Upgrade(conn)
	if err != nil {
		return nil, nil, err
	}

	// 检查压缩协商结果
	// 如果客户端支持压缩且协商成功，则创建压缩状态对象
	if ext != nil {
		if params, accepted := ext.Accepted(); accepted {
			compressionState = &compression.State{
				Enabled:    true,
				Extension:  ext,
				Parameters: params,
			}
			u.logger.Info("压缩协商成功",slog.Any("negotiated_params", params))
		} else {
			u.logger.Warn("压缩协商失败，降级到无压缩模式")
		}
	}
	return ss, compressionState, nil
}

// getUserInfo 从请求URI中解析用户信息
// 该方法负责从WebSocket升级请求的URI中提取JWT token并解析用户身份信息
// 
// URI格式示例: ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
func (u *Upgrader) getUserInfo(uri string) (session.UserInfo, error) {
	// 解析URI字符串，提取查询参数
	uu, err := url.Parse(uri)
	if err != nil {
		return session.UserInfo{}, ErrInvalidURI  // URI格式错误
	}

	// 获取查询参数
	params := uu.Query()
	token := params.Get("token")  // 提取token参数
	
	// 使用JWT处理器解码和验证token
	userClaims, err := u.token.Decode(token)
	if err != nil {
		// token无效、过期或格式错误
		return session.UserInfo{}, fmt.Errorf("%w: %w", ErrInvalidUserToken, err)
	}

	// 构造用户信息对象
	// 注意：AutoClose字段将在OnHeader回调中根据HTTP头部设置
	return session.UserInfo{
		BizID:  userClaims.BizID,   // 业务ID，用于区分不同的业务域
		UserID: userClaims.UserID,  // 用户ID，唯一标识用户
		// AutoClose将在OnHeader回调中设置
	}, nil
}