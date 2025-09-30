package types

import (
	"time"

	"github.com/YaoAzure/wsgateway/pkg/session"
)

// Link 表示一个抽象的用户连接，它封装了底层的网络连接（如 WebSocket、TCP），
// 并与一个用户会话 (Session) 绑定。它提供了面向业务的、统一的连接操作接口。
type Link interface {
	// ID 返回此连接的唯一标识符
	ID() string
	// Session 返回与此连接绑定的用户会话信息
	Session() session.Session
	// Send 向客户端异步发送一条消息。
	// 如果发送失败（例如，缓冲区已满或连接已关闭），则返回错误。
	Send(msg []byte) error
	// Receive 返回一个只读通道，用于从客户端接收消息。
	// 调用方可以从该通道中持续读取客户端上行的数据。
	Receive() <-chan []byte
	// Close 主动关闭此连接，并释放相关资源。
	Close() error
	// HasClosed 返回一个只读通道，该通道在连接被关闭时会关闭（手动关闭）。
	// 这是一种非阻塞的、事件驱动的机制，用于监听连接的关闭事件。
	// 例如： `select { case <-link.HasClosed(): ... }`
	// 主要作用是用来让其他组件监听 Link 是否断开
	HasClose() <-chan struct{}
	// UpdateActiveTime 更新连接的最后活跃时间戳。
	// 通常在收到客户端消息或成功发送消息后调用，用于空闲连接检测。
	UpdateActiveTime()
	// TryCloseIfIdle 检查连接是否超过指定的空闲超时时间。
	// 如果已空闲超时，则关闭连接并返回 true；否则返回 false。
	TryCloseIfIdle(timeout time.Duration) bool
}