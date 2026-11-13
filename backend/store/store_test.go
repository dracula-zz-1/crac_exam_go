package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crac_exam_v3/backend/models"
)

func TestOpenPathWithSpacesAndChinese(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "测试 数据")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "crac exam.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}

	// 基本写入与唯一约束
	if err := SaveProgress(db, 1, "A", 3); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if err := SaveProgress(db, 1, "A", 7); err != nil {
		t.Fatalf("save progress 2: %v", err)
	}
	var idx, rows int
	if err := db.QueryRow(`SELECT current_index FROM practice_progress WHERE user_id=1 AND category='A'`).Scan(&idx); err != nil {
		t.Fatalf("query progress: %v", err)
	}
	if idx != 7 {
		t.Fatalf("current_index = %d, want 7", idx)
	}
	db.QueryRow(`SELECT COUNT(*) FROM practice_progress`).Scan(&rows)
	if rows != 1 {
		t.Fatalf("progress rows = %d, want 1", rows)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file missing: %v", err)
	}
}

func TestOpenToleratesLegacyDuplicateJPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// 模拟旧库：删掉唯一索引并写入 J/P/I 重复的数据
	if _, err := db.Exec(`DROP INDEX idx_questions_jpi`); err != nil {
		t.Fatalf("drop index: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := db.Exec(`INSERT INTO questions (J,P,I,Q,T,A,B,C,D,type)
			VALUES ('J','P','I','旧题','A','a','b','c','d',1)`); err != nil {
			t.Fatalf("insert duplicate: %v", err)
		}
	}
	db.Close()

	// 重新打开：唯一索引创建失败应只告警，不应阻止启动
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("open with legacy duplicates should not fail: %v", err)
	}
	defer db2.Close()
	var count int
	if err := db2.QueryRow(`SELECT COUNT(*) FROM questions`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("questions = %d, want 2", count)
	}
	var idx int
	db2.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_questions_jpi'`).Scan(&idx)
	if idx != 0 {
		t.Fatalf("index should not exist with duplicate legacy data")
	}
}

func TestMigrateAddsWrongCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// 模拟旧版本没有 wrong_count 列
	if _, err := db.Exec(`ALTER TABLE error_questions DROP COLUMN wrong_count`); err != nil {
		t.Fatalf("drop column: %v", err)
	}
	db.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	var has int
	if err := db2.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('error_questions') WHERE name='wrong_count'`).Scan(&has); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if has != 1 {
		t.Fatalf("wrong_count column missing after migration")
	}
}

func TestUniqueJPI(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "jpi.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	insert := func(q models.Question) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := InsertQuestions(tx, []models.Question{q}); err != nil {
			return err
		}
		return tx.Commit()
	}

	base := models.Question{J: "J1", P: "P1", I: "I1", Q: "题目", T: "A",
		A: "a", B: "b", C: "c", D: "d", Type: 1}
	if err := insert(base); err != nil {
		t.Fatalf("first insert: %v", err)
	}

	// 同 J/P/I，不同题干：应被联合唯一索引拒绝
	dup := base
	dup.Q = "重复题目"
	if err := insert(dup); err == nil || !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("duplicate J/P/I should fail, got %v", err)
	}

	// J 相同但 P/I 不同：允许
	other := base
	other.P, other.I, other.Q = "P2", "I2", "另一题"
	if err := insert(other); err != nil {
		t.Fatalf("same J with different P/I should be allowed: %v", err)
	}

	exists, err := QuestionExistsByJPI(db, "J1", "P1", "I1", 0)
	if err != nil || !exists {
		t.Fatalf("expected J/P/I exists: %v %v", exists, err)
	}
	exists, err = QuestionExistsByJPI(db, "J1", "P1", "I1", 1)
	if err != nil || exists {
		t.Fatalf("exclude id should not count itself: %v %v", exists, err)
	}
}
