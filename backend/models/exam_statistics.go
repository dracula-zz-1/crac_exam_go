package models

import "time"

// ExamStatisticsData 考试统计数据
type ExamStatisticsData struct {
	ID               int64     `json:"id" gorm:"-"`
	Category         string    `json:"category" gorm:"-"`
	ExamDate         time.Time `json:"exam_date" gorm:"-"`
	TotalQuestions   int       `json:"total_questions" gorm:"-"`
	CorrectQuestions int       `json:"correct_questions" gorm:"-"`
	PassRate         float64   `json:"pass_rate" gorm:"-"`
	DurationSeconds  float64   `json:"duration_seconds" gorm:"-"`
	Score            float64   `json:"score" gorm:"-"`
}
