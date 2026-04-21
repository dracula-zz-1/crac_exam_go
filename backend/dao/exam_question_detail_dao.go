package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"

	"gorm.io/gorm"
)

// ExamQuestionDetailDAO 考试题目详情数据访问对象
type ExamQuestionDetailDAO struct {
	*BaseDAO
}

// NewExamQuestionDetailDAO 创建 ExamQuestionDetailDAO 实例
func NewExamQuestionDetailDAO(db *gorm.DB) *ExamQuestionDetailDAO {
	return &ExamQuestionDetailDAO{
		BaseDAO: NewBaseDAO(db, "exam_question_details"),
	}
}

// Create 创建考试题目详情
func (dao *ExamQuestionDetailDAO) Create(detail *models.ExamQuestionDetail) (int64, error) {
	result := dao.db.Create(detail)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("ExamQuestionDetailDAO", "创建考试题目详情成功", map[string]interface{}{
		"detail_id":   detail.ID,
		"exam_id":     detail.ExamID,
		"question_id": detail.QuestionID,
	})

	return detail.ID, nil
}

// GetByID 根据 ID 获取考试题目详情
func (dao *ExamQuestionDetailDAO) GetByID(detailID int64) (*models.ExamQuestionDetail, error) {
	detail := &models.ExamQuestionDetail{}
	result := dao.db.First(detail, detailID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return detail, nil
}

// GetByExamID 根据考试 ID 获取所有题目详情
func (dao *ExamQuestionDetailDAO) GetByExamID(examID int64) ([]*models.ExamQuestionDetail, error) {
	var details []*models.ExamQuestionDetail
	result := dao.db.Where("exam_id = ?", examID).Find(&details)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("ExamQuestionDetailDAO", "获取考试题目详情成功", map[string]interface{}{
		"exam_id": examID,
		"count":   len(details),
	})

	return details, nil
}

// GetByExamAndQuestion 根据考试 ID 和题目 ID 获取题目详情
func (dao *ExamQuestionDetailDAO) GetByExamAndQuestion(examID int64, questionID int64) (*models.ExamQuestionDetail, error) {
	detail := &models.ExamQuestionDetail{}
	result := dao.db.Where("exam_id = ? AND question_id = ?", examID, questionID).First(detail)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return detail, nil
}

// GetIncorrectQuestions 获取考试中的错题详情
func (dao *ExamQuestionDetailDAO) GetIncorrectQuestions(examID int64) ([]*models.ExamQuestionDetail, error) {
	var details []*models.ExamQuestionDetail
	result := dao.db.Where("exam_id = ? AND is_correct = 0", examID).Find(&details)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("ExamQuestionDetailDAO", "获取考试错题成功", map[string]interface{}{
		"exam_id": examID,
		"count":   len(details),
	})

	return details, nil
}

// Update 更新考试题目详情
func (dao *ExamQuestionDetailDAO) Update(detail *models.ExamQuestionDetail) error {
	result := dao.db.Save(detail)
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("ExamQuestionDetailDAO", "更新考试题目详情成功", map[string]interface{}{
		"detail_id": detail.ID,
	})

	return nil
}

// Delete 删除考试题目详情
func (dao *ExamQuestionDetailDAO) Delete(id int64) error {
	result := dao.db.Delete(&models.ExamQuestionDetail{}, id)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamQuestionDetailDAO", "删除考试题目详情成功", map[string]interface{}{
		"detail_id": id,
	})

	return nil
}

// DeleteByExamID 根据考试 ID 删除所有题目详情
func (dao *ExamQuestionDetailDAO) DeleteByExamID(examID int64) error {
	result := dao.db.Where("exam_id = ?", examID).Delete(&models.ExamQuestionDetail{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamQuestionDetailDAO", "删除考试题目详情成功", map[string]interface{}{
		"exam_id":       examID,
		"rows_affected": result.RowsAffected,
	})

	return nil
}

// BulkCreate 批量创建考试题目详情
func (dao *ExamQuestionDetailDAO) BulkCreate(details []*models.ExamQuestionDetail) error {
	if len(details) == 0 {
		return nil
	}

	result := dao.db.CreateInBatches(details, 100)
	if result.Error != nil {
		return result.Error
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

	tx := dao.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, detail := range details {
		result := tx.Model(&models.ExamQuestionDetail{}).Where("id = ?", detail.ID).
			Updates(map[string]interface{}{
				"user_answer": detail.UserAnswer,
				"is_correct":  detail.IsCorrect,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	return tx.Commit().Error
}

// DeleteByExamIDWithTx 在事务中根据考试 ID 删除所有题目详情
func (dao *ExamQuestionDetailDAO) DeleteByExamIDWithTx(examID int64, tx *gorm.DB) error {
	result := tx.Where("exam_id = ?", examID).Delete(&models.ExamQuestionDetail{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamQuestionDetailDAO", "删除考试题目详情成功（事务）", map[string]interface{}{
		"exam_id":       examID,
		"rows_affected": result.RowsAffected,
	})

	return nil
}
