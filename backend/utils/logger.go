package utils

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once sync.Once
	log  *slog.Logger
)

// Init 初始化文件日志（带轮转）。已初始化则忽略。
func Init(logDir string) {
	once.Do(func() {
		_ = os.MkdirAll(logDir, 0o755)
		writer := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "app.log"),
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}
		handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})
		log = slog.New(handler)
	})
}

func logger() *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}

func Info(module, msg string, args ...any) {
	logger().Info(msg, append([]any{"module", module}, args...)...)
}

func Warn(module, msg string, args ...any) {
	logger().Warn(msg, append([]any{"module", module}, args...)...)
}

func Error(module, msg string, err error, args ...any) {
	attrs := []any{"module", module}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	attrs = append(attrs, args...)
	logger().Error(msg, attrs...)
}
