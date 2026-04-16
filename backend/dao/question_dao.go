package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"fmt"
)

// QuestionDAO 题目数据访问对象
type QuestionDAO struct {
	*BaseDAO
}

// NewQuestionDAO 创建 QuestionDAO 实例
func NewQuestionDAO(db *sql.DB) *QuestionDAO {
	return &QuestionDAO{
		BaseDAO: NewBaseDAO(db, "questions"),
	}
}

// Create 创建题目
func (dao *QuestionDAO) Create(question *models.Question) (int64, error) {
	query := `INSERT INTO questions (J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := dao.ExecuteUpdate(query,
		question.J, question.P, question.I, question.Q, question.T,
		question.A, question.B, question.C, question.D, question.F,
		question.LA, question.LB, question.LC, question.Type, question.UserID)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	question.ID = id
	utils.Info("QuestionDAO", "创建题目成功", map[string]interface{}{
		"question_id": question.ID,
		"type":        question.Type,
	})

	return question.ID, nil
}

// GetByID 根据 ID 获取题目
func (dao *QuestionDAO) GetByID(id int64) (*models.Question, error) {
	query := `SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
              FROM questions WHERE id = ?`

	row := dao.QueryRow(query, id)
	question, err := dao.scanQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return question, nil
}

// GetByType 根据题型获取题目
func (dao *QuestionDAO) GetByType(typeValue int) ([]*models.Question, error) {
	query := `SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
              FROM questions WHERE type = ?`

	rows, err := dao.ExecuteQuery(query, typeValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// GetRandomQuestions 随机获取指定数量的题目
func (dao *QuestionDAO) GetRandomQuestions(typeValue int, count int) ([]*models.Question, error) {
	query := `SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
              FROM questions WHERE type = ? ORDER BY RANDOM() LIMIT ?`

	rows, err := dao.ExecuteQuery(query, typeValue, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// Search 搜索题目
func (dao *QuestionDAO) Search(keyword string) ([]*models.Question, error) {
	query := `SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
              FROM questions 
              WHERE Q LIKE ? OR A LIKE ? OR B LIKE ? OR C LIKE ? OR D LIKE ?`

	likeKeyword := "%" + keyword + "%"
	rows, err := dao.ExecuteQuery(query, likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// GetByCategory 根据类别获取题目
func (dao *QuestionDAO) GetByCategory(category string) ([]*models.Question, error) {
	var query string
	if category == "all" {
		// 获取所有题目
		query = `SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
	              FROM questions`
	} else {
		categoryField := dao.getCategoryField(category)
		query = fmt.Sprintf(`SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
		                      FROM questions WHERE %s = 1`, categoryField)
	}

	rows, err := dao.ExecuteQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// GetRandomByCategoryAndType 根据类别和题型随机获取题目
func (dao *QuestionDAO) GetRandomByCategoryAndType(category string, typeValue int, count int) ([]*models.Question, error) {
	categoryField := dao.getCategoryField(category)
	query := fmt.Sprintf(`SELECT id, J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id 
	                      FROM questions WHERE %s = 1 AND type = ? ORDER BY RANDOM() LIMIT ?`, categoryField)

	rows, err := dao.ExecuteQuery(query, typeValue, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// Update 更新题目
func (dao *QuestionDAO) Update(question *models.Question) error {
	query := `UPDATE questions 
              SET J=?, P=?, I=?, Q=?, T=?, A=?, B=?, C=?, D=?, F=?, LA=?, LB=?, LC=?, type=?, user_id=? 
              WHERE id=?`

	_, err := dao.ExecuteUpdate(query,
		question.J, question.P, question.I, question.Q, question.T,
		question.A, question.B, question.C, question.D, question.F,
		question.LA, question.LB, question.LC, question.Type, question.UserID,
		question.ID)
	if err != nil {
		return err
	}

	utils.Debug("QuestionDAO", "更新题目成功", map[string]interface{}{
		"question_id": question.ID,
	})

	return nil
}

// Delete 删除题目
func (dao *QuestionDAO) Delete(id int64) error {
	query := `DELETE FROM questions WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, id)
	if err != nil {
		return err
	}

	utils.Info("QuestionDAO", "删除题目成功", map[string]interface{}{
		"question_id": id,
	})

	return nil
}

// BatchInsert 批量插入题目
func (dao *QuestionDAO) BatchInsert(questions []*models.Question) error {
	if len(questions) == 0 {
		return nil
	}

	// 开始事务
	tx, err := dao.db.Begin()
	if err != nil {
		utils.Error("QuestionDAO", "开始事务失败", err, nil)
		return err
	}

	query := `INSERT INTO questions (J, P, I, Q, T, A, B, C, D, F, LA, LB, LC, type, user_id) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, q := range questions {
		_, err := tx.Exec(query,
			q.J, q.P, q.I, q.Q, q.T,
			q.A, q.B, q.C, q.D, q.F,
			q.LA, q.LB, q.LC, q.Type, q.UserID)
		if err != nil {
			tx.Rollback()
			utils.Error("QuestionDAO", "批量插入题目失败", err, map[string]interface{}{
				"question": q,
			})
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		utils.Error("QuestionDAO", "提交事务失败", err, nil)
		return err
	}

	utils.Info("QuestionDAO", "批量插入题目成功", map[string]interface{}{
		"count": len(questions),
	})

	return nil
}

// ResetTable 清空题库表
func (dao *QuestionDAO) ResetTable() error {
	query := `DELETE FROM questions`
	_, err := dao.ExecuteUpdate(query)
	if err != nil {
		utils.Error("QuestionDAO", "清空题库表失败", err, nil)
		return err
	}

	utils.Info("QuestionDAO", "清空题库表成功", nil)
	return nil
}

// ClearAll 清空所有题目数据
func (dao *QuestionDAO) ClearAll() error {
	return dao.ResetTable()
}

// GetCount 获取题目总数
func (dao *QuestionDAO) GetCount() (int, error) {
	query := `SELECT COUNT(*) FROM questions`

	row := dao.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCountByCategory 根据类别获取题目总数
func (dao *QuestionDAO) GetCountByCategory(category string) (int, error) {
	categoryField := dao.getCategoryField(category)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM questions WHERE %s = 1`, categoryField)

	row := dao.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// getCategoryField 获取类别字段
func (dao *QuestionDAO) getCategoryField(category string) string {
	switch category {
	case "A":
		return "LA"
	case "B":
		return "LB"
	case "C":
		return "LC"
	default:
		return "LA"
	}
}

// scanQuestion 扫描单行数据
func (dao *QuestionDAO) scanQuestion(row Scanner) (*models.Question, error) {
	question := &models.Question{}
	err := row.Scan(
		&question.ID, &question.J, &question.P, &question.I, &question.Q, &question.T,
		&question.A, &question.B, &question.C, &question.D, &question.F,
		&question.LA, &question.LB, &question.LC, &question.Type, &question.UserID)
	if err != nil {
		return nil, err
	}
	return question, nil
}

// scanQuestionRow 扫描多行数据中的一行
func (dao *QuestionDAO) scanQuestionRow(rows *sql.Rows) (*models.Question, error) {
	question := &models.Question{}
	err := rows.Scan(
		&question.ID, &question.J, &question.P, &question.I, &question.Q, &question.T,
		&question.A, &question.B, &question.C, &question.D, &question.F,
		&question.LA, &question.LB, &question.LC, &question.Type, &question.UserID)
	if err != nil {
		return nil, err
	}
	return question, nil
}
