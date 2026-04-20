package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
)

// ExamQuestionDetailDAO 考试题目详情数据访问对象
type ExamQuestionDetailDAO struct {
	*BaseDAO
}

// NewExamQuestionDetailDAO 创建 ExamQuestionDetailDAO 实例
func NewExamQuestionDetailDAO(db *sql.DB) *ExamQuestionDetailDAO {
	return &ExamQuestionDetailDAO{
		BaseDAO: NewBaseDAO(db, "exam_question_details"),
	}
}

// Create 创建考试题目详情
func (dao *ExamQuestionDetailDAO) Create(detail *models.ExamQuestionDetail) (int64, error) {
	query := `INSERT INTO exam_question_details (exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := dao.ExecuteUpdate(query,
		detail.ExamID, detail.QuestionID, detail.QuestionText,
		detail.OptionA, detail.OptionB, detail.OptionC, detail.OptionD,
		detail.CorrectAnswer, detail.UserAnswer, detail.IsCorrect,
		detail.Type, detail.ImageData)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	detail.ID = id
	utils.Debug("ExamQuestionDetailDAO", "创建考试题目详情成功", map[string]interface{}{
		"detail_id":   detail.ID,
		"exam_id":     detail.ExamID,
		"question_id": detail.QuestionID,
	})

	return detail.ID, nil
}

// GetByID 根据 ID 获取考试题目详情
func (dao *ExamQuestionDetailDAO) GetByID(detailID int64) (*models.ExamQuestionDetail, error) {
	query := `SELECT id, exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data 
              FROM exam_question_details WHERE id = ?`

	row := dao.QueryRow(query, detailID)
	detail, err := dao.scanExamQuestionDetail(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return detail, nil
}

// GetByExamID 根据考试 ID 获取所有题目详情
func (dao *ExamQuestionDetailDAO) GetByExamID(examID int64) ([]*models.ExamQuestionDetail, error) {
	query := `SELECT id, exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data 
              FROM exam_question_details WHERE exam_id = ?`

	rows, err := dao.ExecuteQuery(query, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []*models.ExamQuestionDetail
	for rows.Next() {
		detail, err := dao.scanExamQuestionDetailRow(rows)
		if err != nil {
			return nil, err
		}
		details = append(details, detail)
	}

	utils.Debug("ExamQuestionDetailDAO", "获取考试题目详情成功", map[string]interface{}{
		"exam_id": examID,
		"count":   len(details),
	})

	return details, nil
}

// GetByExamAndQuestion 根据考试 ID 和题目 ID 获取题目详情
func (dao *ExamQuestionDetailDAO) GetByExamAndQuestion(examID int64, questionID int64) (*models.ExamQuestionDetail, error) {
	query := `SELECT id, exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data 
              FROM exam_question_details WHERE exam_id = ? AND question_id = ?`

	row := dao.QueryRow(query, examID, questionID)
	detail, err := dao.scanExamQuestionDetail(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return detail, nil
}

// DeleteByExamIDWithTx 根据考试 ID 删除题目详情（支持事务）
func (dao *ExamQuestionDetailDAO) DeleteByExamIDWithTx(examID int64, tx *sql.Tx) error {
	query := `DELETE FROM exam_question_details WHERE exam_id = ?`

	_, err := tx.Exec(query, examID)
	if err != nil {
		utils.Error("ExamQuestionDetailDAO", "删除考试题目详情失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	utils.Debug("ExamQuestionDetailDAO", "删除考试题目详情成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// GetIncorrectQuestions 获取考试中的错题详情
func (dao *ExamQuestionDetailDAO) GetIncorrectQuestions(examID int64) ([]*models.ExamQuestionDetail, error) {
	query := `SELECT id, exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data 
              FROM exam_question_details WHERE exam_id = ? AND is_correct = 0`

	rows, err := dao.ExecuteQuery(query, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []*models.ExamQuestionDetail
	for rows.Next() {
		detail, err := dao.scanExamQuestionDetailRow(rows)
		if err != nil {
			return nil, err
		}
		details = append(details, detail)
	}

	utils.Debug("ExamQuestionDetailDAO", "获取考试错题成功", map[string]interface{}{
		"exam_id": examID,
		"count":   len(details),
	})

	return details, nil
}

// Update 更新考试题目详情
func (dao *ExamQuestionDetailDAO) Update(detail *models.ExamQuestionDetail) error {
	query := `UPDATE exam_question_details 
              SET exam_id=?, question_id=?, question_text=?, option_a=?, option_b=?, option_c=?, option_d=?, correct_answer=?, user_answer=?, is_correct=?, type=?, image_data=? 
              WHERE id=?`

	_, err := dao.ExecuteUpdate(query,
		detail.ExamID, detail.QuestionID, detail.QuestionText,
		detail.OptionA, detail.OptionB, detail.OptionC, detail.OptionD,
		detail.CorrectAnswer, detail.UserAnswer, detail.IsCorrect,
		detail.Type, detail.ImageData, detail.ID)
	if err != nil {
		return err
	}

	utils.Debug("ExamQuestionDetailDAO", "更新考试题目详情成功", map[string]interface{}{
		"detail_id": detail.ID,
	})

	return nil
}

// Delete 删除考试题目详情
func (dao *ExamQuestionDetailDAO) Delete(id int64) error {
	query := `DELETE FROM exam_question_details WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, id)
	if err != nil {
		return err
	}

	utils.Info("ExamQuestionDetailDAO", "删除考试题目详情成功", map[string]interface{}{
		"detail_id": id,
	})

	return nil
}

