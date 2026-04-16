package models

// Question 题目实体
type Question struct {
	ID     int64  `json:"id" gorm:"primaryKey"`
	J      string `json:"J" gorm:"column:J"`             // 编号 1
	P      string `json:"P" gorm:"column:P"`             // 编号 2
	I      string `json:"I" gorm:"column:I"`             // 编号 3
	Q      string `json:"Q" gorm:"column:Q"`             // 题干
	T      string `json:"T" gorm:"column:T"`             // 答案
	A      string `json:"A" gorm:"column:A"`             // 选项 A
	B      string `json:"B" gorm:"column:B"`             // 选项 B
	C      string `json:"C" gorm:"column:C"`             // 选项 C
	D      string `json:"D" gorm:"column:D"`             // 选项 D
	F      string `json:"F" gorm:"column:F"`             // 图片 (Base64)
	LA     int    `json:"LA" gorm:"column:LA"`           // A 类题库标记
	LB     int    `json:"LB" gorm:"column:LB"`           // B 类题库标记
	LC     int    `json:"LC" gorm:"column:LC"`           // C 类题库标记
	Type   int    `json:"type" gorm:"column:type"`       // 题型：1 单选题，2 多选题
	UserID int64  `json:"user_id" gorm:"column:user_id"` // 用户 ID
	// 以下字段不存入数据库，仅用于前端显示
	TypeText string `json:"type_text" gorm:"-"` // 题型文字描述
	HasImage string `json:"has_image" gorm:"-"` // 是否有图片
}

// TableName 指定表名
func (Question) TableName() string {
	return "questions"
}

// ToMap 转换为 map
func (q *Question) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":      q.ID,
		"J":       q.J,
		"P":       q.P,
		"I":       q.I,
		"Q":       q.Q,
		"T":       q.T,
		"A":       q.A,
		"B":       q.B,
		"C":       q.C,
		"D":       q.D,
		"F":       q.F,
		"LA":      q.LA,
		"LB":      q.LB,
		"LC":      q.LC,
		"type":    q.Type,
		"user_id": q.UserID,
	}
}

// FromMap 从 map 创建
func (q *Question) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		q.ID = v
	}
	if v, ok := data["J"].(string); ok {
		q.J = v
	}
	if v, ok := data["P"].(string); ok {
		q.P = v
	}
	if v, ok := data["I"].(string); ok {
		q.I = v
	}
	if v, ok := data["Q"].(string); ok {
		q.Q = v
	}
	if v, ok := data["T"].(string); ok {
		q.T = v
	}
	if v, ok := data["A"].(string); ok {
		q.A = v
	}
	if v, ok := data["B"].(string); ok {
		q.B = v
	}
	if v, ok := data["C"].(string); ok {
		q.C = v
	}
	if v, ok := data["D"].(string); ok {
		q.D = v
	}
	if v, ok := data["F"].(string); ok {
		q.F = v
	}
	if v, ok := data["LA"].(int); ok {
		q.LA = v
	} else if v, ok := data["LA"].(float64); ok {
		q.LA = int(v)
	}
	if v, ok := data["LB"].(int); ok {
		q.LB = v
	} else if v, ok := data["LB"].(float64); ok {
		q.LB = int(v)
	}
	if v, ok := data["LC"].(int); ok {
		q.LC = v
	} else if v, ok := data["LC"].(float64); ok {
		q.LC = int(v)
	}
	if v, ok := data["type"].(int); ok {
		q.Type = v
	}
	if v, ok := data["user_id"].(int64); ok {
		q.UserID = v
	}
}

// QuestionType 题目类型
type QuestionType int

const (
	SingleChoice   QuestionType = 1 // 单选题
	MultipleChoice QuestionType = 2 // 多选题
)
