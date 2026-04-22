package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"crac_exam_go/backend/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
)

// CustomFormatter 自定义日志格式
type CustomFormatter struct{}

// Format 格式化日志输出
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 级别（小写）
	level := entry.Level.String()
	b.WriteString(fmt.Sprintf("[%s] ", level))

	// 时间
	b.WriteString(entry.Time.Format("2006-01-02 15:04:05"))
	b.WriteString(" ")

	// 模块名
	if module, ok := entry.Data["module"]; ok {
		b.WriteString(fmt.Sprintf("%s ", module))
		delete(entry.Data, "module")
	}

	// 消息
	b.WriteString(entry.Message)

	// 其他字段（按字母排序）
	if len(entry.Data) > 0 {
		keys := make([]string, 0, len(entry.Data))
		for k := range entry.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		b.WriteString(" ")
		for i, key := range keys {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(fmt.Sprintf("%s=%v", key, entry.Data[key]))
		}
	}

	b.WriteString("\n")
	return b.Bytes(), nil
}

// GetLogger 获取全局日志实例
func GetLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		logger = logrus.New()

		// 设置自定义日志格式
		logger.SetFormatter(&CustomFormatter{})

		// 同时输出到控制台和文件
		logFile := filepath.Join(config.GetLogPath(), "backend.log")
		fileWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10,    // 每个日志文件最大 10MB
			MaxBackups: 7,     // 保留 7 个旧文件
			MaxAge:     30,    // 保留 30 天
			Compress:   false, // 不压缩
		}
		// 使用 MultiWriter 同时输出到控制台和文件
		multiWriter := io.MultiWriter(os.Stdout, fileWriter)
		logger.SetOutput(multiWriter)

		// 设置日志级别
		level := logrus.InfoLevel
		if os.Getenv("APP_ENV") == "development" {
			level = logrus.DebugLevel
		}
		logger.SetLevel(level)
	})

	return logger
}

// Info 记录 info 级别日志
func Info(module string, message string, fields logrus.Fields) {
	entry := GetLogger().WithField("module", module)
	if fields != nil {
		for k, v := range fields {
			entry = entry.WithField(k, v)
		}
	}
	entry.Info(message)
}

// Error 记录 error 级别日志
func Error(module string, message string, err error, fields logrus.Fields) {
	entry := GetLogger().WithField("module", module)
	if fields != nil {
		for k, v := range fields {
			entry = entry.WithField(k, v)
		}
	}
	if err != nil {
		entry.WithError(err).Error(message)
	} else {
		entry.Error(message)
	}
}

// Debug 记录 debug 级别日志
func Debug(module string, message string, fields logrus.Fields) {
	entry := GetLogger().WithField("module", module)
	if fields != nil {
		for k, v := range fields {
			entry = entry.WithField(k, v)
		}
	}
	entry.Debug(message)
}

// Warn 记录 warn 级别日志
func Warn(module string, message string, fields logrus.Fields) {
	entry := GetLogger().WithField("module", module)
	if fields != nil {
		for k, v := range fields {
			entry = entry.WithField(k, v)
		}
	}
	entry.Warn(message)
}
