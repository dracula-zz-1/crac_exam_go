package store

import (
	"database/sql"
	"fmt"
	"strings"

	"crac_exam_v3/backend/models"
)

const questionColumns = `id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanQuestion(s rowScanner) (*models.Question, error) {
	q := &models.Question{}
	err := s.Scan(&q.ID, &q.J, &q.P, &q.I, &q.Q, &q.T, &q.A, &q.B, &q.C, &q.D, &q.F,
		&q.LA, &q.LB, &q.LC, &q.Type, &q.UserID)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func queryQuestions(db *sql.DB, query string, args ...any) ([]models.Question, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Question, 0, 64)
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, rows.Err()
}

func categoryColumn(category string) (string, error) {
	switch strings.ToUpper(category) {
	case "A":
		return "LA", nil
	case "B":
		return "LB", nil
	case "C":
		return "LC", nil
	default:
		return "", fmt.Errorf("无效的类别: %s", category)
	}
}

// QuestionsByCategory 返回某类别全部题目，按 id 稳定排序。
func QuestionsByCategory(db *sql.DB, category string) ([]models.Question, error) {
	col, err := categoryColumn(category)
	if err != nil {
		return nil, err
	}
	return queryQuestions(db, fmt.Sprintf(`SELECT %s FROM questions WHERE %s = 1 ORDER BY id`, questionColumns, col))
}

// RandomQuestions 按类别和题型随机抽题。
func RandomQuestions(db *sql.DB, category string, qType, limit int) ([]models.Question, error) {
	col, err := categoryColumn(category)
	if err != nil {
		return nil, err
	}
	return queryQuestions(db, fmt.Sprintf(
		`SELECT %s FROM questions WHERE %s = 1 AND type = ? ORDER BY RANDOM() LIMIT ?`, questionColumns, col),
		qType, limit)
}

func GetQuestion(db *sql.DB, id int64) (*models.Question, error) {
	row := db.QueryRow(fmt.Sprintf(`SELECT %s FROM questions WHERE id = ?`, questionColumns), id)
	q, err := scanQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

// QuestionIDsByCategory 返回类别内全部题目 id（用于统计）。
func QuestionIDsByCategory(db *sql.DB, category string) ([]int64, error) {
	col, err := categoryColumn(category)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(fmt.Sprintf(`SELECT id FROM questions WHERE %s = 1`, col))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// PageQuestions 题库分页查询。多个类别筛选为 OR 语义。
func PageQuestions(db *sql.DB, page, pageSize int, search string, fa, fb, fc bool) (*models.QuestionPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var conditions []string
	var args []any
	var catConds []string
	if fa {
		catConds = append(catConds, "LA = 1")
	}
	if fb {
		catConds = append(catConds, "LB = 1")
	}
	if fc {
		catConds = append(catConds, "LC = 1")
	}
	if len(catConds) > 0 {
		conditions = append(conditions, "("+strings.Join(catConds, " OR ")+")")
	}
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		conditions = append(conditions, "(Q LIKE ? OR A LIKE ? OR B LIKE ? OR C LIKE ? OR D LIKE ? OR J LIKE ? OR P LIKE ? OR I LIKE ?)")
		for i := 0; i < 8; i++ {
			args = append(args, like)
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM questions"+where, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := db.Query(fmt.Sprintf(
		"SELECT %s FROM questions%s ORDER BY id LIMIT ? OFFSET ?", questionColumns, where), listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]models.Question, 0, pageSize)
	for rows.Next() {
		q, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		data = append(data, *q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	return &models.QuestionPage{
		Data: data, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
	}, nil
}

// QuestionExistsByJPI 检查 J/P/I 联合键是否已被其他题目占用。
func QuestionExistsByJPI(db *sql.DB, j, p, i string, excludeID int64) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM questions
		WHERE J = ? AND P = ? AND I = ? AND id <> ?`, j, p, i, excludeID).Scan(&n)
	return n > 0, err
}

// UpdateQuestion 全字段更新（列名固定，不做动态拼接）。
func UpdateQuestion(db *sql.DB, q *models.Question) error {
	res, err := db.Exec(`UPDATE questions SET J=?, P=?, I=?, Q=?, T=?, A=?, B=?, C=?, D=?, F=?,
		LA=?, LB=?, LC=?, type=? WHERE id=?`,
		q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F, q.LA, q.LB, q.LC, q.Type, q.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("题目不存在")
	}
	return nil
}

// DeleteQuestion 删除题目，并清理该题在错题本/收藏中的引用（考试历史保留快照）。
func DeleteQuestion(db *sql.DB, id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM error_questions WHERE question_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM favorite_questions WHERE question_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM questions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("题目不存在")
	}
	return tx.Commit()
}

// InsertQuestions 在事务中批量写入题目。
func InsertQuestions(tx *sql.Tx, questions []models.Question) error {
	stmt, err := tx.Prepare(`INSERT INTO questions
		(J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range questions {
		q := &questions[i]
		if _, err := stmt.Exec(q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F,
			q.LA, q.LB, q.LC, q.Type, q.UserID); err != nil {
			return fmt.Errorf("写入第 %d 题失败: %w", i+1, err)
		}
	}
	return nil
}

// CategoryCounts 统计三类题目数量及带图数量。
func CategoryCounts(db *sql.DB) ([]models.CategoryCount, error) {
	out := make([]models.CategoryCount, 0, 3)
	for _, cat := range []string{"A", "B", "C"} {
		col, _ := categoryColumn(cat)
		var c models.CategoryCount
		c.Category = cat
		err := db.QueryRow(fmt.Sprintf(`SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN type=1 THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN type=2 THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN F <> '' THEN 1 ELSE 0 END),0)
			FROM questions WHERE %s = 1`, col)).
			Scan(&c.Total, &c.Single, &c.Multiple, &c.WithImg)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func CountQuestions(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM questions`).Scan(&n)
	return n, err
}

func CountImages(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM questions WHERE F <> ''`).Scan(&n)
	return n, err
}

// QuestionCountByCategory 单类别题目总数。
func QuestionCountByCategory(db *sql.DB, category string) (int, error) {
	col, err := categoryColumn(category)
	if err != nil {
		return 0, err
	}
	var n int
	err = db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM questions WHERE %s = 1`, col)).Scan(&n)
	return n, err
}
