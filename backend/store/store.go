package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"crac_exam_v3/backend/utils"

	_ "modernc.org/sqlite"
)

// Open 打开（必要时创建）数据库，并启用 WAL、忙等待与外键。
func Open(path string) (*sql.DB, error) {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p}
	dsn := u.String() + "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	// SQLite 单写者：串行化连接可彻底避免 database is locked。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			id_card TEXT NOT NULL UNIQUE,
			last_login INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			J TEXT NOT NULL DEFAULT '',
			P TEXT NOT NULL DEFAULT '',
			I TEXT NOT NULL DEFAULT '',
			Q TEXT NOT NULL DEFAULT '',
			T TEXT NOT NULL DEFAULT '',
			A TEXT NOT NULL DEFAULT '',
			B TEXT NOT NULL DEFAULT '',
			C TEXT NOT NULL DEFAULT '',
			D TEXT NOT NULL DEFAULT '',
			F TEXT NOT NULL DEFAULT '',
			LA INTEGER NOT NULL DEFAULT 0,
			LB INTEGER NOT NULL DEFAULT 0,
			LC INTEGER NOT NULL DEFAULT 0,
			type INTEGER NOT NULL DEFAULT 1,
			user_id INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_questions_la ON questions(LA, type)`,
		`CREATE INDEX IF NOT EXISTS idx_questions_lb ON questions(LB, type)`,
		`CREATE INDEX IF NOT EXISTS idx_questions_lc ON questions(LC, type)`,
		`CREATE TABLE IF NOT EXISTS exam_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT NOT NULL,
			exam_date INTEGER NOT NULL,
			started_at INTEGER NOT NULL,
			submitted_at INTEGER NOT NULL DEFAULT 0,
			duration_seconds INTEGER NOT NULL DEFAULT 0,
			user_id INTEGER NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			total_questions INTEGER NOT NULL DEFAULT 0,
			correct_count INTEGER NOT NULL DEFAULT 0,
			pass_score INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_exam_user_status ON exam_records(user_id, status, exam_date)`,
		`CREATE TABLE IF NOT EXISTS exam_question_details (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			exam_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			seq INTEGER NOT NULL,
			question_text TEXT NOT NULL DEFAULT '',
			option_a TEXT NOT NULL DEFAULT '',
			option_b TEXT NOT NULL DEFAULT '',
			option_c TEXT NOT NULL DEFAULT '',
			option_d TEXT NOT NULL DEFAULT '',
			correct_answer TEXT NOT NULL DEFAULT '',
			user_answer TEXT NOT NULL DEFAULT '',
			is_correct INTEGER NOT NULL DEFAULT 0,
			type INTEGER NOT NULL DEFAULT 1,
			image_data TEXT NOT NULL DEFAULT '',
			UNIQUE(exam_id, seq),
			FOREIGN KEY(exam_id) REFERENCES exam_records(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_details_exam ON exam_question_details(exam_id, seq)`,
		`CREATE TABLE IF NOT EXISTS error_questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			category TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			wrong_count INTEGER NOT NULL DEFAULT 1,
			UNIQUE(user_id, question_id, category)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_error_user ON error_questions(user_id, category)`,
		`CREATE TABLE IF NOT EXISTS favorite_questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			category TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			UNIQUE(user_id, question_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_favorite_user ON favorite_questions(user_id, category)`,
		`CREATE TABLE IF NOT EXISTS practice_progress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category TEXT NOT NULL,
			current_index INTEGER NOT NULL DEFAULT 0,
			last_accessed INTEGER NOT NULL DEFAULT 0,
			UNIQUE(user_id, category)
		)`,
		`CREATE TABLE IF NOT EXISTS app_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS random_cycles (
			user_id INTEGER NOT NULL,
			category TEXT NOT NULL,
			cycle_no INTEGER NOT NULL DEFAULT 1,
			queue TEXT NOT NULL DEFAULT '[]',
			position INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY(user_id, category)
		)`,
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("初始化数据库表失败: %w", err)
		}
	}

	// 旧库补列：错题累计次数（用于随机练习权重）。
	if _, err := db.Exec(`ALTER TABLE error_questions ADD COLUMN wrong_count INTEGER NOT NULL DEFAULT 1`); err != nil &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		utils.Warn("store", "添加 wrong_count 列失败", "error", err.Error())
	}

	// 题目自然唯一键：J + P + I。数据库内部仍保留自增 id 作为引用键
	// （考试详情、错题、收藏都引用 id），二者职责分离。
	// 旧库若已存在联合键重复，无法创建唯一索引；此时记录警告，应用仍可启动，
	// 导入与编辑路径也会主动去重/校验。
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_jpi ON questions(J, P, I)`); err != nil {
		utils.Warn("store", "创建 J/P/I 联合唯一索引失败（数据库中可能存在重复题目）", "error", err.Error())
	}
	return nil
}

func nowMS() int64 { return time.Now().UnixMilli() }

func toTime(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func normalizeAnswer(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	seen := [4]bool{}
	var b strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'D' && !seen[r-'A'] {
			seen[r-'A'] = true
			b.WriteRune(r)
		}
	}
	// A-D 固定排序
	out := []byte(b.String())
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return string(out)
}

// CleanupAbandonedExams 清理上次运行遗留的“进行中”考试（启动时调用）。
func CleanupAbandonedExams(db *sql.DB) {
	res, err := db.Exec(`DELETE FROM exam_records WHERE status = 0`)
	if err != nil {
		utils.Error("store", "清理未完成考试失败", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		utils.Info("store", "已清理未完成考试", "count", n)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
