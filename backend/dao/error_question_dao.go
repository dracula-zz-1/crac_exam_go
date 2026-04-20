package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// ErrorQuestionDAO 错题数据访问对象
type ErrorQuestionDAO struct {
	*BaseDAO
}

// NewErrorQuestionDAO 创建 ErrorQuestionDAO 实例
func NewErrorQuestionDAO(db *sql.DB) *ErrorQuestionDAO {
	return &ErrorQuestionDAO{
		BaseDAO: NewBaseDAO(db, "error_questions"),
	}
}

// Create 创建错题记录（只存储引用信息）
func (dao *ErrorQuestionDAO) Create(errorQuestion *models.ErrorQuestion) (int64, error) {
	query := `INSERT INTO error_questions (question_id, category, user_id, created_at) 
              VALUES (?, ?, ?, ?)`

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := dao.ExecuteUpdate(query,
		errorQuestion.QuestionID, errorQuestion.Category, errorQuestion.UserID, now)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	errorQuestion.ID = id
	errorQuestion.CreatedAt = now
	utils.Debug("ErrorQuestionDAO", "创建错题记录成功", map[string]interface{}{
		"error_id":    errorQuestion.ID,
		"question_id": errorQuestion.QuestionID,
		"user_id":     errorQuestion.UserID,
		"category":    errorQuestion.Category,
	})

	return errorQuestion.ID, nil
}

// GetByID 根据 ID 获取错题（包含题目详情）
func (dao *ErrorQuestionDAO) GetByID(id int64) (*models.ErrorQuestion, error) {
	query := `SELECT id, question_id, category, user_id, created_at 
              FROM error_questions WHERE id = ?`

	row := dao.QueryRow(query, id)
	errorQuestion, err := dao.scanErrorQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return errorQuestion, nil
}

