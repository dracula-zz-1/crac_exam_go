package dao

import (
	"crac_exam_go/backend/utils"
	"database/sql"
	"fmt"
)

// BaseDAO 通用数据访问对象基类
type BaseDAO struct {
	db        *sql.DB
	tableName string
}

// NewBaseDAO 创建 BaseDAO 实例
func NewBaseDAO(db *sql.DB, tableName string) *BaseDAO {
	return &BaseDAO{
		db:        db,
		tableName: tableName,
	}
}

// ExecuteQuery 执行查询 SQL
func (dao *BaseDAO) ExecuteQuery(query string, args ...interface{}) (*sql.Rows, error) {
	utils.Debug("DAO", fmt.Sprintf("执行查询：%s", query), map[string]interface{}{
		"args":  args,
		"table": dao.tableName,
	})

	rows, err := dao.db.Query(query, args...)
	if err != nil {
		utils.Error("DAO", "执行查询失败", err, map[string]interface{}{
			"query": query,
			"args":  args,
			"table": dao.tableName,
		})
		return nil, err
	}

	return rows, nil
}

// ExecuteUpdate 执行更新 SQL（INSERT/UPDATE/DELETE）
func (dao *BaseDAO) ExecuteUpdate(query string, args ...interface{}) (sql.Result, error) {
	utils.Debug("DAO", fmt.Sprintf("执行更新：%s", query), map[string]interface{}{
		"args":  args,
		"table": dao.tableName,
	})

	result, err := dao.db.Exec(query, args...)
	if err != nil {
		utils.Error("DAO", "执行更新失败", err, map[string]interface{}{
			"query": query,
			"args":  args,
			"table": dao.tableName,
		})
		return nil, err
	}

	return result, nil
}

// GetDB 获取数据库实例
func (dao *BaseDAO) GetDB() *sql.DB {
	return dao.db
}

// GetLastInsertID 获取最后插入的 ID
func (dao *BaseDAO) GetLastInsertID(result sql.Result) (int64, error) {
	return result.LastInsertId()
}

// QueryRow 查询单行
func (dao *BaseDAO) QueryRow(query string, args ...interface{}) *sql.Row {
	return dao.db.QueryRow(query, args...)
}
