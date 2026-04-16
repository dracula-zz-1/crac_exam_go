package dao

import (
	"crac_exam_go/backend/config"
	"crac_exam_go/backend/utils"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
	dbError    error
)

// GetDB 获取数据库实例 (单例)
func GetDB() (*gorm.DB, error) {
	dbOnce.Do(func() {
		dbPath := config.GetDatabasePath()
		utils.Info("Database", "初始化数据库连接", map[string]interface{}{
			"path": dbPath,
		})

		// 打开数据库连接
		dbInstance, dbError = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // 静默模式，我们自己管理日志
		})

		if dbError != nil {
			utils.Error("Database", "数据库连接失败", dbError, nil)
			return
		}

		utils.Info("Database", "数据库连接成功", nil)
	})

	return dbInstance, dbError
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if dbInstance != nil {
		sqlDB, err := dbInstance.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
