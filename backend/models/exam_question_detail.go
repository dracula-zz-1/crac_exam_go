package models

// ExamQuestionDetail 考试题目详情实体
type ExamQuestionDetail struct {
	ID            int64  `json:"id" gorm:"primaryKey"`
	ExamID        int64  `json:"exam_id" gorm:"column:exam_id"`               // 考试 ID
	QuestionID    int64  `json:"question_id" gorm:"column:question_id"`       // 题目 ID
	QuestionText  string `json:"question_text" gorm:"column:question_text"`   // 题目文本
	OptionA       string `json:"option_a" gorm:"column:option_a"`             // 选项 A
	OptionB       string `json:"option_b" gorm:"column:option_b"`             // 选项 B
	OptionC       string `json:"option_c" gorm:"column:option_c"`             // 选项 C
	OptionD       string `json:"option_d" gorm:"column:option_d"`             // 选项 D
	CorrectAnswer string `json:"correct_answer" gorm:"column:correct_answer"` // 正确答案
	UserAnswer    string `json:"user_answer" gorm:"column:user_answer"`       // 用户答案
	IsCorrect     bool   `json:"is_correct" gorm:"column:is_correct"`         // 是否正确
	Type          int    `json:"type" gorm:"column:type"`                     // 题目类型
	ImageData     string `json:"image_data" gorm:"column:image_data"`         // 图片数据
}

// TableName 指定表名
func (ExamQuestionDetail) TableName() string {
	return "exam_question_details"
}

// ToMap 转换为 map
func (e *ExamQuestionDetail) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"exam_id":        e.ExamID,
		"question_id":    e.QuestionID,
		"question_text":  e.QuestionText,
		"option_a":       e.OptionA,
		"option_b":       e.OptionB,
		"option_c":       e.OptionC,
		"option_d":       e.OptionD,
		"correct_answer": e.CorrectAnswer,
		"user_answer":    e.UserAnswer,
		"is_correct":     e.IsCorrect,
		"type":           e.Type,
		"image_data":     e.ImageData,
	}
}

// FromMap 从 map 创建
func (e *ExamQuestionDetail) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		e.ID = v
	}
	if v, ok := data["exam_id"].(int64); ok {
		e.ExamID = v
	}
	if v, ok := data["question_id"].(int64); ok {
		e.QuestionID = v
	}
	if v, ok := data["question_text"].(string); ok {
		e.QuestionText = v
	}
	if v, ok := data["option_a"].(string); ok {
		e.OptionA = v
	}
	if v, ok := data["option_b"].(string); ok {
		e.OptionB = v
	}
	if v, ok := data["option_c"].(string); ok {
		e.OptionC = v
	}
	if v, ok := data["option_d"].(string); ok {
		e.OptionD = v
	}
	if v, ok := data["correct_answer"].(string); ok {
		e.CorrectAnswer = v
	}
	if v, ok := data["user_answer"].(string); ok {
		e.UserAnswer = v
	}
	if v, ok := data["is_correct"].(bool); ok {
		e.IsCorrect = v
	}
	if v, ok := data["type"].(int); ok {
		e.Type = v
	}
	if v, ok := data["image_data"].(string); ok {
		e.ImageData = v
	}
}
