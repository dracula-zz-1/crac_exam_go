package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"

	"gorm.io/gorm"
)

// FavoriteQuestionDAO 收藏题目数据访问对象
type FavoriteQuestionDAO struct {
	*BaseDAO
}

// NewFavoriteQuestionDAO 创建 FavoriteQuestionDAO 实例
func NewFavoriteQuestionDAO(db *gorm.DB) *FavoriteQuestionDAO {
	return &FavoriteQuestionDAO{
		BaseDAO: NewBaseDAO(db, "favorite_questions"),
	}
}

// Create 创建收藏题目
func (dao *FavoriteQuestionDAO) Create(favorite *models.FavoriteQuestion) (int64, error) {
	result := dao.db.Create(favorite)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "创建收藏题目成功", map[string]interface{}{
		"favorite_id": favorite.ID,
		"question_id": favorite.QuestionID,
		"user_id":     favorite.UserID,
		"category":    favorite.Category,
	})

	return favorite.ID, nil
}

// GetByID 根据 ID 获取收藏
func (dao *FavoriteQuestionDAO) GetByID(id int64) (*models.FavoriteQuestion, error) {
	favorite := &models.FavoriteQuestion{}
	result := dao.db.First(favorite, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return favorite, nil
}

// GetByUserID 根据用户 ID 获取所有收藏
func (dao *FavoriteQuestionDAO) GetByUserID(userID int64) ([]*models.FavoriteQuestion, error) {
	var favorites []*models.FavoriteQuestion
	result := dao.db.Where("user_id = ?", userID).Find(&favorites)
	if result.Error != nil {
		return nil, result.Error
	}

	return favorites, nil
}

// GetByUserAndQuestion 根据用户 ID 和题目 ID 获取收藏
func (dao *FavoriteQuestionDAO) GetByUserAndQuestion(userID int64, questionID int64) (*models.FavoriteQuestion, error) {
	favorite := &models.FavoriteQuestion{}
	result := dao.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(favorite)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return favorite, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取收藏
func (dao *FavoriteQuestionDAO) GetByUserAndCategory(userID int64, category string) ([]*models.FavoriteQuestion, error) {
	var favorites []*models.FavoriteQuestion
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Find(&favorites)
	if result.Error != nil {
		return nil, result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户类别收藏成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(favorites),
	})

	return favorites, nil
}

// Delete 删除收藏
func (dao *FavoriteQuestionDAO) Delete(id int64) error {
	result := dao.db.Delete(&models.FavoriteQuestion{}, id)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("FavoriteQuestionDAO", "删除收藏成功", map[string]interface{}{
		"favorite_id": id,
	})

	return nil
}

// DeleteByUserAndQuestion 根据用户 ID 和题目 ID 删除收藏
func (dao *FavoriteQuestionDAO) DeleteByUserAndQuestion(userID int64, questionID int64) error {
	result := dao.db.Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.FavoriteQuestion{})
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "删除用户收藏成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	return nil
}

// GetCountByUser 获取用户收藏数量
func (dao *FavoriteQuestionDAO) GetCountByUser(userID int64) (int64, error) {
	var count int64
	result := dao.db.Model(&models.FavoriteQuestion{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户收藏数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// GetCountByUserAndCategory 获取用户指定类别的收藏数量
func (dao *FavoriteQuestionDAO) GetCountByUserAndCategory(userID int64, category string) (int64, error) {
	var count int64
	result := dao.db.Model(&models.FavoriteQuestion{}).Where("user_id = ? AND category = ?", userID, category).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户类别收藏数量成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    count,
	})

	return count, nil
}

// GetCount 获取收藏总数
func (dao *FavoriteQuestionDAO) GetCount() (int64, error) {
	var count int64
	result := dao.db.Model(&models.FavoriteQuestion{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Debug("FavoriteQuestionDAO", "获取收藏总数成功", map[string]interface{}{
		"count": count,
	})

	return count, nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除收藏
func (dao *FavoriteQuestionDAO) DeleteByUserAndCategory(userID int64, category string) error {
	result := dao.db.Where("user_id = ? AND category = ?", userID, category).Delete(&models.FavoriteQuestion{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("FavoriteQuestionDAO", "删除用户类别收藏成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的收藏
func (dao *FavoriteQuestionDAO) ClearByUser(userID int64) error {
	result := dao.db.Where("user_id = ?", userID).Delete(&models.FavoriteQuestion{})
	if result.Error != nil {
		return result.Error
	}

	utils.Info("FavoriteQuestionDAO", "清空用户收藏成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}
