package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"crac_exam_v3/backend/importer"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ImportService struct {
	db *sql.DB
	mu sync.Mutex
}

func NewImportService(db *sql.DB) *ImportService {
	return &ImportService{db: db}
}

func (s *ImportService) ServiceName() string { return "ImportService" }

// PickZipFile 打开系统文件选择框，只允许 ZIP。
func (s *ImportService) PickZipFile() (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("应用尚未就绪")
	}
	path, err := app.Dialog.OpenFile().
		SetTitle("选择题库压缩包（ZIP）").
		AddFilter("题库压缩包 (*.zip)", "*.zip").
		PromptForSingleSelection()
	if err != nil {
		if dialogCancelled(err) {
			return "", nil
		}
		return "", fmt.Errorf("打开文件选择框失败：%w", err)
	}
	return path, nil
}

// ImportZip 解析并原子替换题库。只有 ZIP 中的 PDF 会被读取。
func (s *ImportService) ImportZip(zipPath string) (*models.ImportResult, error) {
	zipPath = strings.TrimSpace(zipPath)
	if zipPath == "" {
		return nil, fmt.Errorf("未选择文件")
	}

	// 同一时间只允许一个导入任务
	s.mu.Lock()
	defer s.mu.Unlock()

	utils.Info("ImportService", "开始导入题库", "path", zipPath)
	emitProgress(models.ImportProgress{Stage: "start", Message: "开始导入…", Percent: 1})

	questions, result, err := importer.ImportZip(zipPath, emitProgress)
	if err != nil {
		utils.Error("ImportService", "解析题库失败", err)
		return nil, err
	}

	emitProgress(models.ImportProgress{Stage: "write", Message: "正在写入数据库…", Percent: 90})
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败：%w", err)
	}
	defer tx.Rollback()

	// 先按 J/P/I 快照用户数据，再替换题库，最后映射回新题目。
	refs, err := store.SnapshotImportRefs(tx)
	if err != nil {
		return nil, fmt.Errorf("读取用户数据失败：%w", err)
	}
	if err := store.RemoveAllQuestionsForImport(tx); err != nil {
		return nil, fmt.Errorf("清理旧题库失败：%w", err)
	}
	if err := store.InsertQuestions(tx, questions); err != nil {
		return nil, fmt.Errorf("写入题库失败：%w", err)
	}
	fav, errs, prog, err := store.RestoreImportRefs(tx, refs)
	if err != nil {
		return nil, fmt.Errorf("恢复用户数据失败：%w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交题库失败：%w", err)
	}

	result.PreservedFavorites = fav
	result.PreservedErrors = errs
	result.PreservedProgress = prog
	result.Message = fmt.Sprintf("成功导入 %d 道题目", len(questions))
	result.Imported = len(questions)

	// 记录本次导入来源，供自动更新比对。手动导入先清空远程来源信息，
	// 自动更新流程在导入成功后会回填 URL/ETag/Last-Modified。
	if hash, size, err := fileSHA256Size(zipPath); err == nil {
		now := time.Now()
		values := map[string]string{
			store.MetaBankSHA256:       hash,
			store.MetaBankSize:         strconv.FormatInt(size, 10),
			store.MetaBankSourceName:   filepath.Base(zipPath),
			store.MetaBankSourceURL:    "",
			store.MetaBankETag:         "",
			store.MetaBankLastModified: "",
			store.MetaBankImportedAt:   strconv.FormatInt(now.UnixMilli(), 10),
			store.MetaBankVersion:      filepath.Base(zipPath),
		}
		if err := store.SetMetas(s.db, values); err != nil {
			utils.Warn("ImportService", "写入题库来源信息失败", "error", err.Error())
		}
	}

	emitProgress(models.ImportProgress{Stage: "done", Message: result.Message, Percent: 100})
	utils.Info("ImportService", "题库导入完成", "count", len(questions), "elapsed_ms", result.ElapsedMS,
		"preserved_favorites", fav, "preserved_errors", errs, "preserved_progress", prog)
	return result, nil
}

func fileSHA256Size(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func emitProgress(p models.ImportProgress) {
	app := application.Get()
	if app == nil {
		return
	}
	app.Event.Emit("import:progress", p)
}
