package services

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// CSVImporter CSV 导入器
type CSVImporter struct{}

// NewCSVImporter 创建 CSV 导入器实例
func NewCSVImporter() *CSVImporter {
	return &CSVImporter{}
}

// ProcessCSVData 处理 CSV 文件导入
// Python 原版：import_csv.process_csv_data(file_path, ...)
func (i *CSVImporter) ProcessCSVData(filePath string) ([]*models.Question, error) {
	utils.Info("CSVImporter", "开始处理 CSV 文件", map[string]interface{}{
		"file_path": filePath,
	})

	// 打开 CSV 文件
	file, err := os.Open(filePath)
	if err != nil {
		utils.Error("CSVImporter", "打开 CSV 文件失败", err, nil)
		return nil, err
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)
	// 支持 UTF-8 BOM
	// reader.LazyQuotes = true

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		utils.Error("CSVImporter", "读取 CSV 数据失败", err, nil)
		return nil, err
	}

	utils.Debug("CSVImporter", "CSV 记录数量", map[string]interface{}{
		"count": len(records),
	})

	// 解析题目数据
	questions := make([]*models.Question, 0)

	// 假设第一行是标题行，从第二行开始是数据
	for idx, record := range records {
		// 跳过标题行
		if idx == 0 {
			continue
		}

		question, err := i.parseCSVRecord(record, idx)
		if err != nil {
			utils.Warn("CSVImporter", "解析 CSV 记录失败", map[string]interface{}{
				"row":   idx + 1,
				"error": err.Error(),
			})
			continue
		}

		if question != nil {
			questions = append(questions, question)
		}
	}

	utils.Info("CSVImporter", "CSV 文件处理完成", map[string]interface{}{
		"total_records":   len(records) - 1, // 减去标题行
		"valid_questions": len(questions),
	})

	return questions, nil
}

// ProcessCSVDataFromContent 从字符串内容处理 CSV 数据（支持前端直接传递内容）
func (i *CSVImporter) ProcessCSVDataFromContent(content string) ([]*models.Question, error) {
	utils.Info("CSVImporter", "开始处理 CSV 内容", nil)

	// 创建 CSV 读取器
	reader := csv.NewReader(strings.NewReader(content))

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		utils.Error("CSVImporter", "读取 CSV 数据失败", err, nil)
		return nil, err
	}

	utils.Debug("CSVImporter", "CSV 记录数量", map[string]interface{}{
		"count": len(records),
	})

	// 解析题目数据（CSV 没有标题行，全部作为数据处理）
	questions := make([]*models.Question, 0)

	// 从第 1 行开始解析所有记录
	for idx := 0; idx < len(records); idx++ {
		record := records[idx]

		question, err := i.parseCSVRecord(record, idx+1)
		if err != nil {
			utils.Warn("CSVImporter", "解析 CSV 记录失败", map[string]interface{}{
				"row":   idx + 1,
				"error": err.Error(),
			})
			continue
		}

		if question != nil {
			questions = append(questions, question)
		}
	}

	utils.Info("CSVImporter", "CSV 内容处理完成", map[string]interface{}{
		"total_records":   len(records),
		"valid_questions": len(questions),
	})

	return questions, nil
}

// parseCSVRecord 解析单行 CSV 记录
func (i *CSVImporter) parseCSVRecord(record []string, rowNum int) (*models.Question, error) {
	// 检查字段数量（至少需要 9 个字段：J,P,I,Q,T,A,B,C,D）
	if len(record) < 9 {
		return nil, nil // 跳过不完整的记录
	}

	// 解析字段
	question := &models.Question{}

	// 基本字段
	question.J = record[0] // 编号 1
	question.P = record[1] // 编号 2
	question.I = record[2] // 编号 3
	question.Q = record[3] // 题干
	question.T = record[4] // 答案
	question.A = record[5] // 选项 A
	question.B = record[6] // 选项 B
	question.C = record[7] // 选项 C
	question.D = record[8] // 选项 D

	// 可选字段
	if len(record) > 9 {
		question.F = record[9] // 图片（Base64）
	}

	// 解析分类字段
	if len(record) > 10 {
		laVal, err := strconv.Atoi(record[10])
		if err == nil {
			question.LA = laVal
		} else {
			utils.Debug("CSVImporter", "LA 解析失败", logrus.Fields{
				"value": record[10],
				"error": err,
			})
		}
	}
	if len(record) > 11 {
		lbVal, err := strconv.Atoi(record[11])
		if err == nil {
			question.LB = lbVal
		} else {
			utils.Debug("CSVImporter", "LB 解析失败", logrus.Fields{
				"value": record[11],
				"error": err,
			})
		}
	}
	if len(record) > 12 {
		lcVal, err := strconv.Atoi(record[12])
		if err == nil {
			question.LC = lcVal
		} else {
			utils.Debug("CSVImporter", "LC 解析失败", logrus.Fields{
				"value": record[12],
				"error": err,
			})
		}
	}

	// 解析题型
	if len(record) > 13 {
		typeVal, err := strconv.Atoi(record[13])
		if err != nil {
			// 根据答案长度自动判断题型
			if len(question.T) == 1 {
				question.Type = 1 // 单选题
			} else if len(question.T) > 1 {
				question.Type = 2 // 多选题
			} else {
				question.Type = 0
			}
		} else {
			question.Type = typeVal
		}
	} else {
		// 根据答案长度自动判断题型
		if len(question.T) == 1 {
			question.Type = 1 // 单选题
		} else if len(question.T) > 1 {
			question.Type = 2 // 多选题
		} else {
			question.Type = 0
		}
	}

	// 验证必要字段
	if question.J == "" || question.Q == "" || question.T == "" {
		return nil, nil // 跳过缺少必要字段的记录
	}

	return question, nil
}
