package wswrapper

import (
	"compress/flate"
	"io"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/gobwas/ws/wsutil"
)

// Reader WebSocket连接读取器
// 封装了WebSocket连接的读取功能，支持压缩数据的自动解压缩
// 可以同时用于服务端和客户端模式
type Reader struct {
	conn           net.Conn                    // 底层网络连接
	reader         *wsutil.Reader              // WebSocket帧读取器，负责解析WebSocket协议帧
	controlHandler wsutil.FrameHandlerFunc     // 控制帧处理器，用于处理ping/pong/close等控制帧
	messageState   *wsflate.MessageState       // 消息压缩状态管理器，跟踪压缩相关的状态信息
	flateReader    *wsflate.Reader             // deflate解压缩读取器，用于解压缩接收到的数据
}

// NewServerSideReader 创建服务端模式的WebSocket读取器
// 用于服务端接收和处理客户端发送的WebSocket消息
func NewServerSideReader(conn net.Conn) *Reader {
	// 创建消息压缩状态管理器，用于跟踪压缩相关信息
	messageState := &wsflate.MessageState{}
	// 创建控制帧处理器，设置为服务端模式
	controlHandler := wsutil.ControlFrameHandler(conn, ws.StateServerSide)
	return &Reader{
		conn: conn,
		reader: &wsutil.Reader{
			Source:         conn,                                    // 数据源为网络连接
			State:          ws.StateServerSide | ws.StateExtended,   // 设置为服务端模式并启用扩展支持
			Extensions:     []wsutil.RecvExtension{messageState},    // 注册压缩扩展
			OnIntermediate: controlHandler,                          // 设置控制帧处理回调
		},
		controlHandler: controlHandler,
		messageState:   messageState,
		flateReader: wsflate.NewReader(nil, func(r io.Reader) wsflate.Decompressor {
			return flate.NewReader(r) // 使用标准库的deflate解压缩实现
		}),
	}
}

// NewClientSideReader 创建客户端模式的WebSocket读取器
// 用于客户端接收和处理服务端发送的WebSocket消息
func NewClientSideReader(conn net.Conn) *Reader {
	// 创建消息压缩状态管理器，用于跟踪压缩相关信息
	messageState := &wsflate.MessageState{}
	// 创建控制帧处理器，设置为客户端模式
	controlHandler := wsutil.ControlFrameHandler(conn, ws.StateClientSide)
	return &Reader{
		conn: conn,
		reader: &wsutil.Reader{
			Source:         conn,                                    // 数据源为网络连接
			State:          ws.StateClientSide | ws.StateExtended,   // 设置为客户端模式并启用扩展支持
			Extensions:     []wsutil.RecvExtension{messageState},    // 注册压缩扩展
			OnIntermediate: controlHandler,                          // 设置控制帧处理回调
		},
		controlHandler: controlHandler,
		messageState:   messageState,
		flateReader: wsflate.NewReader(nil, func(r io.Reader) wsflate.Decompressor {
			return flate.NewReader(r) // 使用标准库的deflate解压缩实现
		}),
	}
}

// Read 从WebSocket连接中读取一条完整的消息
// 该方法会自动处理WebSocket协议的各种帧类型，包括控制帧和数据帧
// 对于压缩的数据会自动进行解压缩处理
func (r *Reader) Read() (payload []byte, err error) {
	// 循环读取WebSocket帧，直到获取到数据帧
	for {
		// 读取下一个WebSocket帧的头部信息
		header, err1 := r.reader.NextFrame()
		if err1 != nil {
			return nil, err1
		}

		// 检查是否为控制帧（ping、pong、close等）
		if header.OpCode.IsControl() {
			// 使用控制帧处理器处理控制帧
			if err2 := r.controlHandler(header, r.reader); err2 != nil {
				return nil, err2
			}
			continue // 控制帧处理完毕，继续读取下一帧
		}

		// 处理数据帧：检查消息是否被压缩
		if r.messageState.IsCompressed() {
			// 如果数据被压缩，使用deflate解压缩器进行解压
			r.flateReader.Reset(r.reader)
			return io.ReadAll(r.flateReader)
		}
		// 如果数据未压缩，直接读取原始数据
		return io.ReadAll(r.reader)
	}
}