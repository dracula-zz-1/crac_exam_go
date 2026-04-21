package services

import (
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// QuestionsBankService 题库服务
type QuestionsBankService struct {
	dao *dao.QuestionsBankDAO
}

// NewQuestionsBankService 创建 QuestionsBankService 实例
func NewQuestionsBankService(db *gorm.DB) *QuestionsBankService {
	return &QuestionsBankService{
		dao: dao.NewQuestionsBankDAO(db),
	}
}

// PageDataResult 分页数据结果
type PageDataResult struct {
	Data       []*models.Question `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// GetPageData 获取指定页的数据，并进行必要的格式转换
// Python 原版：get_page_data(page_num, page_size, search_query, filter_la, filter_lb, filter_lc) -> List[Dict]
func (s *QuestionsBankService) GetPageData(pageNum, pageSize int, searchQuery string, filterLA, filterLB, filterLC bool) (*PageDataResult, error) {
	utils.Info("QuestionsBankService", "获取题库分页数据", map[string]interface{}{
		"page":      pageNum,
		"page_size": pageSize,
		"search":    searchQuery,
		"filter_la": filterLA,
		"filter_lb": filterLB,
		"filter_lc": filterLC,
	})

	// 获取分页数据
	result, err := s.dao.GetPageData(pageNum, pageSize, searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		utils.Error("QuestionsBankService", "获取分页数据失败", err, nil)
		return nil, err
	}

	// 进行数据格式转换
	for _, item := range result.Data {
		// 转换题型为文字描述
		item.TypeText = s.getTypeText(item.Type)
		// 处理图片显示
		item.HasImage = s.hasImage(item.F)
	}

	// 计算总页数
	totalPages := int((result.Total + int64(pageSize) - 1) / int64(pageSize))

	pageResult := &PageDataResult{
		Data:       result.Data,
		Total:      int(result.Total),
		Page:       pageNum,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	utils.Debug("QuestionsBankService", "获取题库分页数据成功", map[string]interface{}{
		"total":       result.Total,
		"page":        pageNum,
		"total_pages": totalPages,
	})

	return pageResult, nil
}

// UpdateQuestion 更新题目数据
// Python 原版：update_question(question_id, updated_data) -> bool
func (s *QuestionsBankService) UpdateQuestion(questionID int64, updatedData map[string]interface{}) error {
	utils.Info("QuestionsBankService", "更新题目", map[string]interface{}{
		"question_id":  questionID,
		"updated_data": updatedData,
	})

	// 数据验证
	if questionID == 0 {
		utils.Error("QuestionsBankService", "更新题目失败", nil, map[string]interface{}{
			"error": "题目 ID 不能为空",
		})
		return nil
	}

	if len(updatedData) == 0 {
		utils.Error("QuestionsBankService", "更新题目失败", nil, map[string]interface{}{
			"error": "更新数据不能为空",
		})
		return nil
	}

	// 调用 DAO 层更新数据
	err := s.dao.UpdateQuestion(questionID, updatedData)
	if err != nil {
		utils.Error("QuestionsBankService", "更新题目失败", err, map[string]interface{}{
			"question_id": questionID,
		})
		return err
	}

	utils.Info("QuestionsBankService", "更新题目成功", map[string]interface{}{
		"question_id": questionID,
	})

	return nil
}

// GetQuestionByID 根据 ID 获取题目
// Python 原版：get_question_by_id(question_id) -> Optional[Dict]
func (s *QuestionsBankService) GetQuestionByID(questionID int64) (*models.Question, error) {
	utils.Debug("QuestionsBankService", "根据 ID 获取题目", map[string]interface{}{
		"question_id": questionID,
	})

	question, err := s.dao.GetQuestionByID(questionID)
	if err != nil {
		utils.Error("QuestionsBankService", "根据 ID 获取题目失败", err, map[string]interface{}{
			"question_id": questionID,
		})
		return nil, err
	}

	if question != nil {
		// 转换题型为文字描述
		question.TypeText = s.getTypeText(question.Type)
	}

	return question, nil
}

// GetFilteredRecordsCount 获取筛选后的记录总数
// Python 原版：get_filtered_records_count(search_query, filter_la, filter_lb, filter_lc) -> int
func (s *QuestionsBankService) GetFilteredRecordsCount(searchQuery string, filterLA, filterLB, filterLC bool) (int, error) {
	count, err := s.dao.GetTotalRecords(searchQuery, filterLA, filterLB, filterLC)
	if err != nil {
		utils.Error("QuestionsBankService", "获取筛选记录数失败", err, nil)
		return 0, err
	}

	return int(count), nil
}

// DeleteQuestion 删除题目
func (s *QuestionsBankService) DeleteQuestion(questionID int64) error {
	utils.Info("QuestionsBankService", "删除题目", map[string]interface{}{
		"question_id": questionID,
	})

	err := s.dao.DeleteQuestion(questionID)
	if err != nil {
		utils.Error("QuestionsBankService", "删除题目失败", err, map[string]interface{}{
			"question_id": questionID,
		})
		return err
	}

	return nil
}

// getTypeText 转换题型为文字描述
func (s *QuestionsBankService) getTypeText(typeVal int) string {
	switch typeVal {
	case 1:
		return "单选"
	case 2:
		return "多选"
	default:
		return "未知"
	}
}

// hasImage 判断是否有图片
func (s *QuestionsBankService) hasImage(imageData string) string {
	if imageData != "" && (len(imageData) > 10 && (imageData[:10] == "data:image" || imageData[:4] == "img/")) {
		return "有"
	}
	return "无"
}

// ExportToJSON 导出题目到 JSON 文件
// Python 原版：export_questions_to_json(db_path, output_path) -> bool
func (s *QuestionsBankService) ExportToJSON(outputPath string) (bool, error) {
	utils.Info("QuestionsBankService", "导出题目到 JSON", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.dao.GetAllQuestions()
	if err != nil {
		utils.Error("QuestionsBankService", "获取所有题目失败", err, nil)
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
			"type": strconv.Itoa(q.Type),
		}
		// 去除 I 字段中的空格
		if question["I"] != nil {
			question["I"] = q.I
		}
		// 如果 F 字段为空，设置为 null
		if question["F"] == "" {
			question["F"] = nil
		}
		questions[i] = question
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("QuestionsBankService", "创建目录失败", err, map[string]interface{}{
			"dir":         dir,
			"output_path": outputPath,
		})
		return false, err
	}

	// 写入 JSON 文件
	file, err := os.Create(outputPath)
	if err != nil {
		utils.Error("QuestionsBankService", "创建文件失败", err, map[string]interface{}{
			"output_path": outputPath,
		})
		return false, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(questions); err != nil {
		utils.Error("QuestionsBankService", "写入 JSON 文件失败", err, nil)
		return false, err
	}

	utils.Info("QuestionsBankService", "导出题目到 JSON 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(questions),
	})

	return true, nil
}

// ExportToCSV 导出题目到 CSV 文件
// Python 原版：json_to_csv(json_path, csv_path) -> bool
func (s *QuestionsBankService) ExportToCSV(outputPath string) (bool, error) {
	utils.Info("QuestionsBankService", "导出题目到 CSV", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.dao.GetAllQuestions()
	if err != nil {
		utils.Error("QuestionsBankService", "获取所有题目失败", err, nil)
		return false, err
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("QuestionsBankService", "创建目录失败", err, map[string]interface{}{
			"dir": dir,
		})
		return false, err
	}

	// 创建 CSV 文件
	file, err := os.Create(outputPath)
	if err != nil {
		utils.Error("QuestionsBankService", "创建文件失败", err, map[string]interface{}{
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
		utils.Error("QuestionsBankService", "写入 CSV 表头失败", err, nil)
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
			strconv.Itoa(int(q.Type)),
		}
		if err := writer.Write(record); err != nil {
			utils.Error("QuestionsBankService", "写入 CSV 记录失败", err, nil)
			return false, err
		}
	}

	utils.Info("QuestionsBankService", "导出题目到 CSV 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(allQuestions),
	})

	return true, nil
}

// ExportToExcel 导出题目到 Excel 文件
// Python 原版：convert_json_to_excel(json_path, excel_path) -> bool
func (s *QuestionsBankService) ExportToExcel(outputPath string) (bool, error) {
	utils.Info("QuestionsBankService", "导出题目到 Excel", map[string]interface{}{
		"output_path": outputPath,
	})

	// 获取所有题目
	allQuestions, err := s.dao.GetAllQuestions()
	if err != nil {
		utils.Error("QuestionsBankService", "获取所有题目失败", err, nil)
		return false, err
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("QuestionsBankService", "创建目录失败", err, map[string]interface{}{
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
		utils.Error("QuestionsBankService", "保存 Excel 文件失败", err, nil)
		return false, err
	}

	utils.Info("QuestionsBankService", "导出题目到 Excel 成功", map[string]interface{}{
		"output_path": outputPath,
		"count":       len(allQuestions),
	})

	return true, nil
}
