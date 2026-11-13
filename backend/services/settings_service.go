package services

import (
	"database/sql"
	"fmt"
	"runtime"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

type SettingsService struct {
	db *sql.DB
}

func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) ServiceName() string { return "SettingsService" }

func (s *SettingsService) ClearUserData(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("请先登录")
	}
	if err := store.ClearUserData(s.db, userID); err != nil {
		utils.Error("SettingsService", "清空用户数据失败", err, "user_id", userID)
		return fmt.Errorf("清空用户数据失败：%w", err)
	}
	utils.Info("SettingsService", "清空用户数据", "user_id", userID)
	return nil
}

func (s *SettingsService) ClearQuestionBank() error {
	if err := store.ClearQuestionBank(s.db); err != nil {
		utils.Error("SettingsService", "清空题库失败", err)
		return fmt.Errorf("清空题库失败：%w", err)
	}
	utils.Info("SettingsService", "清空题库")
	return nil
}

func (s *SettingsService) GetDBStats() (*models.DBStats, error) {
	stats := &models.DBStats{}
	var err error
	if stats.Users, err = store.CountUsers(s.db); err != nil {
		return nil, err
	}
	if stats.Questions, err = store.CountQuestions(s.db); err != nil {
		return nil, err
	}
	if stats.ExamRecords, err = store.CountExamRecords(s.db); err != nil {
		return nil, err
	}
	if stats.ErrorRecords, err = store.CountErrorRecords(s.db); err != nil {
		return nil, err
	}
	if stats.Favorites, err = store.CountFavoriteRecords(s.db); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *SettingsService) GetAppInfo() *models.AppInfo {
	return &models.AppInfo{
		Name:        config.AppName,
		Version:     config.AppVersion,
		Author:      config.AppAuthor,
		Description: config.AppDesc,
		Copyright:   "Copyright © 2024-2026 " + config.AppAuthor,
		GoVersion:   runtime.Version(),
		DataDir:     config.DataDir(),
	}
}
