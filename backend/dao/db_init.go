package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
)

// InitDatabase 初始化所有数据库表
func InitDatabase() error {
	utils.Info("Database", "开始初始化数据库表结构", nil)

	db, err := GetDB()
	if err != nil {
		utils.Error("Database", "获取数据库实例失败", err, nil)
		return err
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.User{},
		&models.Question{},
		&models.ExamRecord{},
		&models.ExamQuestionDetail{},
		&models.ErrorQuestion{},
		&models.FavoriteQuestion{},
		&models.PracticeProgress{},
	)

	if err != nil {
		utils.Error("Database", "数据库表结构创建失败", err, nil)
		return err
	}

	utils.Info("Database", "数据库表结构初始化完成", nil)
	return nil
}

// ResetQuestionsTable 重置题库表 (删除并重新创建)
func ResetQuestionsTable() error {
	utils.Info("Database", "开始重置题库表", nil)

	db, err := GetDB()
	if err != nil {
		utils.Error("Database", "获取数据库实例失败", err, nil)
		return err
	}

	// 删除题库表
	err = db.Migrator().DropTable(&models.Question{})
	if err != nil {
		utils.Error("Database", "删除题库表失败", err, nil)
		return err
	}

	// 重新创建题库表
	err = db.AutoMigrate(&models.Question{})
	if err != nil {
		utils.Error("Database", "重新创建题库表失败", err, nil)
		return err
	}

	utils.Info("Database", "题库表重置完成", nil)
	return nil
}

// ClearUserData 清空指定用户的数据
func ClearUserData(userID int64) error {
	utils.Info("Database", "开始清空用户数据", map[string]interface{}{
		"user_id": userID,
	})

	db, err := GetDB()
	if err != nil {
		utils.Error("Database", "获取数据库实例失败", err, nil)
		return err
	}

	// 开启事务
	tx := db.Begin()

	// 删除该用户的错题
	if err := tx.Where("user_id = ?", userID).Delete(&models.ErrorQuestion{}).Error; err != nil {
		tx.Rollback()
		utils.Error("Database", "删除错题失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 删除该用户的练习进度
	if err := tx.Where("user_id = ?", userID).Delete(&models.PracticeProgress{}).Error; err != nil {
		tx.Rollback()
		utils.Error("Database", "删除练习进度失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 删除该用户的收藏
	if err := tx.Where("user_id = ?", userID).Delete(&models.FavoriteQuestion{}).Error; err != nil {
		tx.Rollback()
		utils.Error("Database", "删除收藏失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 获取该用户的所有考试记录
	var examRecords []models.ExamRecord
	if err := tx.Where("user_id = ?", userID).Find(&examRecords).Error; err != nil {
		tx.Rollback()
		utils.Error("Database", "查询考试记录失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 删除考试题目详情
	for _, record := range examRecords {
		if err := tx.Where("exam_id = ?", record.ID).Delete(&models.ExamQuestionDetail{}).Error; err != nil {
			tx.Rollback()
			utils.Error("Database", "删除考试题目详情失败", err, map[string]interface{}{
				"exam_id": record.ID,
			})
			return err
		}
	}

	// 删除考试记录
	if err := tx.Where("user_id = ?", userID).Delete(&models.ExamRecord{}).Error; err != nil {
		tx.Rollback()
		utils.Error("Database", "删除考试记录失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 提交事务
	tx.Commit()

	utils.Info("Database", "用户数据清空完成", map[string]interface{}{
		"user_id": userID,
	})
	return nil
}
