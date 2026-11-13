package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type SettingsService struct {
	db *sql.DB
}

func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) ServiceName() string { return "SettingsService" }

func (s *SettingsService) ClearUserData(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("请先登录")
	}
	if err := store.ClearUserData(s.db, userID); err != nil {
		utils.Error("SettingsService", "清空用户数据失败", err, "user_id", userID)
		return fmt.Errorf("清空用户数据失败：%w", err)
	}
	utils.Info("SettingsService", "清空用户数据", "user_id", userID)
	return nil
}

func (s *SettingsService) ClearQuestionBank() error {
	if err := store.ClearQuestionBank(s.db); err != nil {
		utils.Error("SettingsService", "清空题库失败", err)
		return fmt.Errorf("清空题库失败：%w", err)
	}
	utils.Info("SettingsService", "清空题库")
	return nil
}

func (s *SettingsService) GetDBStats() (*models.DBStats, error) {
	stats := &models.DBStats{}
	var err error
	if stats.Users, err = store.CountUsers(s.db); err != nil {
		return nil, err
	}
	if stats.Questions, err = store.CountQuestions(s.db); err != nil {
		return nil, err
	}
	if stats.ExamRecords, err = store.CountExamRecords(s.db); err != nil {
		return nil, err
	}
	if stats.ErrorRecords, err = store.CountErrorRecords(s.db); err != nil {
		return nil, err
	}
	if stats.Favorites, err = store.CountFavoriteRecords(s.db); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *SettingsService) GetAppInfo() *models.AppInfo {
	return &models.AppInfo{
		Name:        config.AppName,
		Version:     config.AppVersion,
		Author:      config.AppAuthor,
		Description: config.AppDesc,
		Copyright:   "Copyright © 2024-2026 " + config.AppAuthor,
		GoVersion:   runtime.Version(),
		DataDir:     config.DataDir(),
	}
}

// PickDataDir 打开系统目录选择框，返回用户选择的数据目录。
func (s *SettingsService) PickDataDir() (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("应用尚未就绪")
	}
	dir, err := app.Dialog.OpenFile().
		SetTitle("选择数据目录").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		SetDirectory(config.DataDir()).
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("打开目录选择框失败：%w", err)
	}
	return dir, nil
}

// SetDataDir 设置数据目录。migrate 为真时把当前数据库复制到新目录。
// 数据目录在启动时读取，因此需要重启应用才会生效。
func (s *SettingsService) SetDataDir(dir string, migrate bool) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("请先选择数据目录")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("数据目录无效")
	}
	current := config.DataDir()
	if currentAbs, err := filepath.Abs(current); err == nil && samePath(currentAbs, abs) {
		return "", fmt.Errorf("新目录与当前数据目录相同")
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("无法创建目录：%w", err)
	}
	// 数据目录不能放在当前数据目录的子目录里，否则容易反复嵌套
	if rel, err := filepath.Rel(current, abs); err == nil &&
		rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("不能选择当前数据目录的子目录")
	}

	target := filepath.Join(abs, "crac_exam.db")
	if migrate {
		if _, statErr := os.Stat(target); statErr == nil {
			return "", fmt.Errorf("目标目录里已有 crac_exam.db，请换一个目录或选择“仅切换”")
		}
		if err := s.copyDatabaseTo(target); err != nil {
			return "", err
		}
	}

	if err := config.SetDataDir(abs); err != nil {
		return "", fmt.Errorf("保存数据目录失败：%w", err)
	}
	utils.Info("SettingsService", "数据目录已更改", "from", current, "to", abs, "migrate", migrate)
	return "数据目录已设为 " + abs + "，重启应用后生效", nil
}

// ResetDataDir 恢复默认数据目录。
func (s *SettingsService) ResetDataDir() (string, error) {
	current := config.DataDir()
	if err := config.SetDataDir(""); err != nil {
		return "", fmt.Errorf("恢复默认数据目录失败：%w", err)
	}
	utils.Info("SettingsService", "数据目录已恢复默认", "from", current, "to", config.DataDir())
	return "已恢复默认数据目录 " + config.DataDir() + "，重启应用后生效", nil
}

// copyDatabaseTo 用 VACUUM INTO 生成一份一致的数据库快照。
func (s *SettingsService) copyDatabaseTo(target string) error {
	if _, err := s.db.Exec(`VACUUM INTO ?`, target); err != nil {
		_ = os.Remove(target)
		return fmt.Errorf("复制现有数据失败：%w", err)
	}
	return nil
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
