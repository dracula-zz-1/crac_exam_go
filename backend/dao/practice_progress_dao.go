package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"time"

	"gorm.io/gorm"
)

// PracticeProgressDAO 练习进度数据访问对象
type PracticeProgressDAO struct {
	*BaseDAO
}

// NewPracticeProgressDAO 创建 PracticeProgressDAO 实例
func NewPracticeProgressDAO(db *gorm.DB) *PracticeProgressDAO {
	return &PracticeProgressDAO{
		BaseDAO: NewBaseDAO(db, "practice_progress"),
	}
}

// Create 创建练习进度
func (dao *PracticeProgressDAO) Create(progress *models.PracticeProgress) (int64, error) {
	result := dao.db.Create(progress)
	if result.Error != nil {
		return 0, result.Error
	}

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
	progress := &models.PracticeProgress{}
	result := dao.db.First(progress, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return progress, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取练习进度
func (dao *PracticeProgressDAO) GetByUserAndCategory(userID int64, category string) (*models.PracticeProgress, error) {
	progress := &models.PracticeProgress{}
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).First(progress)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
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
	var progresses []*models.PracticeProgress
	result := dao.db.Where("user_id = ?", userID).Find(&progresses)
	if result.Error != nil {
		return nil, result.Error
	}

	return progresses, nil
}

// Update 更新练习进度
func (dao *PracticeProgressDAO) Update(progress *models.PracticeProgress) error {
	result := dao.db.Save(progress)
	if result.Error != nil {
		return result.Error
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
		result := dao.db.Model(&models.PracticeProgress{}).
			Where("user_id = ? AND category = ?", userID, category).
			Updates(map[string]interface{}{
				"current_index": currentIndex,
				"last_accessed": lastAccessed,
			})
		if result.Error != nil {
			return result.Error
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
func (dao *PracticeProgressDAO) GetCountByUser(userID int64) (int64, error) {
	var count int64
	result := dao.db.Model(&models.PracticeProgress{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("PracticeProgressDAO", "获取用户练习进度数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// Delete 删除练习进度
func (dao *PracticeProgressDAO) Delete(id int64) error {
	result := dao.db.Delete(&models.PracticeProgress{}, id)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("PracticeProgressDAO", "删除练习进度成功", map[string]interface{}{
		"progress_id": id,
	})

	return nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除练习进度
func (dao *PracticeProgressDAO) DeleteByUserAndCategory(userID int64, category string) error {
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Delete(&models.PracticeProgress{})
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("PracticeProgressDAO", "删除用户练习进度成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的练习进度
func (dao *PracticeProgressDAO) ClearByUser(userID int64) error {
	result := dao.db.Where("user_id = ?", userID).Delete(&models.PracticeProgress{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("PracticeProgressDAO", "清空用户练习进度成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}
