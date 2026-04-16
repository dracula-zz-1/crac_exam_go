package models

import "time"

// PracticeProgress 练习进度实体
type PracticeProgress struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Category     string    `json:"category" gorm:"column:category"`           // 题目分类
	CurrentIndex int       `json:"current_index" gorm:"column:current_index"` // 当前索引
	LastAccessed time.Time `json:"last_accessed" gorm:"column:last_accessed"` // 最后访问时间
	UserID       int64     `json:"user_id" gorm:"column:user_id"`             // 用户 ID
}

// TableName 指定表名
func (PracticeProgress) TableName() string {
	return "practice_progress"
}

// ToMap 转换为 map
func (p *PracticeProgress) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            p.ID,
		"category":      p.Category,
		"current_index": p.CurrentIndex,
		"last_accessed": p.LastAccessed,
		"user_id":       p.UserID,
	}
}

// FromMap 从 map 创建
func (p *PracticeProgress) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		p.ID = v
	}
	if v, ok := data["category"].(string); ok {
		p.Category = v
	}
	if v, ok := data["current_index"].(int); ok {
		p.CurrentIndex = v
	}
	if v, ok := data["user_id"].(int64); ok {
		p.UserID = v
	}
}
