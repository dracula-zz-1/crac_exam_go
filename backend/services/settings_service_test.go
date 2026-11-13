package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"database/sql"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/store"
)

func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCopyDatabaseToProducesReadableCopy(t *testing.T) {
	db := openTempDB(t)
	svc := NewSettingsService(db)

	target := filepath.Join(t.TempDir(), "copy.db")
	if err := svc.copyDatabaseTo(target); err != nil {
		t.Fatalf("copyDatabaseTo: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("copy missing: %v", err)
	}

	copied, err := store.Open(target)
	if err != nil {
		t.Fatalf("open copy: %v", err)
	}
	defer copied.Close()
	var n int
	if err := copied.QueryRow(`SELECT count(*) FROM questions`).Scan(&n); err != nil {
		t.Fatalf("query copy: %v", err)
	}
}

func TestSetDataDirValidatesTarget(t *testing.T) {
	// 用环境变量把默认目录固定到临时目录，避免污染真实配置目录
	base := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", base)

	db := openTempDB(t)
	svc := NewSettingsService(db)

	if _, err := svc.SetDataDir("   ", false); err == nil {
		t.Fatal("空目录应当报错")
	}
	if _, err := svc.SetDataDir(config.DataDir(), false); err == nil {
		t.Fatal("与当前目录相同应当报错")
	}
	if _, err := svc.SetDataDir(filepath.Join(config.DataDir(), "nested"), false); err == nil {
		t.Fatal("当前目录的子目录应当报错")
	}

	okDir := filepath.Join(t.TempDir(), "crac-data")
	if _, err := svc.SetDataDir(okDir, true); err != nil {
		t.Fatalf("SetDataDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(okDir, "crac_exam.db")); err != nil {
		t.Fatalf("数据库未复制到新目录: %v", err)
	}
	// 目标目录已有数据库时，再次复制应当被拒绝
	if _, err := svc.SetDataDir(okDir, true); err == nil {
		t.Fatal("目标已存在数据库应当报错")
	}
	// 仅切换不复制时允许复用已有目录
	if _, err := svc.SetDataDir(okDir, false); err != nil {
		t.Fatalf("仅切换应当成功: %v", err)
	}
}

func TestDataDirPointerRoundTrip(t *testing.T) {
	base := t.TempDir()
	t.Setenv("CRAC_EXAM_CONFIG_DIR", t.TempDir()) // 让记录文件落在临时目录，别碰真实配置

	custom := filepath.Join(base, "elsewhere")
	if err := config.SetDataDir(custom); err != nil {
		t.Fatalf("SetDataDir: %v", err)
	}
	t.Cleanup(func() { _ = config.SetDataDir("") })

	if got := config.CustomDataDir(); got != custom {
		t.Fatalf("CustomDataDir() = %q, want %q", got, custom)
	}
	if got := config.DataDir(); got != custom {
		t.Fatalf("DataDir() = %q, want %q", got, custom)
	}
	if err := config.SetDataDir(""); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if got := config.CustomDataDir(); got != "" {
		t.Fatalf("CustomDataDir() after reset = %q", got)
	}
}

func TestDialogCancelled(t *testing.T) {
	if dialogCancelled(nil) {
		t.Fatal("nil 不应视为取消")
	}
	for _, msg := range []string{"cancelled by user", "Canceled", "user cancelled the dialog"} {
		if !dialogCancelled(errors.New(msg)) {
			t.Errorf("dialogCancelled(%q) = false, want true", msg)
		}
	}
	for _, msg := range []string{"permission denied", "打开目录选择框失败", "disk full"} {
		if dialogCancelled(errors.New(msg)) {
			t.Errorf("dialogCancelled(%q) = true, want false", msg)
		}
	}
}
