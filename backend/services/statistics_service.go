package services

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
)

type StatisticsService struct {
	db *sql.DB
}

func NewStatisticsService(db *sql.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

func (s *StatisticsService) ServiceName() string { return "StatisticsService" }

// GetDashboard 首页聚合数据：一次调用返回全部统计。
func (s *StatisticsService) GetDashboard(userID int64) (*models.Dashboard, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("请先登录")
	}
	user, err := store.GetUser(s.db, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	categories, err := store.CategoryCounts(s.db)
	if err != nil {
		return nil, err
	}
	totalQuestions, err := store.CountQuestions(s.db)
	if err != nil {
		return nil, err
	}
	withImages, err := store.CountImages(s.db)
	if err != nil {
		return nil, err
	}

	totalExams, passedExams, avgScore, err := store.ExamTotals(s.db, userID)
	if err != nil {
		return nil, err
	}
	bestScores, err := store.BestScores(s.db, userID)
	if err != nil {
		return nil, err
	}
	recent, err := store.ListExamSummaries(s.db, userID, 8)
	if err != nil {
		return nil, err
	}
	errorCount, err := store.CountErrors(s.db, userID)
	if err != nil {
		return nil, err
	}
	favoriteCount, err := store.CountFavorites(s.db, userID)
	if err != nil {
		return nil, err
	}
	totalDuration, err := store.ExamTotalDuration(s.db, userID)
	if err != nil {
		return nil, err
	}

	progressMap, err := store.AllProgress(s.db, userID)
	if err != nil {
		return nil, err
	}
	progress := make([]models.CategoryProgress, 0, 3)
	for _, cat := range config.Categories() {
		count, err := store.QuestionCountByCategory(s.db, cat)
		if err != nil {
			return nil, err
		}
		p := progressMap[cat]
		p.Category = cat
		p.Total = count
		if p.Answered > count {
			p.Answered = count
		}
		if p.Answered < 0 {
			p.Answered = 0
		}
		if count > 0 {
			p.Percent = int(math.Round(float64(p.Answered) / float64(count) * 100))
		}
		progress = append(progress, p)
	}

	passRate := 0.0
	if totalExams > 0 {
		passRate = float64(passedExams) / float64(totalExams) * 100
	}
	var lastExam *models.ExamSummary
	if len(recent) > 0 {
		lastExam = &recent[0]
	}

	return &models.Dashboard{
		User:             user,
		TotalQuestions:   totalQuestions,
		WithImages:       withImages,
		Categories:       categories,
		TotalExams:       totalExams,
		PassedExams:      passedExams,
		PassRate:         math.Round(passRate*10) / 10,
		BestScores:       bestScores,
		RecentExams:      recent,
		ErrorCount:       errorCount,
		FavoriteCount:    favoriteCount,
		Practice:         progress,
		LastExam:         lastExam,
		AverageScore:     math.Round(avgScore*10) / 10,
		TotalStudyMinute: int(totalDuration / 60),
	}, nil
}

// GetScoreSeries 成绩趋势。days <= 0 表示全部时间。
func (s *StatisticsService) GetScoreSeries(userID int64, category string, days int) ([]models.ScorePoint, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("请先登录")
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	since := time.Time{}
	if days > 0 {
		since = time.Now().AddDate(0, 0, -days)
	}
	return store.ScorePoints(s.db, userID, category, since)
}
