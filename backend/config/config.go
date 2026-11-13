package config

import (
	"os"
	"path/filepath"
)

const (
	AppName   = "业余无线电模拟考试"
	AppAuthor = "BA4RHH"
	AppDesc   = "业余无线电台操作技术能力验证模拟考试系统"

	// DefaultGitHubRepo 是应用在线更新默认检查的 GitHub 仓库。
	// 也可以通过设置或 CRAC_EXAM_GITHUB_REPO 环境变量覆盖。
	DefaultGitHubRepo = "dracula-zz-1/crac_exam_go"
)

// AppVersion 当前应用版本。发布构建会通过 -ldflags -X 注入 Git tag，
// 因此这里必须是 var 而不是 const。
var AppVersion = "3.0.0"

// DataDir 返回应用数据目录。优先使用环境变量，便于测试与便携部署。
func DataDir() string {
	if dir := os.Getenv("CRAC_EXAM_DATA_DIR"); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		return dir
	}
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		exe, exeErr := os.Executable()
		if exeErr == nil {
			base = filepath.Dir(exe)
		} else {
			base = "."
		}
	}
	dir := filepath.Join(base, "CracExamV3")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func DBPath() string {
	return filepath.Join(DataDir(), "crac_exam.db")
}

func LogDir() string {
	dir := filepath.Join(DataDir(), "logs")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func ImportTempDir() string {
	dir := filepath.Join(DataDir(), "temp")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// UpdatesDir 保存从 GitHub Release 下载的应用更新包。
func UpdatesDir() string {
	dir := filepath.Join(DataDir(), "updates")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// ExamConfig 考试配置。
type ExamConfig struct {
	Total       int `json:"total"`
	Single      int `json:"single"`
	Multiple    int `json:"multiple"`
	TimeMinutes int `json:"timeMinutes"`
	PassScore   int `json:"passScore"`
}

var ExamConfigs = map[string]ExamConfig{
	"A": {Total: 40, Single: 32, Multiple: 8, TimeMinutes: 40, PassScore: 30},
	"B": {Total: 60, Single: 45, Multiple: 15, TimeMinutes: 60, PassScore: 45},
	"C": {Total: 90, Single: 70, Multiple: 20, TimeMinutes: 90, PassScore: 70},
}

func ExamConfigFor(category string) (ExamConfig, bool) {
	cfg, ok := ExamConfigs[category]
	return cfg, ok
}

func Categories() []string { return []string{"A", "B", "C"} }
