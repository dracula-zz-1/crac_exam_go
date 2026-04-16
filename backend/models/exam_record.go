package models

import "time"

// ExamRecord 考试记录实体
type ExamRecord struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	Category        string    `json:"category" gorm:"column:category"`                 // 考试类别 A/B/C
	ExamDate        time.Time `json:"exam_date" gorm:"column:exam_date"`               // 考试日期
	DurationSeconds float64   `json:"duration_seconds" gorm:"column:duration_seconds"` // 考试时长 (秒)
	UserID          int64     `json:"user_id" gorm:"column:user_id"`                   // 用户 ID
	Score           int       `json:"score" gorm:"column:score"`                       // 成绩（答对题数，每题 1 分）
	TotalQuestions  int       `json:"total_questions" gorm:"column:total_questions"`   // 总题数
	CorrectCount    int       `json:"correct_count" gorm:"column:correct_count"`       // 正确题数
}

// TableName 指定表名
func (ExamRecord) TableName() string {
	return "exam_records"
}

// ToMap 转换为 map
func (e *ExamRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"category":         e.Category,
		"exam_date":        e.ExamDate,
		"duration_seconds": e.DurationSeconds,
		"user_id":          e.UserID,
		"score":            e.Score,
		"total_questions":  e.TotalQuestions,
		"correct_count":    e.CorrectCount,
	}
}

// FromMap 从 map 创建
func (e *ExamRecord) FromMap(data map[string]interface{}) {
	if v, ok := data["id"].(int64); ok {
		e.ID = v
	}
	if v, ok := data["category"].(string); ok {
		e.Category = v
	}
	if v, ok := data["score"].(int); ok {
		e.Score = v
	}
	if v, ok := data["total_questions"].(int); ok {
		e.TotalQuestions = v
	}
	if v, ok := data["correct_count"].(int); ok {
		e.CorrectCount = v
	}
	if v, ok := data["user_id"].(int64); ok {
		e.UserID = v
	}
}
