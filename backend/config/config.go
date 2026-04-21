package config

import (
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	AppName      string
	Version      string
	DatabasePath string
	LogPath      string
}

var AppConfig *Config

func init() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// 使用用户目录下的 .crac_exam 文件夹
	baseDir := filepath.Join(homeDir, ".crac_exam")

	// 确保目录存在
	os.MkdirAll(baseDir, os.ModePerm)
	os.MkdirAll(filepath.Join(baseDir, "data"), os.ModePerm)
	os.MkdirAll(filepath.Join(baseDir, "logs"), os.ModePerm)

	AppConfig = &Config{
		AppName:      "业余无线电模拟考试系统",
		Version:      "1.0.0",
		DatabasePath: filepath.Join(baseDir, "data", "exam_questions.db"),
		LogPath:      filepath.Join(baseDir, "logs"),
	}
}

// GetDatabasePath 获取数据库文件路径
func GetDatabasePath() string {
	return AppConfig.DatabasePath
}

// GetLogPath 获取日志文件路径
func GetLogPath() string {
	return AppConfig.LogPath
}
