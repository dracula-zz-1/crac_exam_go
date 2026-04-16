package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// PracticeProgressDAO 练习进度数据访问对象
type PracticeProgressDAO struct {
	*BaseDAO
}

// NewPracticeProgressDAO 创建 PracticeProgressDAO 实例
func NewPracticeProgressDAO(db *sql.DB) *PracticeProgressDAO {
	return &PracticeProgressDAO{
		BaseDAO: NewBaseDAO(db, "practice_progress"),
	}
}

// Create 创建练习进度
func (dao *PracticeProgressDAO) Create(progress *models.PracticeProgress) (int64, error) {
	query := `INSERT INTO practice_progress (category, current_index, last_accessed, user_id) VALUES (?, ?, ?, ?)`

	result, err := dao.ExecuteUpdate(query, progress.Category, progress.CurrentIndex, progress.LastAccessed, progress.UserID)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	progress.ID = id
	utils.Debug("PracticeProgressDAO", "创建练习进度成功", map[string]interface{}{
		"progress_id":   progress.ID,
		"category":      progress.Category,
		"user_id":       progress.UserID,
		"current_index": progress.CurrentIndex,
	})

	return progress.ID, nil
}

// GetByID 根据 ID 获取练习进度
func (dao *PracticeProgressDAO) GetByID(id int64) (*models.PracticeProgress, error) {
	query := `SELECT id, category, current_index, last_accessed, user_id FROM practice_progress WHERE id = ?`

	row := dao.QueryRow(query, id)
	progress, err := dao.scanPracticeProgress(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return progress, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取练习进度
func (dao *PracticeProgressDAO) GetByUserAndCategory(userID int64, category string) (*models.PracticeProgress, error) {
	query := `SELECT id, category, current_index, last_accessed, user_id FROM practice_progress WHERE user_id = ? AND category = ?`

	row := dao.QueryRow(query, userID, category)
	progress, err := dao.scanPracticeProgress(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	utils.Debug("PracticeProgressDAO", "获取用户练习进度成功", map[string]interface{}{
		"user_id":     userID,
		"category":    category,
		"progress_id": progress.ID,
	})

	return progress, nil
}

// GetAllByUser 获取用户的所有练习进度
func (dao *PracticeProgressDAO) GetAllByUser(userID int64) ([]*models.PracticeProgress, error) {
	query := `SELECT id, category, current_index, last_accessed, user_id FROM practice_progress WHERE user_id = ?`

	rows, err := dao.ExecuteQuery(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progresses []*models.PracticeProgress
	for rows.Next() {
		progress, err := dao.scanPracticeProgressRow(rows)
		if err != nil {
			return nil, err
		}
		progresses = append(progresses, progress)
	}

	return progresses, nil
}

// Update 更新练习进度
func (dao *PracticeProgressDAO) Update(progress *models.PracticeProgress) error {
	query := `UPDATE practice_progress SET category=?, current_index=?, last_accessed=?, user_id=? WHERE id=?`

	_, err := dao.ExecuteUpdate(query, progress.Category, progress.CurrentIndex, progress.LastAccessed, progress.UserID, progress.ID)
	if err != nil {
		return err
	}

	utils.Debug("PracticeProgressDAO", "更新练习进度成功", map[string]interface{}{
		"progress_id": progress.ID,
	})

	return nil
}

// UpdateByUserAndCategory 根据用户 ID 和类别更新练习进度
func (dao *PracticeProgressDAO) UpdateByUserAndCategory(userID int64, category string, currentIndex int, lastAccessed time.Time) error {
	// 先检查是否存在
	existing, err := dao.GetByUserAndCategory(userID, category)
	if err != nil {
		return err
	}

	if existing != nil {
		// 存在则更新
		query := `UPDATE practice_progress SET current_index=?, last_accessed=? WHERE user_id=? AND category=?`
		_, err := dao.ExecuteUpdate(query, currentIndex, lastAccessed, userID, category)
		if err != nil {
			return err
		}
		utils.Debug("PracticeProgressDAO", "更新练习进度成功", map[string]interface{}{
			"user_id":     userID,
			"category":    category,
			"progress_id": existing.ID,
		})
	} else {
		// 不存在则创建
		newProgress := &models.PracticeProgress{
			Category:     category,
			CurrentIndex: currentIndex,
			LastAccessed: lastAccessed,
			UserID:       userID,
		}
		_, err := dao.Create(newProgress)
		if err != nil {
			return err
		}
		utils.Debug("PracticeProgressDAO", "创建练习进度成功", map[string]interface{}{
			"user_id":  userID,
			"category": category,
		})
	}

	return nil
}

// GetCountByUser 获取用户练习进度记录数量
func (dao *PracticeProgressDAO) GetCountByUser(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM practice_progress WHERE user_id = ?`

	row := dao.QueryRow(query, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("PracticeProgressDAO", "获取用户练习进度数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// Delete 删除练习进度
func (dao *PracticeProgressDAO) Delete(id int64) error {
	query := `DELETE FROM practice_progress WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, id)
	if err != nil {
		return err
	}

	utils.Info("PracticeProgressDAO", "删除练习进度成功", map[string]interface{}{
		"progress_id": id,
	})

	return nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除练习进度
func (dao *PracticeProgressDAO) DeleteByUserAndCategory(userID int64, category string) error {
	query := `DELETE FROM practice_progress WHERE user_id = ? AND category = ?`

	_, err := dao.ExecuteUpdate(query, userID, category)
	if err != nil {
		return err
	}

	utils.Debug("PracticeProgressDAO", "删除用户练习进度成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的练习进度
func (dao *PracticeProgressDAO) ClearByUser(userID int64) error {
	query := `DELETE FROM practice_progress WHERE user_id = ?`

	_, err := dao.ExecuteUpdate(query, userID)
	if err != nil {
		return err
	}

	utils.Info("PracticeProgressDAO", "清空用户练习进度成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// scanPracticeProgress 扫描单行数据
func (dao *PracticeProgressDAO) scanPracticeProgress(row Scanner) (*models.PracticeProgress, error) {
	progress := &models.PracticeProgress{}
	err := row.Scan(&progress.ID, &progress.Category, &progress.CurrentIndex, &progress.LastAccessed, &progress.UserID)
	if err != nil {
		return nil, err
	}
	return progress, nil
}

// scanPracticeProgressRow 扫描多行数据中的一行
func (dao *PracticeProgressDAO) scanPracticeProgressRow(rows *sql.Rows) (*models.PracticeProgress, error) {
	progress := &models.PracticeProgress{}
	err := rows.Scan(&progress.ID, &progress.Category, &progress.CurrentIndex, &progress.LastAccessed, &progress.UserID)
	if err != nil {
		return nil, err
	}
	return progress, nil
}
