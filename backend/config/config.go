package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	AppName   = "业余无线电模拟考试"
	AppAuthor = "BA4RHH"
	AppDesc   = "业余无线电台操作技术能力验证模拟考试系统"

	// DefaultGitHubRepo 是应用在线更新默认检查的 GitHub 仓库。
	// 也可以通过设置或 CRAC_EXAM_GITHUB_REPO 环境变量覆盖。
	DefaultGitHubRepo = "dracula-zz-1/crac_exam_go"

	// DefaultGiteeRepo 是应用在线更新可选的 Gitee 镜像仓库。
	// 也可以通过设置或 CRAC_EXAM_GITEE_REPO 环境变量覆盖。
	DefaultGiteeRepo = "dracula_zz/crac_exam_go"
)

// AppVersion 当前应用版本。发布构建会通过 -ldflags -X 注入 Git tag，
// 因此这里必须是 var 而不是 const。
var AppVersion = "3.0.0"

// ConfigDir 返回应用配置目录，独立于数据目录，不会被“更换数据目录”影响。
func ConfigDir() string {
	if dir := os.Getenv("CRAC_EXAM_CONFIG_DIR"); dir != "" {
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

// DefaultDataDir 返回默认数据目录（不看用户自定义设置）。
func DefaultDataDir() string {
	if dir := os.Getenv("CRAC_EXAM_DATA_DIR"); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		return dir
	}
	dir := ConfigDir()
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// dataDirPointerFile 记录用户自定义的数据目录。
// 它放在配置目录里，这样即使数据目录被换到别处也仍能找到这条记录。
func dataDirPointerFile() string {
	return filepath.Join(ConfigDir(), "datadir.txt")
}

// CustomDataDir 返回用户自定义的数据目录；未设置或无效时返回空串。
func CustomDataDir() string {
	if os.Getenv("CRAC_EXAM_DATA_DIR") != "" {
		return "" // 环境变量优先级最高，忽略自定义设置
	}
	raw, err := os.ReadFile(dataDirPointerFile())
	if err != nil {
		return ""
	}
	dir := strings.TrimSpace(string(raw))
	if dir == "" {
		return ""
	}
	if _, err := os.Stat(dir); err != nil {
		return ""
	}
	return dir
}

// SetDataDir 记住自定义数据目录（传空串表示恢复默认）。
func SetDataDir(dir string) error {
	dir = strings.TrimSpace(dir)
	path := dataDirPointerFile()
	if dir == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if !filepath.IsAbs(dir) {
		return errors.New("数据目录必须是绝对路径")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(dir), 0o644)
}

// DataDir 返回应用数据目录。优先级：环境变量 > 用户自定义 > 默认目录。
func DataDir() string {
	if dir := os.Getenv("CRAC_EXAM_DATA_DIR"); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		return dir
	}
	if dir := CustomDataDir(); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		return dir
	}
	dir := DefaultDataDir()
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
