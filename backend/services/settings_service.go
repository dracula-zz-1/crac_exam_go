package services

import (
	"bytes"
	"database/sql"
	"encoding/base64"
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

const maxCertificateImageBytes = 32 << 20 // 32 MiB，够 2 倍尺寸的证书 PNG

var pngMagic = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

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
// 用户取消时返回空串且不报错。
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
		if dialogCancelled(err) {
			return "", nil
		}
		return "", fmt.Errorf("打开目录选择框失败：%w", err)
	}
	return dir, nil
}

// SaveCertificateImage 弹出保存对话框，把前端渲染好的电子操作证 PNG 写入用户选择的位置。
// dataURL 可以是 "data:image/png;base64,…" 或纯 base64；用户取消时返回空串且不报错。
func (s *SettingsService) SaveCertificateImage(fileName, dataURL string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("应用尚未就绪")
	}

	raw := strings.TrimSpace(dataURL)
	if strings.HasPrefix(raw, "data:") {
		if i := strings.Index(raw, ","); i >= 0 {
			raw = raw[i+1:]
		}
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("图片数据无效：%w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("图片数据为空")
	}
	if len(data) > maxCertificateImageBytes {
		return "", fmt.Errorf("图片数据过大，已取消保存")
	}
	if !bytes.HasPrefix(data, pngMagic) {
		return "", fmt.Errorf("只支持 PNG 图片")
	}

	saveDialog := app.Dialog.SaveFile()
	saveDialog.SetOptions(&application.SaveFileDialogOptions{
		Title:    "导出电子操作证",
		Filename: certificateFileName(fileName),
	})
	target, err := saveDialog.
		AddFilter("PNG 图片 (*.png)", "*.png").
		PromptForSingleSelection()
	if err != nil {
		if dialogCancelled(err) {
			return "", nil
		}
		return "", fmt.Errorf("打开保存对话框失败：%w", err)
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return "", nil // 用户取消
	}
	if !strings.HasSuffix(strings.ToLower(target), ".png") {
		target += ".png"
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return "", fmt.Errorf("保存图片失败：%w", err)
	}
	utils.Info("SettingsService", "导出电子操作证", "path", target, "bytes", len(data))
	return target, nil
}

func certificateFileName(name string) string {
	name = safeUpdateName(name)
	if !strings.HasSuffix(strings.ToLower(name), ".png") {
		name += ".png"
	}
	return name
}

// dialogCancelled 判断错误是否来自“用户取消对话框”。
// Wails 内部以 "cancelled by user" 表示取消，取消不是错误。
func dialogCancelled(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "cancel")
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
