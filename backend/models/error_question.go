package models

// ErrorQuestion 错题实体（只存储引用信息）
type ErrorQuestion struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	QuestionID int64  `json:"question_id" gorm:"column:question_id"` // 题目 ID
	Category   string `json:"category" gorm:"column:category"`       // 题目分类
	UserID     int64  `json:"user_id" gorm:"column:user_id"`         // 用户 ID
	CreatedAt  string `json:"created_at" gorm:"column:created_at"`   // 创建时间
	// 以下是通过 JOIN 查询获取的题目字段（不存储到数据库）
	J    string `json:"J,omitempty" gorm:"-"`
	P    string `json:"P,omitempty" gorm:"-"`
	I    string `json:"I,omitempty" gorm:"-"`
	Q    string `json:"Q,omitempty" gorm:"-"`
	T    string `json:"T,omitempty" gorm:"-"`
	A    string `json:"A,omitempty" gorm:"-"`
	B    string `json:"B,omitempty" gorm:"-"`
	C    string `json:"C,omitempty" gorm:"-"`
	D    string `json:"D,omitempty" gorm:"-"`
	F    string `json:"F,omitempty" gorm:"-"`
	LA   int    `json:"LA,omitempty" gorm:"-"`
	LB   int    `json:"LB,omitempty" gorm:"-"`
	LC   int    `json:"LC,omitempty" gorm:"-"`
	Type int    `json:"type,omitempty" gorm:"-"`
}

// TableName 指定表名
func (ErrorQuestion) TableName() string {
	return "error_questions"
}

// ToMap 转换为 map
func (e *ErrorQuestion) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"question_id": e.QuestionID,
		"category":    e.Category,
		"user_id":     e.UserID,
		"created_at":  e.CreatedAt,
	}
}

// FromMap 从 map 创建
func (e *ErrorQuestion) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		e.ID = v
	}
	if v, ok := data["question_id"].(int64); ok {
		e.QuestionID = v
	}
	if v, ok := data["category"].(string); ok {
		e.Category = v
	}
	if v, ok := data["user_id"].(int64); ok {
		e.UserID = v
	}
	if v, ok := data["created_at"].(string); ok {
		e.CreatedAt = v
	}
}
