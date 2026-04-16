package models

import "time"

// User 用户实体
type User struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"column:username"`
	IDCard    string    `json:"id_card" gorm:"column:id_card"`
	LastLogin time.Time `json:"last_login" gorm:"column:last_login"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// ToMap 转换为 map
func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"id_card":    u.IDCard,
		"last_login": u.LastLogin,
	}
}

// FromMap 从 map 创建
func (u *User) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		u.ID = v
	}
	if v, ok := data["username"].(string); ok {
		u.Username = v
	}
	if v, ok := data["id_card"].(string); ok {
		u.IDCard = v
	}
}