// GetByUserID 根据用户 ID 获取所有错题（包含题目详情）
func (dao *ErrorQuestionDAO) GetByUserID(userID int64) ([]*models.ErrorQuestion, error) {
	query := `SELECT id, question_id, category, user_id, created_at 
              FROM error_questions WHERE user_id = ?`

	rows, err := dao.ExecuteQuery(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorQuestions []*models.ErrorQuestion
	for rows.Next() {
		errorQuestion, err := dao.scanErrorQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		errorQuestions = append(errorQuestions, errorQuestion)
	}

	return errorQuestions, nil
}

// GetByUserAndQuestion 根据用户 ID 和题目 ID 获取错题
func (dao *ErrorQuestionDAO) GetByUserAndQuestion(userID int64, questionID int64) (*models.ErrorQuestion, error) {
	query := `SELECT id, question_id, category, user_id, created_at 
              FROM error_questions WHERE user_id = ? AND question_id = ?`

	row := dao.QueryRow(query, userID, questionID)
	errorQuestion, err := dao.scanErrorQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return errorQuestion, nil
}

// GetByUserQuestionAndCategory 根据用户 ID、题目 ID 和类别获取错题
func (dao *ErrorQuestionDAO) GetByUserQuestionAndCategory(userID int64, questionID int64, category string) (*models.ErrorQuestion, error) {
	query := `SELECT id, question_id, category, user_id, created_at 
              FROM error_questions WHERE user_id = ? AND question_id = ? AND category = ?`

	row := dao.QueryRow(query, userID, questionID, category)
	errorQuestion, err := dao.scanErrorQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return errorQuestion, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取错题（只返回错题记录）
func (dao *ErrorQuestionDAO) GetByUserAndCategory(userID int64, category string) ([]*models.ErrorQuestion, error) {
	query := `SELECT id, question_id, category, user_id, created_at 
              FROM error_questions WHERE user_id = ? AND category = ?`

	rows, err := dao.ExecuteQuery(query, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorQuestions []*models.ErrorQuestion
	for rows.Next() {
		errorQuestion, err := dao.scanErrorQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		errorQuestions = append(errorQuestions, errorQuestion)
	}

	utils.Debug("ErrorQuestionDAO", "获取用户类别错题成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(errorQuestions),
	})

	return errorQuestions, nil
}

// Delete 删除错题
func (dao *ErrorQuestionDAO) Delete(id int64) error {
	query := `DELETE FROM error_questions WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, id)
	if err != nil {
		return err
	}

	utils.Info("ErrorQuestionDAO", "删除错题成功", map[string]interface{}{
		"error_id": id,
	})

	return nil
}

// DeleteByUserAndQuestion 根据用户 ID 和题目 ID 删除错题
func (dao *ErrorQuestionDAO) DeleteByUserAndQuestion(userID int64, questionID int64) error {
	query := `DELETE FROM error_questions WHERE user_id = ? AND question_id = ?`

	_, err := dao.ExecuteUpdate(query, userID, questionID)
	if err != nil {
		return err
	}

	utils.Debug("ErrorQuestionDAO", "删除用户错题成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	return nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除错题
func (dao *ErrorQuestionDAO) DeleteByUserAndCategory(userID int64, category string) error {
	query := `DELETE FROM error_questions WHERE user_id = ? AND category = ?`

	_, err := dao.ExecuteUpdate(query, userID, category)
	if err != nil {
		return err
	}

	utils.Info("ErrorQuestionDAO", "删除用户类别错题成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的错题
func (dao *ErrorQuestionDAO) ClearByUser(userID int64) error {
	query := `DELETE FROM error_questions WHERE user_id = ?`

	_, err := dao.ExecuteUpdate(query, userID)
	if err != nil {
		return err
	}

	utils.Info("ErrorQuestionDAO", "清空用户错题成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// BulkCreate 批量创建错题记录
func (dao *ErrorQuestionDAO) BulkCreate(errorQuestions []*models.ErrorQuestion) error {
	if len(errorQuestions) == 0 {
		return nil
	}

	// 开始事务
	tx, err := dao.db.Begin()
	if err != nil {
		utils.Error("ErrorQuestionDAO", "开始事务失败", err, nil)
		return err
	}

	query := `INSERT INTO error_questions (question_id, category, user_id, created_at) 
              VALUES (?, ?, ?, ?)`

	now := time.Now().Format("2006-01-02 15:04:05")
	for _, eq := range errorQuestions {
		_, err := tx.Exec(query, eq.QuestionID, eq.Category, eq.UserID, now)
		if err != nil {
			tx.Rollback()
			utils.Error("ErrorQuestionDAO", "批量创建错题记录失败", err, map[string]interface{}{
				"user_id": eq.UserID,
			})
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		utils.Error("ErrorQuestionDAO", "提交事务失败", err, nil)
		return err
	}

	utils.Info("ErrorQuestionDAO", "批量创建错题记录成功", map[string]interface{}{
		"count": len(errorQuestions),
	})

	return nil
}

// GetCountByUser 获取用户错题数量
func (dao *ErrorQuestionDAO) GetCountByUser(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM error_questions WHERE user_id = ?`

	row := dao.QueryRow(query, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("ErrorQuestionDAO", "获取用户错题数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// GetCountByUserAndCategory 获取用户指定类别的错题数量
func (dao *ErrorQuestionDAO) GetCountByUserAndCategory(userID int64, category string) (int, error) {
	query := `SELECT COUNT(*) FROM error_questions WHERE user_id = ? AND category = ?`

	row := dao.QueryRow(query, userID, category)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("ErrorQuestionDAO", "获取用户类别错题数量成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    count,
	})

	return count, nil
}

// GetCount 获取错题总数
func (dao *ErrorQuestionDAO) GetCount() (int, error) {
	query := `SELECT COUNT(*) FROM error_questions`

	row := dao.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("ErrorQuestionDAO", "获取错题总数成功", map[string]interface{}{
		"count": count,
	})

	return count, nil
}

// GetErrorQuestionsWithDetails 根据用户 ID 和类别获取错题，并关联查询题目详情
// Python 原版：get_error_questions_with_details(user_id, category) -> List[Question]
// 使用 JOIN 查询获取最新题目信息（当题库修改时，错题本中的题目也会更新）
func (dao *ErrorQuestionDAO) GetErrorQuestionsWithDetails(userID int64, category string) ([]*models.ErrorQuestion, error) {
	// 使用 JOIN 查询错题表和题目表，获取最新题目信息
	query := `
		SELECT 
			eq.id, eq.question_id, eq.category, eq.user_id, eq.created_at,
			q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F, q.LA, q.LB, q.LC, q.type
		FROM error_questions eq
		INNER JOIN questions q ON eq.question_id = q.id
		WHERE eq.user_id = ? AND eq.category = ?
	`

	rows, err := dao.ExecuteQuery(query, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorQuestions []*models.ErrorQuestion
	for rows.Next() {
		eq, err := dao.scanErrorQuestionWithDetailsRow(rows)
		if err != nil {
			return nil, err
		}
		errorQuestions = append(errorQuestions, eq)
	}

	utils.Debug("ErrorQuestionDAO", "获取用户类别错题详情成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(errorQuestions),
	})

	return errorQuestions, nil
}

// scanErrorQuestion 扫描单行数据（只扫描错题表字段）
func (dao *ErrorQuestionDAO) scanErrorQuestion(row Scanner) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	err := row.Scan(
		&eq.ID, &eq.QuestionID, &eq.Category, &eq.UserID, &eq.CreatedAt)
	if err != nil {
		return nil, err
	}
	return eq, nil
}

// scanErrorQuestionRow 扫描多行数据中的一行（只扫描错题表字段）
func (dao *ErrorQuestionDAO) scanErrorQuestionRow(rows *sql.Rows) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	err := rows.Scan(
		&eq.ID, &eq.QuestionID, &eq.Category, &eq.UserID, &eq.CreatedAt)
	if err != nil {
		return nil, err
	}
	return eq, nil
}

// scanErrorQuestionWithDetailsRow 扫描 JOIN 查询结果（包含错题表和题目表字段）
func (dao *ErrorQuestionDAO) scanErrorQuestionWithDetailsRow(rows *sql.Rows) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	var j, p, i, q_text, t, a, b, c, d, f string
	var la, lb, lc, questionType int

	err := rows.Scan(
		&eq.ID, &eq.QuestionID, &eq.Category, &eq.UserID, &eq.CreatedAt,
		&j, &p, &i, &q_text, &t,
		&a, &b, &c, &d, &f,
		&la, &lb, &lc, &questionType)
	if err != nil {
		return nil, err
	}

	// 将题目字段填充到 ErrorQuestion 中（用于返回给 PracticeService）
	// 注意：这里只是为了兼容现有接口，实际使用时应该转换为 Question 对象
	eq.J = j
	eq.P = p
	eq.I = i
	eq.Q = q_text
	eq.T = t
	eq.A = a
	eq.B = b
	eq.C = c
	eq.D = d
	eq.F = f
	eq.LA = la
	eq.LB = lb
	eq.LC = lc
	eq.Type = questionType

	return eq, nil
}
