package models

// FavoriteQuestion 收藏题目实体
type FavoriteQuestion struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	QuestionID int64  `json:"question_id" gorm:"column:question_id"` // 题目 ID
	Category   string `json:"category" gorm:"column:category"`       // 题目分类
	UserID     int64  `json:"user_id" gorm:"column:user_id"`         // 用户 ID
}

// TableName 指定表名
func (FavoriteQuestion) TableName() string {
	return "favorite_questions"
}

// ToMap 转换为 map
func (f *FavoriteQuestion) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          f.ID,
		"question_id": f.QuestionID,
		"category":    f.Category,
		"user_id":     f.UserID,
	}
}

// FromMap 从 map 创建
func (f *FavoriteQuestion) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		f.ID = v
	}
	if v, ok := data["question_id"].(int64); ok {
		f.QuestionID = v
	}
	if v, ok := data["category"].(string); ok {
		f.Category = v
	}
	if v, ok := data["user_id"].(int64); ok {
		f.UserID = v
	}
}
