package services

import "gorm.io/gorm"

var dbInstance *gorm.DB

// SetDB 设置数据库实例（由 main.go 初始化时调用）
func SetDB(db *gorm.DB) {
	dbInstance = db
}

// GetDB 获取 GORM 数据库实例
func GetDB() *gorm.DB {
	return dbInstance
}
