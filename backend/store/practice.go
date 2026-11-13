package store

import (
	"database/sql"
	"time"

	"crac_exam_v3/backend/models"
)

// ---------- 练习进度 ----------

func GetProgress(db *sql.DB, userID int64, category string) (int, int64, error) {
	var idx int
	var last int64
	err := db.QueryRow(`SELECT current_index, last_accessed FROM practice_progress
		WHERE user_id = ? AND category = ?`, userID, category).Scan(&idx, &last)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return idx, last, err
}

func SaveProgress(db *sql.DB, userID int64, category string, index int) error {
	_, err := db.Exec(`INSERT INTO practice_progress (user_id, category, current_index, last_accessed)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, category) DO UPDATE SET current_index = excluded.current_index,
			last_accessed = excluded.last_accessed`,
		userID, category, index, time.Now().UnixMilli())
	return err
}

func AllProgress(db *sql.DB, userID int64) (map[string]models.CategoryProgress, error) {
	rows, err := db.Query(`SELECT category, current_index, last_accessed FROM practice_progress WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]models.CategoryProgress{}
	for rows.Next() {
		p := models.CategoryProgress{}
		if err := rows.Scan(&p.Category, &p.Answered, &p.LastAccess); err != nil {
			return nil, err
		}
		out[p.Category] = p
	}
	return out, rows.Err()
}

// ---------- 收藏 ----------

func AddFavorite(db *sql.DB, userID, questionID int64, category string) (bool, error) {
	res, err := db.Exec(`INSERT OR IGNORE INTO favorite_questions (user_id, question_id, category, created_at)
		VALUES (?, ?, ?, ?)`, userID, questionID, category, nowMS())
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func RemoveFavorite(db *sql.DB, userID, questionID int64) error {
	_, err := db.Exec(`DELETE FROM favorite_questions WHERE user_id = ? AND question_id = ?`, userID, questionID)
	return err
}

func IsFavorite(db *sql.DB, userID, questionID int64) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM favorite_questions WHERE user_id = ? AND question_id = ?`,
		userID, questionID).Scan(&n)
	return n > 0, err
}

func FavoriteIDs(db *sql.DB, userID int64, category string) ([]int64, error) {
	query := `SELECT question_id FROM favorite_questions WHERE user_id = ?`
	args := []any{userID}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0, 32)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func FavoriteQuestions(db *sql.DB, userID int64, category string) ([]models.Question, error) {
	query := `SELECT q.id, q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F,
		q.LA, q.LB, q.LC, q.type, q.user_id
		FROM favorite_questions f JOIN questions q ON q.id = f.question_id
		WHERE f.user_id = ?`
	args := []any{userID}
	if category != "" {
		query += ` AND f.category = ?`
		args = append(args, category)
	}
	query += ` ORDER BY f.created_at DESC, f.id DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Question, 0, 32)
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, rows.Err()
}

func CountFavorites(db *sql.DB, userID int64) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM favorite_questions WHERE user_id = ?`, userID).Scan(&n)
	return n, err
}

// ---------- 错题 ----------

// RecordWrong 记录一次答错：首次插入，之后累计 wrong_count。
// 返回值 created 表示是否是第一次进入错题本。
func RecordWrong(db *sql.DB, userID, questionID int64, category string) (bool, int, error) {
	var count int
	err := db.QueryRow(`INSERT INTO error_questions
		(user_id, question_id, category, created_at, wrong_count)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(user_id, question_id, category) DO UPDATE SET
			wrong_count = wrong_count + 1,
			created_at = excluded.created_at
		RETURNING wrong_count`, userID, questionID, category, nowMS()).Scan(&count)
	if err != nil {
		return false, 0, err
	}
	return count == 1, count, nil
}

func AddError(db *sql.DB, userID, questionID int64, category string) (bool, error) {
	created, _, err := RecordWrong(db, userID, questionID, category)
	return created, err
}

// WrongCounts 返回指定用户/类别的每个题目累计答错次数。
func WrongCounts(db *sql.DB, userID int64, category string) (map[int64]int, error) {
	rows, err := db.Query(`SELECT question_id, wrong_count FROM error_questions
		WHERE user_id = ? AND category = ?`, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int{}
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

func RemoveError(db *sql.DB, userID, questionID int64) error {
	_, err := db.Exec(`DELETE FROM error_questions WHERE user_id = ? AND question_id = ?`, userID, questionID)
	return err
}

func ErrorQuestions(db *sql.DB, userID int64, category string) ([]models.Question, error) {
	query := `SELECT q.id, q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F,
		q.LA, q.LB, q.LC, q.type, q.user_id, e.wrong_count
		FROM error_questions e JOIN questions q ON q.id = e.question_id
		WHERE e.user_id = ?`
	args := []any{userID}
	if category != "" {
		query += ` AND e.category = ?`
		args = append(args, category)
	}
	query += ` ORDER BY e.created_at DESC, e.id DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Question, 0, 32)
	for rows.Next() {
		q := &models.Question{}
		if err := rows.Scan(&q.ID, &q.J, &q.P, &q.I, &q.Q, &q.T, &q.A, &q.B, &q.C, &q.D,
			&q.F, &q.LA, &q.LB, &q.LC, &q.Type, &q.UserID, &q.WrongCount); err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, rows.Err()
}

func CountErrors(db *sql.DB, userID int64) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM error_questions WHERE user_id = ?`, userID).Scan(&n)
	return n, err
}

func CountErrorsByCategory(db *sql.DB, userID int64, category string) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM error_questions WHERE user_id = ? AND category = ?`,
		userID, category).Scan(&n)
	return n, err
}

func CountFavoritesByCategory(db *sql.DB, userID int64, category string) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM favorite_questions WHERE user_id = ? AND category = ?`,
		userID, category).Scan(&n)
	return n, err
}

// ---------- 清理 ----------

// ClearUserData 在事务中清空指定用户的全部学习数据。
func ClearUserData(db *sql.DB, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 先删考试记录（题目详情由外键级联删除）
	if _, err := tx.Exec(`DELETE FROM exam_records WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM error_questions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM favorite_questions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM practice_progress WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM random_cycles WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// ClearQuestionBank 清空题库及所有依赖题目的数据；已提交考试的历史快照保留。
func ClearQuestionBank(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM questions`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM error_questions`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM favorite_questions`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM practice_progress`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM random_cycles`); err != nil {
		return err
	}
	// 题库已清空，清除来源/版本信息，但保留自动更新开关。
	if _, err := tx.Exec(`DELETE FROM app_meta WHERE key LIKE 'bank.%'
		AND key NOT IN ('bank.auto_check', 'bank.auto_import')`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM exam_records WHERE status = 0`); err != nil {
		return err
	}
	return tx.Commit()
}

// RemoveAllQuestionsForImport 在导入事务中清空题库与引用数据（保留已提交考试快照）。
func RemoveAllQuestionsForImport(tx *sql.Tx) error {
	stmts := []string{
		`DELETE FROM questions`,
		`DELETE FROM error_questions`,
		`DELETE FROM favorite_questions`,
		`DELETE FROM practice_progress`,
		`DELETE FROM random_cycles`,
		`DELETE FROM exam_records WHERE status = 0`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func CountErrorRecords(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM error_questions`).Scan(&n)
	return n, err
}

func CountFavoriteRecords(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM favorite_questions`).Scan(&n)
	return n, err
}
