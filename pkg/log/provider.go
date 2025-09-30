package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/samber/do/v2"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger = slog.Logger

var Package = do.Package(
	do.Lazy(NewLogger),
)

func NewLogger(i do.Injector) (*Logger, error) {
	logConfig, err := do.Invoke[config.LogConfig](i)
	if err != nil {
		return nil, err
	}

	// 1. 设置日志级别
	var level slog.Level
	switch logConfig.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 2. 设置输出位置 (Writer)
	var writer io.Writer
	fileWriter := &lumberjack.Logger{
		Filename:   logConfig.Output.Path,
		MaxSize:    logConfig.Rotation.MaxSize,
		MaxBackups: logConfig.Rotation.MaxBackups,
		MaxAge:     logConfig.Rotation.MaxAge,
		Compress:   logConfig.Rotation.Compress,
	}

	switch logConfig.Output.Type {
	case "file":
		writer = fileWriter
	case "console":
		writer = os.Stdout
	case "multi":
		writer = io.MultiWriter(os.Stdout, fileWriter)
	default:
		writer = os.Stdout
	}

	// 3. 创建 Handler
	handlerOpts := &slog.HandlerOptions{
		AddSource: logConfig.ShowCaller,
		Level:     level,
	}

	var handler slog.Handler
	if logConfig.Format == "json" {
		handler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		handler = slog.NewTextHandler(writer, handlerOpts)
	}

	// 4. 添加全局字段
	logger := slog.New(handler)
	if len(logConfig.Fields) > 0 {
		attrs := make([]any, 0, len(logConfig.Fields)*2)
		for _, field := range logConfig.Fields {
			attrs = append(attrs, field.Key, field.Value)
		}
		logger = logger.With(attrs...)
	}

	return logger, nil
}
