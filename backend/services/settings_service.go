package services

import (
	"crac_exam_go/backend/config"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// SettingsService 设置服务
type SettingsService struct {
	userDAO             *dao.UserDAO
	errorQuestionDAO    *dao.ErrorQuestionDAO
	practiceProgressDAO *dao.PracticeProgressDAO
	examRecordDAO       *dao.ExamRecordDAO
	examQuestionDetail  *dao.ExamQuestionDetailDAO
	favoriteQuestionDAO *dao.FavoriteQuestionDAO
	questionsBankDAO    *dao.QuestionsBankDAO
	questionDAO         *dao.QuestionDAO
	importService       *ImportService
}

// AppInfo 应用信息
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Copyright   string `json:"copyright"`
}

// NewSettingsService 创建 SettingsService 实例
func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{
		userDAO:             dao.NewUserDAO(db),
		errorQuestionDAO:    dao.NewErrorQuestionDAO(db),
		practiceProgressDAO: dao.NewPracticeProgressDAO(db),
		examRecordDAO:       dao.NewExamRecordDAO(db),
		examQuestionDetail:  dao.NewExamQuestionDetailDAO(db),
		favoriteQuestionDAO: dao.NewFavoriteQuestionDAO(db),
		questionsBankDAO:    dao.NewQuestionsBankDAO(db),
		questionDAO:         dao.NewQuestionDAO(db),
		importService:       NewImportService(db, "python", "./python_scripts"),
	}
}

