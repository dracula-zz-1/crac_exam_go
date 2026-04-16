package services

import (
	"crac_exam_go/backend/config"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// StatisticsService 统计服务
type StatisticsService struct {
	examRecordDAO *dao.ExamRecordDAO
	examDetailDAO *dao.ExamQuestionDetailDAO
	practiceDAO   *dao.PracticeProgressDAO
	errorDAO      *dao.ErrorQuestionDAO
	favoriteDAO   *dao.FavoriteQuestionDAO
	examConfig    map[string]config.ExamConfig
}

// ExamStatisticsResult 考试统计结果
type ExamStatisticsResult struct {
	MaxScore       float64 `json:"max_score"`
	LatestScore    float64 `json:"latest_score"`
	LatestDuration float64 `json:"latest_duration"`
	AvgPassRate    float64 `json:"avg_pass_rate"`
	TotalExams     int     `json:"total_exams"`
}

// UserStatisticsResult 用户统计结果
type UserStatisticsResult struct {
	TotalExams       int        `json:"total_exams"`
	TotalPractices   int        `json:"total_practices"`
	TotalErrors      int        `json:"total_errors"`
	TotalFavorites   int        `json:"total_favorites"`
	AvgExamScore     float64    `json:"avg_exam_score"`
	AvgPracticeRate  float64    `json:"avg_practice_rate"`
	ExamPassRate     float64    `json:"exam_pass_rate"`
	LastExamDate     *time.Time `json:"last_exam_date"`
	LastPracticeDate *time.Time `json:"last_practice_date"`
}

// NewStatisticsService 创建 StatisticsService 实例
func NewStatisticsService(db *sql.DB) *StatisticsService {
	return &StatisticsService{
		examRecordDAO: dao.NewExamRecordDAO(db),
		examDetailDAO: dao.NewExamQuestionDetailDAO(db),
		practiceDAO:   dao.NewPracticeProgressDAO(db),
		errorDAO:      dao.NewErrorQuestionDAO(db),
		favoriteDAO:   dao.NewFavoriteQuestionDAO(db),
		examConfig:    config.EXAM_CONFIG,
	}
}

// GetExamData 获取用户的考试数据
// Python 原版：exam_statistics_dao.get_exam_data(user_id, category, time_range)
func (s *StatisticsService) GetExamData(userID int64, category string, timeRange string) ([]*models.ExamStatisticsData, error) {
	utils.Debug("StatisticsService", "获取考试数据", map[string]interface{}{
		"user_id":    userID,
		"category":   category,
		"time_range": timeRange,
	})

	// 计算时间范围
	now := time.Now()
	var startDate time.Time
	switch timeRange {
	case "7days":
		startDate = now.AddDate(0, 0, -7)
	case "30days":
		startDate = now.AddDate(0, 0, -30)
	case "180days":
		startDate = now.AddDate(0, 0, -180)
	default:
		startDate = time.Time{} // 所有时间
	}

	// 查询考试数据
	examData, err := s.examRecordDAO.GetExamStatistics(userID, category, startDate)
	if err != nil {
		utils.Error("StatisticsService", "获取考试数据失败", err, nil)
		return nil, err
	}

	utils.Debug("StatisticsService", "获取考试数据成功", map[string]interface{}{
		"count": len(examData),
	})

	return examData, nil
}

// CalculateExamStatistics 计算考试统计数据
// Python 原版：exam_statistics_service.calculate_statistics(exam_data)
func (s *StatisticsService) CalculateExamStatistics(examData []*models.ExamStatisticsData) (*ExamStatisticsResult, error) {
	utils.Debug("StatisticsService", "计算考试统计数据", map[string]interface{}{
		"data_count": len(examData),
	})

	if len(examData) == 0 {
		utils.Debug("StatisticsService", "没有考试数据，返回默认统计", nil)
		return &ExamStatisticsResult{
			MaxScore:       0,
			LatestScore:    0,
			LatestDuration: 0,
			AvgPassRate:    0,
			TotalExams:     0,
		}, nil
	}

	// 计算最高分
	var maxScore float64
	for _, exam := range examData {
		if exam.Score > maxScore {
			maxScore = exam.Score
		}
	}

	// 最新一次成绩
	latestExam := examData[len(examData)-1]
	latestScore := latestExam.Score
	latestDuration := latestExam.DurationSeconds

	// 计算通过率
	passedExams := 0
	for _, exam := range examData {
		passScore := s.getPassScore(exam.Category)
		if exam.Score >= float64(passScore) {
			passedExams++
		}
	}

	passRate := float64(passedExams) / float64(len(examData)) * 100

	utils.Debug("StatisticsService", "计算完成", map[string]interface{}{
		"max_score":    maxScore,
		"latest_score": latestScore,
		"pass_rate":    passRate,
		"total_exams":  len(examData),
	})

	return &ExamStatisticsResult{
		MaxScore:       maxScore,
		LatestScore:    latestScore,
		LatestDuration: latestDuration,
		AvgPassRate:    passRate,
		TotalExams:     len(examData),
	}, nil
}

// getPassScore 获取指定类别的通过分数
func (s *StatisticsService) getPassScore(category string) int {
	if config, exists := s.examConfig[category]; exists {
		return config.PassScore
	}
	return 60 // 默认通过分数
}

// GetUserStatistics 获取用户综合统计
func (s *StatisticsService) GetUserStatistics(userID int64) (*UserStatisticsResult, error) {
	utils.Info("StatisticsService", "获取用户统计", map[string]interface{}{
		"user_id": userID,
	})

	result := &UserStatisticsResult{}

	// 获取考试统计
	examData, err := s.GetExamData(userID, "", "all")
	if err != nil {
		utils.Error("StatisticsService", "获取考试数据失败", err, nil)
		return nil, err
	}
	result.TotalExams = len(examData)

	// 计算平均分数和通过率
	if len(examData) > 0 {
		var totalScore float64
		passedExams := 0
		for _, exam := range examData {
			totalScore += exam.Score
			passScore := s.getPassScore(exam.Category)
			if exam.Score >= float64(passScore) {
				passedExams++
			}
		}
		result.AvgExamScore = totalScore / float64(len(examData))
		result.ExamPassRate = float64(passedExams) / float64(len(examData)) * 100
		result.LastExamDate = &examData[len(examData)-1].ExamDate
	}

	// 获取错题数量
	errorCount, err := s.errorDAO.GetCountByUser(userID)
	if err != nil {
		utils.Error("StatisticsService", "获取错题数量失败", err, nil)
		return nil, err
	}
	result.TotalErrors = errorCount

	// 获取收藏数量
	favoriteCount, err := s.favoriteDAO.GetCountByUser(userID)
	if err != nil {
		utils.Error("StatisticsService", "获取收藏数量失败", err, nil)
		return nil, err
	}
	result.TotalFavorites = favoriteCount

	// 获取练习进度统计
	practiceCount, err := s.practiceDAO.GetCountByUser(userID)
	if err != nil {
		utils.Error("StatisticsService", "获取练习进度失败", err, nil)
		return nil, err
	}
	result.TotalPractices = practiceCount

	utils.Info("StatisticsService", "获取用户统计成功", map[string]interface{}{
		"user_id":         userID,
		"total_exams":     result.TotalExams,
		"total_practices": result.TotalPractices,
		"total_errors":    result.TotalErrors,
		"total_favorites": result.TotalFavorites,
		"avg_exam_score":  result.AvgExamScore,
		"exam_pass_rate":  result.ExamPassRate,
	})

	return result, nil
}

// GetCategoryStatistics 获取分类统计
func (s *StatisticsService) GetCategoryStatistics(userID int64, category string) (map[string]interface{}, error) {
	utils.Debug("StatisticsService", "获取分类统计", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	stats := make(map[string]interface{})

	// 获取该类别的考试数据
	examData, err := s.GetExamData(userID, category, "all")
	if err != nil {
		return nil, err
	}

	// 计算该类别的统计
	if len(examData) > 0 {
		var totalScore float64
		passedExams := 0
		for _, exam := range examData {
			totalScore += exam.Score
			passScore := s.getPassScore(exam.Category)
			if exam.Score >= float64(passScore) {
				passedExams++
			}
		}
		stats["total_exams"] = len(examData)
		stats["avg_score"] = totalScore / float64(len(examData))
		stats["pass_rate"] = float64(passedExams) / float64(len(examData)) * 100
	} else {
		stats["total_exams"] = 0
		stats["avg_score"] = 0.0
		stats["pass_rate"] = 0.0
	}

	// 获取该类别的错题数量
	errorCount, err := s.errorDAO.GetCountByUserAndCategory(userID, category)
	if err != nil {
		return nil, err
	}
	stats["total_errors"] = errorCount

	// 获取该类别的收藏数量
	favoriteCount, err := s.favoriteDAO.GetCountByUserAndCategory(userID, category)
	if err != nil {
		return nil, err
	}
	stats["total_favorites"] = favoriteCount

	utils.Debug("StatisticsService", "获取分类统计成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"stats":    stats,
	})

	return stats, nil
}
