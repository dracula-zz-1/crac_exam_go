package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// ExamRecordDAO 考试记录数据访问对象
type ExamRecordDAO struct {
	*BaseDAO
}

// NewExamRecordDAO 创建 ExamRecordDAO 实例
func NewExamRecordDAO(db *sql.DB) *ExamRecordDAO {
	return &ExamRecordDAO{
		BaseDAO: NewBaseDAO(db, "exam_records"),
	}
}

// Create 创建考试记录
func (dao *ExamRecordDAO) Create(record *models.ExamRecord) (int64, error) {
	query := `INSERT INTO exam_records (category, exam_date, duration_seconds, user_id, score, total_questions, correct_count) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := dao.ExecuteUpdate(query,
		record.Category, record.ExamDate, record.DurationSeconds,
		record.UserID, record.Score, record.TotalQuestions, record.CorrectCount)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	record.ID = id
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
	query := `SELECT id, category, exam_date, duration_seconds, user_id, score, total_questions, correct_count 
              FROM exam_records WHERE id = ?`

	row := dao.QueryRow(query, examID)
	record, err := dao.scanExamRecord(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	utils.Debug("ExamRecordDAO", "获取考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})

	return record, nil
}

// GetByUserID 根据用户 ID 获取所有考试记录
func (dao *ExamRecordDAO) GetByUserID(userID int64) ([]*models.ExamRecord, error) {
	query := `SELECT id, category, exam_date, duration_seconds, user_id, score, total_questions, correct_count 
              FROM exam_records WHERE user_id = ? ORDER BY exam_date DESC`

	rows, err := dao.ExecuteQuery(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ExamRecord
	for rows.Next() {
		record, err := dao.scanExamRecordRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	utils.Debug("ExamRecordDAO", "获取用户考试记录成功", map[string]interface{}{
		"user_id": userID,
		"count":   len(records),
	})

	return records, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取考试记录
func (dao *ExamRecordDAO) GetByUserAndCategory(userID int64, category string) ([]*models.ExamRecord, error) {
	query := `SELECT id, category, exam_date, duration_seconds, user_id, score, total_questions, correct_count 
              FROM exam_records WHERE user_id = ? AND category = ? ORDER BY exam_date DESC`

	rows, err := dao.ExecuteQuery(query, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ExamRecord
	for rows.Next() {
		record, err := dao.scanExamRecordRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
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
	query := `DELETE FROM exam_records WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, examID)
	if err != nil {
		utils.Error("ExamRecordDAO", "删除考试记录失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	utils.Info("ExamRecordDAO", "删除考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// DeleteWithTx 删除考试记录（支持事务）
func (dao *ExamRecordDAO) DeleteWithTx(examID int64, tx *sql.Tx) error {
	query := `DELETE FROM exam_records WHERE id = ?`

	_, err := tx.Exec(query, examID)
	if err != nil {
		utils.Error("ExamRecordDAO", "删除考试记录失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	utils.Info("ExamRecordDAO", "删除考试记录成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// GetRecentExams 获取用户最近的考试记录
func (dao *ExamRecordDAO) GetRecentExams(userID int64, limit int) ([]*models.ExamRecord, error) {
	query := `SELECT id, category, exam_date, duration_seconds, user_id, score, total_questions, correct_count 
              FROM exam_records WHERE user_id = ? ORDER BY exam_date DESC LIMIT ?`

	rows, err := dao.ExecuteQuery(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ExamRecord
	for rows.Next() {
		record, err := dao.scanExamRecordRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
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
	query := `UPDATE exam_records 
              SET category=?, exam_date=?, duration_seconds=?, user_id=?, score=?, total_questions=?, correct_count=? 
              WHERE id=?`

	_, err := dao.ExecuteUpdate(query,
		record.Category, record.ExamDate, record.DurationSeconds,
		record.UserID, record.Score, record.TotalQuestions, record.CorrectCount,
		record.ID)
	if err != nil {
		return err
	}

	utils.Debug("ExamRecordDAO", "更新考试记录成功", map[string]interface{}{
		"exam_id": record.ID,
	})

	return nil
}

// ClearByUser 清空用户的考试记录
// Python 原版：clear_user_data 中的 DELETE FROM exam_records WHERE user_id = ?
func (dao *ExamRecordDAO) ClearByUser(userID int64) error {
	utils.Info("ExamRecordDAO", "清空用户考试记录", map[string]interface{}{
		"user_id": userID,
	})

	// 先删除该用户所有考试记录的详情
	deleteDetailsQuery := `DELETE FROM exam_question_details WHERE exam_id IN (SELECT id FROM exam_records WHERE user_id = ?)`
	_, err := dao.ExecuteUpdate(deleteDetailsQuery, userID)
	if err != nil {
		return err
	}

	// 删除该用户的所有考试记录
	deleteQuery := `DELETE FROM exam_records WHERE user_id = ?`
	_, err = dao.ExecuteUpdate(deleteQuery, userID)
	if err != nil {
		return err
	}

	utils.Info("ExamRecordDAO", "清空用户考试记录成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// GetExamStatistics 获取考试统计数据（用于统计图表）
// Python 原版：exam_statistics_dao.get_exam_data(user_id, category, time_range)
func (dao *ExamRecordDAO) GetExamStatistics(userID int64, category string, startDate time.Time) ([]*models.ExamStatisticsData, error) {
	utils.Debug("ExamRecordDAO", "获取考试统计数据", map[string]interface{}{
		"user_id":    userID,
		"category":   category,
		"start_date": startDate,
	})

	// 构建 SQL 查询
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

	// 添加时间范围条件
	if !startDate.IsZero() {
		query += " AND er.exam_date >= ?"
		params = append(params, startDate)
	}

	// 添加类别筛选
	if category != "" {
		query += " AND er.category = ?"
		params = append(params, category)
	}

	query += `
		GROUP BY er.id
		ORDER BY er.exam_date ASC
	`

	rows, err := dao.ExecuteQuery(query, params...)
	if err != nil {
		utils.Error("ExamRecordDAO", "获取考试统计数据失败", err, nil)
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
			utils.Error("ExamRecordDAO", "扫描考试数据失败", err, nil)
			continue
		}

		data.TotalQuestions = totalQuestions
		data.CorrectQuestions = correctQuestions
		data.PassRate = passRate * 100 // 转换为百分比
		data.DurationSeconds = durationSeconds
		data.Score = float64(correctQuestions) // 每题 1 分

		examData = append(examData, data)
	}

	utils.Debug("ExamRecordDAO", "获取考试统计数据成功", map[string]interface{}{
		"count": len(examData),
	})

	return examData, nil
}

// GetCount 获取考试记录总数
func (dao *ExamRecordDAO) GetCount() (int, error) {
	query := `SELECT COUNT(*) FROM exam_records`

	row := dao.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCountByUser 获取用户的考试记录总数
func (dao *ExamRecordDAO) GetCountByUser(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM exam_records WHERE user_id = ?`

	row := dao.QueryRow(query, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// scanExamRecord 扫描单行数据
func (dao *ExamRecordDAO) scanExamRecord(row Scanner) (*models.ExamRecord, error) {
	record := &models.ExamRecord{}
	err := row.Scan(
		&record.ID, &record.Category, &record.ExamDate, &record.DurationSeconds,
		&record.UserID, &record.Score, &record.TotalQuestions, &record.CorrectCount)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// scanExamRecordRow 扫描多行数据中的一行
func (dao *ExamRecordDAO) scanExamRecordRow(rows *sql.Rows) (*models.ExamRecord, error) {
	record := &models.ExamRecord{}
	err := rows.Scan(
		&record.ID, &record.Category, &record.ExamDate, &record.DurationSeconds,
		&record.UserID, &record.Score, &record.TotalQuestions, &record.CorrectCount)
	if err != nil {
		return nil, err
	}
	return record, nil
}