// ClearUserData 清空用户数据
// Python 原版：db_init.clear_user_data(user_id)
func (s *SettingsService) ClearUserData(userID int64) error {
	utils.Info("SettingsService", "开始清空用户数据", map[string]interface{}{
		"user_id": userID,
	})

	// 清空用户的错题本
	err := s.errorQuestionDAO.ClearByUser(userID)
	if err != nil {
		utils.Error("SettingsService", "清空错题本失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 清空用户的收藏夹
	err = s.favoriteQuestionDAO.ClearByUser(userID)
	if err != nil {
		utils.Error("SettingsService", "清空收藏夹失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 清空用户的练习进度
	err = s.practiceProgressDAO.ClearByUser(userID)
	if err != nil {
		utils.Error("SettingsService", "清空练习进度失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	// 清空用户的考试记录
	err = s.examRecordDAO.ClearByUser(userID)
	if err != nil {
		utils.Error("SettingsService", "清空考试记录失败", err, map[string]interface{}{
			"user_id": userID,
		})
		return err
	}

	utils.Info("SettingsService", "用户数据清空完成", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// ClearQuestionBank 清空题库数据
func (s *SettingsService) ClearQuestionBank() error {
	utils.Info("SettingsService", "开始清空题库数据", nil)

	// 清空题库表
	err := s.questionDAO.ClearAll()
	if err != nil {
		utils.Error("SettingsService", "清空题库失败", err, nil)
		return err
	}

	utils.Info("SettingsService", "题库数据清空完成", nil)

	return nil
}

// GetAllUsers 获取所有用户
func (s *SettingsService) GetAllUsers() ([]map[string]interface{}, error) {
	utils.Debug("SettingsService", "获取所有用户", nil)

	users, err := s.userDAO.GetAll()
	if err != nil {
		utils.Error("SettingsService", "获取所有用户失败", err, nil)
		return nil, err
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"id_card":    user.IDCard,
			"last_login": user.LastLogin,
		}
	}

	return result, nil
}

// DeleteUser 删除用户
func (s *SettingsService) DeleteUser(userID int64) error {
	utils.Info("SettingsService", "删除用户", map[string]interface{}{
		"user_id": userID,
	})

	// 先清空用户数据
	err := s.ClearUserData(userID)
	if err != nil {
		return err
	}

	// Python 原版没有删除用户表中的用户记录，只清空用户数据
	// 这里也保持同样的行为

	utils.Info("SettingsService", "用户删除成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// GetUserCount 获取用户数量
func (s *SettingsService) GetUserCount() (int, error) {
	utils.Debug("SettingsService", "获取用户数量", nil)

	users, err := s.userDAO.GetAll()
	if err != nil {
		return 0, err
	}

	return len(users), nil
}

// ResetDatabase 重置数据库（清空题库）
func (s *SettingsService) ResetDatabase() error {
	utils.Info("SettingsService", "重置数据库", nil)

	// 清空题库
	err := s.questionsBankDAO.ResetTable()
	if err != nil {
		utils.Error("SettingsService", "重置题库失败", err, nil)
		return err
	}

	utils.Info("SettingsService", "数据库重置完成", nil)

	return nil
}

// GetDatabaseStats 获取数据库统计信息
func (s *SettingsService) GetDatabaseStats() (map[string]int, error) {
	utils.Debug("SettingsService", "获取数据库统计信息", nil)

	// 获取题目数量
	questions, err := s.questionsBankDAO.GetAllQuestions()
	if err != nil {
		return nil, err
	}

	// 获取用户数量
	users, err := s.userDAO.GetAll()
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total_questions": len(questions),
		"total_users":     len(users),
	}

	return stats, nil
}

// ImportQuestions 导入题目
// Python 原版：import_*.py 系列脚本
func (s *SettingsService) ImportQuestions(filePath string) (*ImportResult, error) {
	utils.Info("SettingsService", "开始导入题目", map[string]interface{}{
		"file_path": filePath,
	})

	return s.importService.ProcessUnifiedData(filePath)
}

// GetQuestionsPage 获取题目分页数据
func (s *SettingsService) GetQuestionsPage(pageNum, pageSize int, searchQuery string, filterLA, filterLB, filterLC bool) (map[string]interface{}, error) {
	utils.Info("SettingsService", "获取题目分页数据", map[string]interface{}{
		"page_num":     pageNum,
		"page_size":    pageSize,
		"search_query": searchQuery,
		"filter_LA":    filterLA,
		"filter_LB":    filterLB,
		"filter_LC":    filterLC,
	})

	result, err := s.questionsBankDAO.GetPageData(pageNum, pageSize, searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		utils.Error("SettingsService", "获取题目分页数据失败", err, nil)
		return nil, err
	}

	utils.Info("SettingsService", "获取题目分页数据成功", map[string]interface{}{
		"total":      result.Total,
		"data_count": len(result.Data),
	})

	// 返回包含数据和总数的对象
	return map[string]interface{}{
		"data":  result.Data,
		"total": result.Total,
	}, nil
}

// UpdateQuestion 更新题目
func (s *SettingsService) UpdateQuestion(questionID int64, updatedData map[string]interface{}) error {
	utils.Info("SettingsService", "更新题目", map[string]interface{}{
		"question_id": questionID,
	})

	question, err := s.questionsBankDAO.GetQuestionByID(questionID)
	if err != nil || question == nil {
		return fmt.Errorf("题目不存在")
	}

	question.FromMap(updatedData)
	err = s.questionsBankDAO.UpdateQuestion(questionID, updatedData)
	if err != nil {
		utils.Error("SettingsService", "更新题目失败", err, map[string]interface{}{
			"question_id": questionID,
		})
		return err
	}

	utils.Info("SettingsService", "题目更新成功", map[string]interface{}{
		"question_id": questionID,
	})

	return nil
}

// DeleteQuestion 删除题目
func (s *SettingsService) DeleteQuestion(questionID int64) error {
	utils.Info("SettingsService", "删除题目", map[string]interface{}{
		"question_id": questionID,
	})

	err := s.questionsBankDAO.DeleteQuestion(questionID)
	if err != nil {
		utils.Error("SettingsService", "删除题目失败", err, map[string]interface{}{
			"question_id": questionID,
		})
		return err
	}

	utils.Info("SettingsService", "题目删除成功", map[string]interface{}{
		"question_id": questionID,
	})

	return nil
}

// GetQuestionByID 根据 ID 获取题目
func (s *SettingsService) GetQuestionByID(questionID int64) (*models.Question, error) {
	utils.Debug("SettingsService", "获取题目详情", map[string]interface{}{
		"question_id": questionID,
	})

	question, err := s.questionsBankDAO.GetQuestionByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("题目不存在")
	}

	if question == nil {
		return nil, fmt.Errorf("题目不存在")
	}

	return question, nil
}

// GetFilteredRecordsCount 获取过滤后的记录数量
func (s *SettingsService) GetFilteredRecordsCount(searchQuery string, filterLA, filterLB, filterLC bool) (int, error) {
	utils.Debug("SettingsService", "获取过滤后的记录数量", map[string]interface{}{
		"search_query": searchQuery,
		"filter_LA":    filterLA,
		"filter_LB":    filterLB,
		"filter_LC":    filterLC,
	})

	// 获取所有符合条件的题目
	result, err := s.questionsBankDAO.GetPageData(1, 10000, searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		return 0, err
	}

	return result.Total, nil
}

// GetAppInfo 获取应用信息
func (s *SettingsService) GetAppInfo() *AppInfo {
	return &AppInfo{
		Name:        "业余无线电模拟考试系统",
		Version:     "1.0.0",
		Description: "业余无线电操作技术能力验证模拟考试系统",
		Author:      "Crac Exam Team",
		Copyright:   "Copyright © 2025",
	}
}

// GetExamConfig 获取考试配置
func (s *SettingsService) GetExamConfig(category string) config.ExamConfig {
	utils.Debug("SettingsService", "获取考试配置", map[string]interface{}{
		"category": category,
	})

	return config.EXAM_CONFIG[category]
}

// GetAllExamConfigs 获取所有考试配置
func (s *SettingsService) GetAllExamConfigs() map[string]config.ExamConfig {
	utils.Debug("SettingsService", "获取所有考试配置", nil)

	return config.EXAM_CONFIG
}

// ExportQuestionsToJSON 导出题目到 JSON 文件
// Python 原版：export_questions_to_json(db_path, output_path) -> bool
func (s *SettingsService) ExportQuestionsToJSON(outputPath string) (bool, error) {
	utils.Info("SettingsService", "导出题目到 JSON", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.questionsBankDAO.GetAllQuestions()
	if err != nil {
		utils.Error("SettingsService", "获取所有题目失败", err, nil)
		return false, err
	}

	// 格式化数据
	questions := make([]map[string]interface{}, len(allQuestions))
	for i, q := range allQuestions {
		question := map[string]interface{}{
			"id":   q.ID,
			"J":    q.J,
			"P":    q.P,
			"I":    q.I,
			"Q":    q.Q,
			"T":    q.T,
			"A":    q.A,
			"B":    q.B,
			"C":    q.C,
			"D":    q.D,
			"F":    q.F,
			"LA":   q.LA,
			"LB":   q.LB,
			"LC":   q.LC,
			"type": fmt.Sprintf("%d", q.Type),
		}
		questions[i] = question
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("SettingsService", "创建目录失败", err, map[string]interface{}{
			"dir": dir,
		})
		return false, err
	}

	// 写入 JSON 文件
	file, err := os.Create(outputPath)
	if err != nil {
		utils.Error("SettingsService", "创建文件失败", err, map[string]interface{}{
			"output_path": outputPath,
		})
		return false, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(questions); err != nil {
		utils.Error("SettingsService", "写入 JSON 文件失败", err, nil)
		return false, err
	}

	utils.Info("SettingsService", "导出题目到 JSON 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(questions),
	})

	return true, nil
}

// ExportQuestionsToCSV 导出题目到 CSV 文件
// Python 原版：json_to_csv(json_path, csv_path) -> bool
func (s *SettingsService) ExportQuestionsToCSV(outputPath string) (bool, error) {
	utils.Info("SettingsService", "导出题目到 CSV", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.questionsBankDAO.GetAllQuestions()
	if err != nil {
		utils.Error("SettingsService", "获取所有题目失败", err, nil)
		return false, err
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("SettingsService", "创建目录失败", err, map[string]interface{}{
			"dir": dir,
		})
		return false, err
	}

	// 创建 CSV 文件
	file, err := os.Create(outputPath)
	if err != nil {
		utils.Error("SettingsService", "创建文件失败", err, map[string]interface{}{
			"output_path": outputPath,
		})
		return false, err
	}
	defer file.Close()

	// 设置 UTF-8 BOM 以支持 Excel 正确读取中文
	file.WriteString("\xef\xbb\xbf")

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头（排除 id 字段）
	headers := []string{"J", "P", "I", "Q", "T", "A", "B", "C", "D", "F", "LA", "LB", "LC", "type"}
	if err := writer.Write(headers); err != nil {
		utils.Error("SettingsService", "写入 CSV 表头失败", err, nil)
		return false, err
	}

	// 写入数据
	for _, q := range allQuestions {
		record := []string{
			q.J,
			q.P,
			q.I,
			q.Q,
			q.T,
			q.A,
			q.B,
			q.C,
			q.D,
			q.F,
			strconv.Itoa(q.LA),
			strconv.Itoa(q.LB),
			strconv.Itoa(q.LC),
			fmt.Sprintf("%d", q.Type),
		}
		if err := writer.Write(record); err != nil {
			utils.Error("SettingsService", "写入 CSV 记录失败", err, nil)
			return false, err
		}
	}

	utils.Info("SettingsService", "导出题目到 CSV 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(allQuestions),
	})

	return true, nil
}

// ExportQuestionsToExcel 导出题目到 Excel 文件
// Python 原版：convert_json_to_excel(json_path, excel_path) -> bool
func (s *SettingsService) ExportQuestionsToExcel(outputPath string) (bool, error) {
	utils.Info("SettingsService", "导出题目到 Excel", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.questionsBankDAO.GetAllQuestions()
	if err != nil {
		utils.Error("SettingsService", "获取所有题目失败", err, nil)
		return false, err
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("SettingsService", "创建目录失败", err, map[string]interface{}{
			"dir": dir,
		})
		return false, err
	}

	// 创建 Excel 文件
	f := excelize.NewFile()
	defer f.Close()

	// 设置工作表名称
	sheetIndex, _ := f.NewSheet("题目")
	f.SetActiveSheet(sheetIndex)

	// 写入表头（排除 id 字段）
	headers := []string{"J", "P", "I", "Q", "T", "A", "B", "C", "D", "F", "LA", "LB", "LC", "type"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("题目", cell, header)
	}

	// 写入数据
	for i, q := range allQuestions {
		row := i + 2
		f.SetCellValue("题目", fmt.Sprintf("A%d", row), q.J)
		f.SetCellValue("题目", fmt.Sprintf("B%d", row), q.P)
		f.SetCellValue("题目", fmt.Sprintf("C%d", row), q.I)
		f.SetCellValue("题目", fmt.Sprintf("D%d", row), q.Q)
		f.SetCellValue("题目", fmt.Sprintf("E%d", row), q.T)
		f.SetCellValue("题目", fmt.Sprintf("F%d", row), q.A)
		f.SetCellValue("题目", fmt.Sprintf("G%d", row), q.B)
		f.SetCellValue("题目", fmt.Sprintf("H%d", row), q.C)
		f.SetCellValue("题目", fmt.Sprintf("I%d", row), q.D)
		f.SetCellValue("题目", fmt.Sprintf("J%d", row), q.F)
		f.SetCellValue("题目", fmt.Sprintf("K%d", row), q.LA)
		f.SetCellValue("题目", fmt.Sprintf("L%d", row), q.LB)
		f.SetCellValue("题目", fmt.Sprintf("M%d", row), q.LC)
		f.SetCellValue("题目", fmt.Sprintf("N%d", row), q.Type)
	}

	// 自动调整列宽
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
	for i, header := range headers {
		colName := columns[i]
		width := float64(len(header) + 4)
		if width < 10 {
			width = 10
		}
		f.SetColWidth("题目", colName, colName, width)
	}

	// 保存文件
	if err := f.SaveAs(outputPath); err != nil {
		utils.Error("SettingsService", "保存 Excel 文件失败", err, nil)
		return false, err
	}

	utils.Info("SettingsService", "导出题目到 Excel 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(allQuestions),
	})

	return true, nil
}
