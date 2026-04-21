package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"time"

	"gorm.io/gorm"
)

// ExamRecordDAO 考试记录数据访问对象
type ExamRecordDAO struct {
	*BaseDAO
}

// NewExamRecordDAO 创建 ExamRecordDAO 实例
func NewExamRecordDAO(db *gorm.DB) *ExamRecordDAO {
	return &ExamRecordDAO{
		BaseDAO: NewBaseDAO(db, "exam_records"),
	}
}

// Create 创建考试记录
func (dao *ExamRecordDAO) Create(record *models.ExamRecord) (int64, error) {
	result := dao.db.Create(record)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Info("ExamRecordDAO", "创建考试记录成功", map[string]interface{}{
		"exam_id":  record.ID,
		"category": record.Category,
		"user_id":  record.UserID,
		"score":    record.Score,
	})

	return record.ID, nil
}

// GetByID 根据 ID 获取考试记录
func (dao *ExamRecordDAO) GetByID(examID int64) (*models.ExamRecord, error) {
	record := &models.ExamRecord{}
	result := dao.db.First(record, examID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	utils.Debug("ExamRecordDAO", "获取考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})

	return record, nil
}

// GetByUserID 根据用户 ID 获取所有考试记录
func (dao *ExamRecordDAO) GetByUserID(userID int64) ([]*models.ExamRecord, error) {
	var records []*models.ExamRecord
	result := dao.db.Where("user_id = ?", userID).Order("exam_date DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("ExamRecordDAO", "获取用户考试记录成功", map[string]interface{}{
		"user_id": userID,
		"count":   len(records),
	})

	return records, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取考试记录
func (dao *ExamRecordDAO) GetByUserAndCategory(userID int64, category string) ([]*models.ExamRecord, error) {
	var records []*models.ExamRecord
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Order("exam_date DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("ExamRecordDAO", "获取用户类别考试记录成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(records),
	})

	return records, nil
}

// Delete 删除考试记录
func (dao *ExamRecordDAO) Delete(examID int64) error {
	result := dao.db.Delete(&models.ExamRecord{}, examID)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamRecordDAO", "删除考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// DeleteByExamID 根据考试 ID 删除所有题目详情
func (dao *ExamRecordDAO) DeleteByExamID(examID int64, tx *gorm.DB) error {
	db := tx
	if db == nil {
		db = dao.db
	}
	result := db.Delete(&models.ExamRecord{}, examID)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamRecordDAO", "删除考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// GetRecentExams 获取用户最近的考试记录
func (dao *ExamRecordDAO) GetRecentExams(userID int64, limit int) ([]*models.ExamRecord, error) {
	var records []*models.ExamRecord
	result := dao.db.Where("user_id = ?", userID).Order("exam_date DESC").Limit(limit).Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("ExamRecordDAO", "获取用户最近考试记录成功", map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"count":   len(records),
	})

	return records, nil
}

// Update 更新考试记录
func (dao *ExamRecordDAO) Update(record *models.ExamRecord) error {
	result := dao.db.Save(record)
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("ExamRecordDAO", "更新考试记录成功", map[string]interface{}{
		"exam_id": record.ID,
	})

	return nil
}

// ClearByUser 清空用户的考试记录
func (dao *ExamRecordDAO) ClearByUser(userID int64) error {
	utils.Info("ExamRecordDAO", "清空用户考试记录", map[string]interface{}{
		"user_id": userID,
	})

	// 删除该用户的所有考试记录
	result := dao.db.Where("user_id = ?", userID).Delete(&models.ExamRecord{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamRecordDAO", "清空用户考试记录成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// GetExamStatistics 获取考试统计数据
func (dao *ExamRecordDAO) GetExamStatistics(userID int64, category string, startDate time.Time) ([]*models.ExamStatisticsData, error) {
	utils.Debug("ExamRecordDAO", "获取考试统计数据", map[string]interface{}{
		"user_id":    userID,
		"category":   category,
		"start_date": startDate,
	})

	// 使用原生 SQL 进行 JOIN 查询
	query := `
		SELECT er.id, er.category, er.exam_date,
		       COUNT(eqd.id) AS total_questions,
		       SUM(CASE WHEN eqd.is_correct = 1 THEN 1 ELSE 0 END) AS correct_questions,
		       AVG(CASE WHEN eqd.is_correct = 1 THEN 1.0 ELSE 0.0 END) AS pass_rate,
		       er.duration_seconds AS duration_seconds
		FROM exam_records er
		JOIN exam_question_details eqd ON er.id = eqd.exam_id
		WHERE er.user_id = ?`

	params := []interface{}{userID}

	if !startDate.IsZero() {
		query += " AND er.exam_date >= ?"
		params = append(params, startDate)
	}

	if category != "" {
		query += " AND er.category = ?"
		params = append(params, category)
	}

	query += `
		GROUP BY er.id
		ORDER BY er.exam_date ASC
	`

	rows, err := dao.db.Raw(query, params...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var examData []*models.ExamStatisticsData
	for rows.Next() {
		data := &models.ExamStatisticsData{}
		var totalQuestions, correctQuestions int
		var passRate float64
		var durationSeconds float64

		err := rows.Scan(
			&data.ID, &data.Category, &data.ExamDate,
			&totalQuestions, &correctQuestions, &passRate, &durationSeconds,
		)
		if err != nil {
			continue
		}

		data.TotalQuestions = totalQuestions
		data.CorrectQuestions = correctQuestions
		data.PassRate = passRate * 100
		data.DurationSeconds = durationSeconds
		data.Score = float64(correctQuestions)

		examData = append(examData, data)
	}

	utils.Debug("ExamRecordDAO", "获取考试统计数据成功", map[string]interface{}{
		"count": len(examData),
	})

	return examData, nil
}

// GetCount 获取考试记录总数
func (dao *ExamRecordDAO) GetCount() (int64, error) {
	var count int64
	result := dao.db.Model(&models.ExamRecord{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetCountByUser 获取用户的考试记录总数
func (dao *ExamRecordDAO) GetCountByUser(userID int64) (int64, error) {
	var count int64
	result := dao.db.Model(&models.ExamRecord{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// DeleteWithTx 在事务中删除考试记录
func (dao *ExamRecordDAO) DeleteWithTx(examID int64, tx *gorm.DB) error {
	result := tx.Delete(&models.ExamRecord{}, examID)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("ExamRecordDAO", "删除考试记录成功（事务）", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// GetDB 获取数据库实例
func (dao *ExamRecordDAO) GetDB() *gorm.DB {
	return dao.db
}
