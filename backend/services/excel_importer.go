package services

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// ExcelImporter Excel 导入器
type ExcelImporter struct{}

// NewExcelImporter 创建 Excel 导入器实例
func NewExcelImporter() *ExcelImporter {
	return &ExcelImporter{}
}

// ProcessExcelData 处理 Excel 文件导入
// Python 原版：import_excel.process_excel_data(file_path, ...)
func (i *ExcelImporter) ProcessExcelData(filePath string) ([]*models.Question, error) {
	utils.Info("ExcelImporter", "开始处理 Excel 文件", map[string]interface{}{
		"file_path": filePath,
	})

	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		utils.Error("ExcelImporter", "打开 Excel 文件失败", err, nil)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			utils.Error("ExcelImporter", "关闭 Excel 文件失败", err, nil)
		}
	}()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		utils.Error("ExcelImporter", "获取工作表失败", nil, nil)
		return nil, nil
	}

	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		utils.Error("ExcelImporter", "读取 Excel 数据失败", err, nil)
		return nil, err
	}

	utils.Debug("ExcelImporter", "Excel 行数", map[string]interface{}{
		"count": len(rows),
	})

	// 解析题目数据
	questions := make([]*models.Question, 0)

	// 假设第一行是标题行，从第二行开始是数据
	for idx, row := range rows {
		// 跳过标题行
		if idx == 0 {
			continue
		}

		question, err := i.parseExcelRow(row, idx)
		if err != nil {
			utils.Warn("ExcelImporter", "解析 Excel 行失败", map[string]interface{}{
				"row":   idx + 1,
				"error": err.Error(),
			})
			continue
		}

		if question != nil {
			questions = append(questions, question)
		}
	}

	utils.Info("ExcelImporter", "Excel 文件处理完成", map[string]interface{}{
		"total_rows":      len(rows) - 1, // 减去标题行
		"valid_questions": len(questions),
	})

	return questions, nil
}

// parseExcelRow 解析单行 Excel 数据
func (i *ExcelImporter) parseExcelRow(row []string, rowNum int) (*models.Question, error) {
	// 检查字段数量（至少需要 9 个字段：J,P,I,Q,T,A,B,C,D）
	if len(row) < 9 {
		return nil, nil // 跳过不完整的记录
	}

	// 解析字段
	question := &models.Question{}

	// 基本字段
	question.J = row[0] // 编号 1
	question.P = row[1] // 编号 2
	question.I = row[2] // 编号 3
	question.Q = row[3] // 题干
	question.T = row[4] // 答案
	question.A = row[5] // 选项 A
	question.B = row[6] // 选项 B
	question.C = row[7] // 选项 C
	question.D = row[8] // 选项 D

	// 可选字段
	if len(row) > 9 {
		question.F = row[9] // 图片（Base64）
	}

	// 解析分类字段
	if len(row) > 10 {
		laVal, err := strconv.Atoi(row[10])
		if err == nil {
			question.LA = laVal
		} else {
			utils.Debug("ExcelImporter", "LA 解析失败", logrus.Fields{
				"value": row[10],
				"error": err,
			})
		}
	}
	if len(row) > 11 {
		lbVal, err := strconv.Atoi(row[11])
		if err == nil {
			question.LB = lbVal
		} else {
			utils.Debug("ExcelImporter", "LB 解析失败", logrus.Fields{
				"value": row[11],
				"error": err,
			})
		}
	}
	if len(row) > 12 {
		lcVal, err := strconv.Atoi(row[12])
		if err == nil {
			question.LC = lcVal
		} else {
			utils.Debug("ExcelImporter", "LC 解析失败", logrus.Fields{
				"value": row[12],
				"error": err,
			})
		}
	}

	// 解析题型
	if len(row) > 13 && row[13] != "" {
		typeVal, err := strconv.Atoi(row[13])
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
