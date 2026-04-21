package dao

import (
	"context"
	"crac_exam_go/backend/utils"
	"fmt"

	"gorm.io/gorm"
)

// BaseDAO 通用数据访问对象基类
type BaseDAO struct {
	db        *gorm.DB
	tableName string
}

// NewBaseDAO 创建 BaseDAO 实例
func NewBaseDAO(db *gorm.DB, tableName string) *BaseDAO {
	return &BaseDAO{
		db:        db,
		tableName: tableName,
	}
}

// ExecuteQuery 执行查询 SQL
func (dao *BaseDAO) ExecuteQuery(query string, args ...interface{}) (*gorm.DB, error) {
	utils.Debug("DAO", fmt.Sprintf("执行查询：%s", query), map[string]interface{}{
		"args":  args,
		"table": dao.tableName,
	})

	rows := dao.db.Raw(query, args...)
	if rows.Error != nil {
		utils.Error("DAO", "执行查询失败", rows.Error, map[string]interface{}{
			"query": query,
			"args":  args,
			"table": dao.tableName,
		})
		return nil, rows.Error
	}

	return rows, nil
}

// ExecuteUpdate 执行更新 SQL（INSERT/UPDATE/DELETE）
func (dao *BaseDAO) ExecuteUpdate(query string, args ...interface{}) (*gorm.DB, error) {
	utils.Debug("DAO", fmt.Sprintf("执行更新：%s", query), map[string]interface{}{
		"args":  args,
		"table": dao.tableName,
	})

	result := dao.db.Exec(query, args...)
	if result.Error != nil {
		utils.Error("DAO", "执行更新失败", result.Error, map[string]interface{}{
			"query": query,
			"args":  args,
			"table": dao.tableName,
		})
		return nil, result.Error
	}

	return result, nil
}

// GetDB 获取数据库实例
func (dao *BaseDAO) GetDB() *gorm.DB {
	return dao.db
}

// GetLastInsertID 获取最后插入的 ID
func (dao *BaseDAO) GetLastInsertID(result *gorm.DB) (int64, error) {
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// QueryRow 查询单行
func (dao *BaseDAO) QueryRow(query string, args ...interface{}) *gorm.DB {
	return dao.db.Raw(query, args...)
}

// Begin 开启事务
func (dao *BaseDAO) Begin() *gorm.DB {
	return dao.db.Begin()
}

// WithContext 设置上下文
func (dao *BaseDAO) WithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}
