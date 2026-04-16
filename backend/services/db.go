package services

import (
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"log"

	"gorm.io/gorm"
)

var (
	dbInstance *sql.DB
	gormDB     *gorm.DB
)

// InitDB 初始化数据库连接并创建表
func InitDB() *sql.DB {
	if dbInstance != nil {
		return dbInstance
	}

	var err error
	gormDB, err = dao.GetDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	dbInstance, err = gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	// 自动创建数据库表
	utils.Info("Database", "开始创建数据库表", nil)
	err = gormDB.AutoMigrate(
		&models.User{},
		&models.Question{},
		&models.ExamRecord{},
		&models.ExamQuestionDetail{},
		&models.ErrorQuestion{},
		&models.FavoriteQuestion{},
		&models.PracticeProgress{},
	)
	if err != nil {
		log.Fatal("Failed to create database tables:", err)
	}
	utils.Info("Database", "数据库表创建成功", nil)

	return dbInstance
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	if dbInstance == nil {
		return InitDB()
	}
	return dbInstance
}

// GetGormDB 获取 GORM 数据库实例
func GetGormDB() *gorm.DB {
	if gormDB == nil {
		InitDB()
	}
	return gormDB
}