// DeleteByExamID 根据考试 ID 删除所有题目详情
func (dao *ExamQuestionDetailDAO) DeleteByExamID(examID int64) error {
	query := `DELETE FROM exam_question_details WHERE exam_id = ?`

	result, err := dao.ExecuteUpdate(query, examID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	utils.Info("ExamQuestionDetailDAO", "删除考试题目详情成功", map[string]interface{}{
		"exam_id":       examID,
		"rows_affected": rowsAffected,
	})

	return nil
}

// BulkCreate 批量创建考试题目详情
func (dao *ExamQuestionDetailDAO) BulkCreate(details []*models.ExamQuestionDetail) error {
	if len(details) == 0 {
		return nil
	}

	// 开始事务
	tx, err := dao.db.Begin()
	if err != nil {
		utils.Error("ExamQuestionDetailDAO", "开始事务失败", err, nil)
		return err
	}

	query := `INSERT INTO exam_question_details (exam_id, question_id, question_text, option_a, option_b, option_c, option_d, correct_answer, user_answer, is_correct, type, image_data) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, detail := range details {
		_, err := tx.Exec(query,
			detail.ExamID, detail.QuestionID, detail.QuestionText,
			detail.OptionA, detail.OptionB, detail.OptionC, detail.OptionD,
			detail.CorrectAnswer, detail.UserAnswer, detail.IsCorrect,
			detail.Type, detail.ImageData)
		if err != nil {
			tx.Rollback()
			utils.Error("ExamQuestionDetailDAO", "批量创建考试题目详情失败", err, map[string]interface{}{
				"exam_id": detail.ExamID,
			})
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		utils.Error("ExamQuestionDetailDAO", "提交事务失败", err, nil)
		return err
	}

	utils.Info("ExamQuestionDetailDAO", "批量创建考试题目详情成功", map[string]interface{}{
		"count": len(details),
	})

	return nil
}

// BulkUpdate 批量更新考试题目详情
func (dao *ExamQuestionDetailDAO) BulkUpdate(details []*models.ExamQuestionDetail) error {
	if len(details) == 0 {
		return nil
	}

	// 开始事务
	tx, err := dao.db.Begin()
	if err != nil {
		utils.Error("ExamQuestionDetailDAO", "开始事务失败", err, nil)
		return err
	}

	query := `UPDATE exam_question_details 
              SET user_answer=?, is_correct=? 
              WHERE id=?`

	for _, detail := range details {
		_, err := tx.Exec(query, detail.UserAnswer, detail.IsCorrect, detail.ID)
		if err != nil {
			tx.Rollback()
			utils.Error("ExamQuestionDetailDAO", "批量更新考试题目详情失败", err, map[string]interface{}{
				"exam_id": detail.ExamID,
			})
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		utils.Error("ExamQuestionDetailDAO", "提交事务失败", err, nil)
		return err
	}

	utils.Info("ExamQuestionDetailDAO", "批量更新考试题目详情成功", map[string]interface{}{
		"count": len(details),
	})

	return nil
}

// scanExamQuestionDetail 扫描单行数据
func (dao *ExamQuestionDetailDAO) scanExamQuestionDetail(row Scanner) (*models.ExamQuestionDetail, error) {
	detail := &models.ExamQuestionDetail{}
	err := row.Scan(
		&detail.ID, &detail.ExamID, &detail.QuestionID, &detail.QuestionText,
		&detail.OptionA, &detail.OptionB, &detail.OptionC, &detail.OptionD,
		&detail.CorrectAnswer, &detail.UserAnswer, &detail.IsCorrect,
		&detail.Type, &detail.ImageData)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// scanExamQuestionDetailRow 扫描多行数据中的一行
func (dao *ExamQuestionDetailDAO) scanExamQuestionDetailRow(rows *sql.Rows) (*models.ExamQuestionDetail, error) {
	detail := &models.ExamQuestionDetail{}
	err := rows.Scan(
		&detail.ID, &detail.ExamID, &detail.QuestionID, &detail.QuestionText,
		&detail.OptionA, &detail.OptionB, &detail.OptionC, &detail.OptionD,
		&detail.CorrectAnswer, &detail.UserAnswer, &detail.IsCorrect,
		&detail.Type, &detail.ImageData)
	if err != nil {
		return nil, err
	}
	return detail, nil
}
