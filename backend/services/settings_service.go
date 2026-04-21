package services

import (
	"crac_exam_go/backend/config"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/utils"
	"fmt"

	"gorm.io/gorm"
)

// SettingsService 设置服务
type SettingsService struct {
	userDAO             *dao.UserDAO
	questionDAO         *dao.QuestionDAO
	examRecordDAO       *dao.ExamRecordDAO
	examDetailDAO       *dao.ExamQuestionDetailDAO
	practiceProgressDAO *dao.PracticeProgressDAO
	errorQuestionDAO    *dao.ErrorQuestionDAO
	favoriteDAO         *dao.FavoriteQuestionDAO
	importService       *ImportService
	questionsBankDAO    *dao.QuestionsBankDAO
	examConfig          map[string]config.ExamConfig
}

// AppInfo 应用信息
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewSettingsService 创建 SettingsService 实例
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		userDAO:             dao.NewUserDAO(db),
		questionDAO:         dao.NewQuestionDAO(db),
		examRecordDAO:       dao.NewExamRecordDAO(db),
		examDetailDAO:       dao.NewExamQuestionDetailDAO(db),
		practiceProgressDAO: dao.NewPracticeProgressDAO(db),
		errorQuestionDAO:    dao.NewErrorQuestionDAO(db),
		favoriteDAO:         dao.NewFavoriteQuestionDAO(db),
		importService:       NewImportService(db),
		questionsBankDAO:    dao.NewQuestionsBankDAO(db),
		examConfig:          config.EXAM_CONFIG,
	}
}

// ImportQuestions 导入题目
func (s *SettingsService) ImportQuestions(filePath string) (*ImportResult, error) {
	return s.importService.ProcessUnifiedData(filePath)
}

// ClearUserData 清空用户数据
func (s *SettingsService) ClearUserData(userID int64) error {
	utils.Info("SettingsService", "清空用户数据", map[string]interface{}{
		"user_id": userID,
	})

	// 清空考试记录
	if err := s.examRecordDAO.ClearByUser(userID); err != nil {
		return err
	}
	// 清空错题
	if err := s.errorQuestionDAO.ClearByUser(userID); err != nil {
		return err
	}
	// 清空收藏
	if err := s.favoriteDAO.ClearByUser(userID); err != nil {
		return err
	}
	// 清空练习进度
	if err := s.practiceProgressDAO.ClearByUser(userID); err != nil {
		return err
	}

	return nil
}

// ClearQuestionBank 清空题库
func (s *SettingsService) ClearQuestionBank() error {
	utils.Info("SettingsService", "清空题库", nil)
	return s.questionDAO.ClearAll()
}

// GetQuestionsPage 获取题库分页数据
func (s *SettingsService) GetQuestionsPage(pageNum, pageSize int, searchQuery string, filterLA, filterLB, filterLC bool) (map[string]interface{}, error) {
	total, err := s.questionsBankDAO.GetTotalRecords(searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		return nil, err
	}

	result, err := s.questionsBankDAO.GetPageData(pageNum, pageSize, searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data":  result.Data,
		"total": total,
	}, nil
}

// UpdateQuestion 更新题目
func (s *SettingsService) UpdateQuestion(questionID int64, updatedData map[string]interface{}) error {
	return s.questionsBankDAO.UpdateQuestion(questionID, updatedData)
}

// DeleteQuestion 删除题目
func (s *SettingsService) DeleteQuestion(questionID int64) error {
	return s.questionsBankDAO.DeleteQuestion(questionID)
}

// GetQuestionByID 根据 ID 获取题目
func (s *SettingsService) GetQuestionByID(questionID int64) (interface{}, error) {
	return s.questionsBankDAO.GetQuestionByID(questionID)
}

// GetAppInfo 获取应用信息
func (s *SettingsService) GetAppInfo() *AppInfo {
	return &AppInfo{
		Name:    config.AppConfig.AppName,
		Version: config.AppConfig.Version,
	}
}

// GetExamConfig 获取考试配置
func (s *SettingsService) GetExamConfig(category string) interface{} {
	if cfg, exists := s.examConfig[category]; exists {
		return cfg
	}
	return s.examConfig["A"]
}

// GetAllExamConfigs 获取所有考试配置
func (s *SettingsService) GetAllExamConfigs() map[string]config.ExamConfig {
	return s.examConfig
}

// DeleteUser 删除用户
func (s *SettingsService) DeleteUser(userID int64) (bool, error) {
	utils.Info("SettingsService", "开始删除用户", map[string]interface{}{
		"user_id": userID,
	})

	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("用户不存在")
	}

	return true, nil
}
