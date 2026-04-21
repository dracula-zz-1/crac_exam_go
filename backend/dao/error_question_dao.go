package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"time"

	"gorm.io/gorm"
)

// ErrorQuestionDAO 错题数据访问对象
type ErrorQuestionDAO struct {
	*BaseDAO
}

// NewErrorQuestionDAO 创建 ErrorQuestionDAO 实例
func NewErrorQuestionDAO(db *gorm.DB) *ErrorQuestionDAO {
	return &ErrorQuestionDAO{
		BaseDAO: NewBaseDAO(db, "error_questions"),
	}
}

// Create 创建错题记录
func (dao *ErrorQuestionDAO) Create(errorQuestion *models.ErrorQuestion) (int64, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	errorQuestion.CreatedAt = now
	result := dao.db.Create(errorQuestion)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("ErrorQuestionDAO", "创建错题记录成功", map[string]interface{}{
		"error_id":    errorQuestion.ID,
		"question_id": errorQuestion.QuestionID,
		"user_id":     errorQuestion.UserID,
		"category":    errorQuestion.Category,
	})

	return errorQuestion.ID, nil
}

// GetByID 根据 ID 获取错题
func (dao *ErrorQuestionDAO) GetByID(id int64) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	result := dao.db.First(eq, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return eq, nil
}

// GetByUserID 根据用户 ID 获取所有错题
func (dao *ErrorQuestionDAO) GetByUserID(userID int64) ([]*models.ErrorQuestion, error) {
	var errorQuestions []*models.ErrorQuestion
	result := dao.db.Where("user_id = ?", userID).Find(&errorQuestions)
	if result.Error != nil {
		return nil, result.Error
	}

	return errorQuestions, nil
}

// BatchGetByUserQuestionAndCategory 批量查询用户题目是否已在错题本中
func (dao *ErrorQuestionDAO) BatchGetByUserQuestionAndCategory(userID int64, category string, questionIDs []int64) (map[int64]bool, error) {
	if len(questionIDs) == 0 {
		return make(map[int64]bool), nil
	}

	var errorQuestions []*models.ErrorQuestion
	result := dao.db.Where("user_id = ? AND category = ? AND question_id IN ?", userID, category, questionIDs).Find(&errorQuestions)
	if result.Error != nil {
		return nil, result.Error
	}

	existingMap := make(map[int64]bool)
	for _, eq := range errorQuestions {
		existingMap[eq.QuestionID] = true
	}

	utils.Debug("ErrorQuestionDAO", "批量查询错题成功", map[string]interface{}{
		"user_id":     userID,
		"category":    category,
		"query_count": len(questionIDs),
		"found_count": len(existingMap),
	})

	return existingMap, nil
}

// GetByUserAndQuestion 根据用户 ID 和题目 ID 获取错题
func (dao *ErrorQuestionDAO) GetByUserAndQuestion(userID int64, questionID int64) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	result := dao.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(eq)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return eq, nil
}

// GetByUserQuestionAndCategory 根据用户 ID、题目 ID 和类别获取错题
func (dao *ErrorQuestionDAO) GetByUserQuestionAndCategory(userID int64, questionID int64, category string) (*models.ErrorQuestion, error) {
	eq := &models.ErrorQuestion{}
	result := dao.db.Where("user_id = ? AND question_id = ? AND category = ?", userID, questionID, category).First(eq)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return eq, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取错题
func (dao *ErrorQuestionDAO) GetByUserAndCategory(userID int64, category string) ([]*models.ErrorQuestion, error) {
	var errorQuestions []*models.ErrorQuestion
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Find(&errorQuestions)
	if result.Error != nil {
		return nil, result.Error
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
	result := dao.db.Delete(&models.ErrorQuestion{}, id)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ErrorQuestionDAO", "删除错题成功", map[string]interface{}{
		"error_id": id,
	})

	return nil
}

// DeleteByUserAndQuestion 根据用户 ID 和题目 ID 删除错题
func (dao *ErrorQuestionDAO) DeleteByUserAndQuestion(userID int64, questionID int64) error {
	result := dao.db.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.ErrorQuestion{})
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("ErrorQuestionDAO", "删除用户错题成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	return nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除错题
func (dao *ErrorQuestionDAO) DeleteByUserAndCategory(userID int64, category string) error {
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Delete(&models.ErrorQuestion{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ErrorQuestionDAO", "删除用户类别错题成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的错题
func (dao *ErrorQuestionDAO) ClearByUser(userID int64) error {
	result := dao.db.Where("user_id = ?", userID).Delete(&models.ErrorQuestion{})
	if result.Error != nil {
		return result.Error
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

	now := time.Now().Format("2006-01-02 15:04:05")
	for _, eq := range errorQuestions {
		eq.CreatedAt = now
	}

	result := dao.db.CreateInBatches(errorQuestions, 100)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ErrorQuestionDAO", "批量创建错题记录成功", map[string]interface{}{
		"count": len(errorQuestions),
	})

	return nil
}

// GetCountByUser 获取用户错题数量
func (dao *ErrorQuestionDAO) GetCountByUser(userID int64) (int64, error) {
	var count int64
	result := dao.db.Model(&models.ErrorQuestion{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("ErrorQuestionDAO", "获取用户错题数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// GetCountByUserAndCategory 获取用户指定类别的错题数量
func (dao *ErrorQuestionDAO) GetCountByUserAndCategory(userID int64, category string) (int64, error) {
	var count int64
	result := dao.db.Model(&models.ErrorQuestion{}).Where("user_id = ? AND category = ?", userID, category).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("ErrorQuestionDAO", "获取用户类别错题数量成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    count,
	})

	return count, nil
}

// GetCount 获取错题总数
func (dao *ErrorQuestionDAO) GetCount() (int64, error) {
	var count int64
	result := dao.db.Model(&models.ErrorQuestion{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("ErrorQuestionDAO", "获取错题总数成功", map[string]interface{}{
		"count": count,
	})

	return count, nil
}

// GetErrorQuestionsWithDetails 根据用户 ID 和类别获取错题，并关联查询题目详情
func (dao *ErrorQuestionDAO) GetErrorQuestionsWithDetails(userID int64, category string) ([]*models.ErrorQuestion, error) {
	query := `
		SELECT 
			eq.id, eq.question_id, eq.category, eq.user_id, eq.created_at,
			q.J, q.P, q.I, q.Q, q.T, q.A, q.B, q.C, q.D, q.F, q.LA, q.LB, q.LC, q.type
		FROM error_questions eq
		INNER JOIN questions q ON eq.question_id = q.id
		WHERE eq.user_id = ? AND eq.category = ?
	`

	rows, err := dao.db.Raw(query, userID, category).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorQuestions []*models.ErrorQuestion
	for rows.Next() {
		eq := &models.ErrorQuestion{}
		var j, p, i, q_text, t, a, b, c, d, f string
		var la, lb, lc, questionType int

		err := rows.Scan(
			&eq.ID, &eq.QuestionID, &eq.Category, &eq.UserID, &eq.CreatedAt,
			&j, &p, &i, &q_text, &t,
			&a, &b, &c, &d, &f,
			&la, &lb, &lc, &questionType)
		if err != nil {
			continue
		}

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

		errorQuestions = append(errorQuestions, eq)
	}

	utils.Debug("ErrorQuestionDAO", "获取用户类别错题详情成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(errorQuestions),
	})

	return errorQuestions, nil
}
