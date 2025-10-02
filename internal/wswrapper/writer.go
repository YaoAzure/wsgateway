package wswrapper

import (
	"compress/flate"
	"io"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/gobwas/ws/wsutil"
)

// Writer WebSocket连接写入器
// 封装了WebSocket连接的写入功能，支持压缩和未压缩数据的发送
// 与Reader不同，Writer接受io.Writer接口，提供更灵活的输出目标
type Writer struct {
	writer       *wsutil.Writer          // WebSocket帧写入器，负责构造和发送WebSocket协议帧
	messageState *wsflate.MessageState   // 消息压缩状态管理器，控制是否启用压缩
	flateWriter  *wsflate.Writer         // deflate压缩写入器，用于压缩待发送的数据（仅在压缩模式下使用）
}

// NewServerSideWriter 创建服务端模式的WebSocket写入器
// 用于服务端向客户端发送WebSocket消息，支持可选的数据压缩
func NewServerSideWriter(dest io.Writer, compressed bool) *Writer {
	// 创建并配置消息压缩状态
	messageState := wsflate.MessageState{}
	messageState.SetCompressed(compressed)
	
	// 设置WebSocket状态：服务端模式 + 扩展支持
	state := ws.StateServerSide | ws.StateExtended
	// 使用二进制操作码，适合传输各种类型的数据
	opCode := ws.OpBinary
	
	w := &Writer{
		writer:       wsutil.NewWriter(dest, state, opCode), // 创建底层WebSocket写入器
		messageState: &messageState,
	}
	
	// 如果启用压缩，初始化deflate压缩写入器
	if compressed {
		w.flateWriter = wsflate.NewWriter(nil, func(w io.Writer) wsflate.Compressor {
			// 使用标准库的deflate压缩器，采用默认压缩级别
			f, _ := flate.NewWriter(w, flate.DefaultCompression)
			return f
		})
	}
	
	// 将压缩状态注册到WebSocket写入器的扩展中
	w.writer.SetExtensions(&messageState)
	return w
}

// writeCompressed 写入压缩消息的内部实现
// 使用deflate算法压缩数据后发送，可以显著减少网络传输量
func (w *Writer) writeCompressed(p []byte) (n int, err error) {
	// 重置deflate压缩写入器，将输出目标设置为WebSocket写入器
	w.flateWriter.Reset(w.writer)

	// 将原始数据写入压缩器，数据会被自动压缩
	n, err = w.flateWriter.Write(p)
	if err != nil {
		return 0, err
	}

	// 关闭deflate写入器，这会写入压缩结束标记并完成压缩流
	err = w.flateWriter.Close()
	if err != nil {
		return 0, err
	}

	// 刷新WebSocket写入器，确保压缩后的数据立即通过网络发送
	return n, w.writer.Flush()
}

// writeUncompressed 写入未压缩消息的内部实现
// 直接发送原始数据，适用于已经压缩的数据或不需要压缩的场景
func (w *Writer) writeUncompressed(p []byte) (n int, err error) {
	// 将原始数据直接写入WebSocket写入器，不进行任何压缩处理
	n, err = w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	// 刷新WebSocket写入器，确保数据立即通过网络发送
	return n, w.writer.Flush()
}