package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"

	"gorm.io/gorm"
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

	// 创建复合索引以优化查询性能
	utils.Info("Database", "开始创建数据库索引", nil)
	createIndexes(db)

	utils.Info("Database", "数据库表结构初始化完成", nil)
	return nil
}

// createIndexes 创建复合索引以优化查询性能
func createIndexes(db *gorm.DB) {
	// error_questions 表的 user_id + category 复合索引
	if !db.Migrator().HasIndex(&models.ErrorQuestion{}, "idx_error_user_category") {
		if err := db.Exec("CREATE INDEX idx_error_user_category ON error_questions(user_id, category)").Error; err != nil {
			utils.Warn("Database", "创建错题索引失败", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			utils.Info("Database", "错题表索引创建成功", nil)
		}
	}

	// favorite_questions 表的 user_id + category 复合索引
	if !db.Migrator().HasIndex(&models.FavoriteQuestion{}, "idx_favorite_user_category") {
		if err := db.Exec("CREATE INDEX idx_favorite_user_category ON favorite_questions(user_id, category)").Error; err != nil {
			utils.Warn("Database", "创建收藏索引失败", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			utils.Info("Database", "收藏表索引创建成功", nil)
		}
	}

	// exam_records 表的 user_id + category 复合索引
	if !db.Migrator().HasIndex(&models.ExamRecord{}, "idx_exam_user_category") {
		if err := db.Exec("CREATE INDEX idx_exam_user_category ON exam_records(user_id, category)").Error; err != nil {
			utils.Warn("Database", "创建考试记录索引失败", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			utils.Info("Database", "考试记录表索引创建成功", nil)
		}
	}

	// practice_progress 表的 user_id + category 复合索引
	if !db.Migrator().HasIndex(&models.PracticeProgress{}, "idx_practice_user_category") {
		if err := db.Exec("CREATE INDEX idx_practice_user_category ON practice_progress(user_id, category)").Error; err != nil {
			utils.Warn("Database", "创建练习进度索引失败", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			utils.Info("Database", "练习进度表索引创建成功", nil)
		}
	}
}

// ResetQuestionsTable 重置题库表 (删除并重新创建)
func ResetQuestionsTable() error {
	utils.Info("Database", "开始重置题库表", nil)

	db, err := GetDB()
	if err != nil {
		utils.Error("Database", "获取数据库实例失败", err, nil)
		return err
	}

	// 清理关联的孤立数据（外键约束：题目被删除时，错题/收藏也应清理）
	orphanErrCount := 0

	// 删除所有错题记录（因为题目将全部被删除）
	if err := db.Where("1=1").Delete(&models.ErrorQuestion{}).Error; err != nil {
		utils.Error("Database", "清理错题记录失败", err, nil)
		orphanErrCount++
	} else {
		utils.Info("Database", "已清理全部错题记录", nil)
	}

	// 删除所有收藏记录
	if err := db.Where("1=1").Delete(&models.FavoriteQuestion{}).Error; err != nil {
		utils.Error("Database", "清理收藏记录失败", err, nil)
		orphanErrCount++
	} else {
		utils.Info("Database", "已清理全部收藏记录", nil)
	}

	// 删除该用户的练习进度记录
	if err := db.Where("1=1").Delete(&models.PracticeProgress{}).Error; err != nil {
		utils.Error("Database", "清理练习进度记录失败", err, nil)
		orphanErrCount++
	} else {
		utils.Info("Database", "已清理全部练习进度记录", nil)
	}

	// 删除该用户的考试记录详情
	if err := db.Where("1=1").Delete(&models.ExamQuestionDetail{}).Error; err != nil {
		utils.Error("Database", "清理考试详情记录失败", err, nil)
		orphanErrCount++
	} else {
		utils.Info("Database", "已清理全部考试详情记录", nil)
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

	if orphanErrCount > 0 {
		utils.Info("Database", "题库表重置完成（部分关联数据清理失败）", map[string]interface{}{
			"error_count": orphanErrCount,
		})
	} else {
		utils.Info("Database", "题库表重置完成（关联数据已清理）", nil)
	}
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
