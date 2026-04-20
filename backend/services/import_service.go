package services

import (
	"bytes"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

// ImportService 导入服务
type ImportService struct {
	questionDAO *dao.QuestionDAO
	pythonPath  string
	scriptDir   string
	db          *sql.DB
}

// ImportResult 导入结果
type ImportResult struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	ImportedCount int            `json:"imported_count"`
	TotalCount    int            `json:"total_count"`
	Stats         map[string]int `json:"stats"`
}

// NewImportService 创建 ImportService 实例
func NewImportService(db *sql.DB, pythonPath string, scriptDir string) *ImportService {
	return &ImportService{
		questionDAO: dao.NewQuestionDAO(db),
		db:          db,
		pythonPath:  pythonPath,
		scriptDir:   scriptDir,
	}
}

// ResetDatabase 重置数据库
func (s *ImportService) ResetDatabase() (bool, error) {
	utils.Info("ImportService", "开始重置数据库", nil)

	// 清空题库表
	err := s.questionDAO.ResetTable()
	if err != nil {
		utils.Error("ImportService", "重置数据库失败", err, nil)
		return false, err
	}

	utils.Info("ImportService", "数据库重置完成", nil)
	return true, nil
}

// ImportQuestions 导入题目到数据库
// Python 原版：import_questions(questions) -> Tuple[bool, str, int, Dict]
func (s *ImportService) ImportQuestions(questions []*models.Question) (*ImportResult, error) {
	utils.Info("ImportService", "开始导入题目", map[string]interface{}{
		"total": len(questions),
	})

	total := len(questions)
	if total == 0 {
		utils.Error("ImportService", "没有题目数据可导入", nil, nil)
		return &ImportResult{
			Success:       false,
			Message:       "题库文件为空",
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	// 重置数据库
	success, err := s.ResetDatabase()
	if err != nil || !success {
		utils.Error("ImportService", "重置数据库失败", err, nil)
		return &ImportResult{
			Success:       false,
			Message:       "重置题库表失败",
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, err
	}

	// 批量插入题目
	err = s.questionDAO.BatchInsert(questions)
	if err != nil {
		utils.Error("ImportService", "批量插入题目失败", err, nil)
		return &ImportResult{
			Success:       false,
			Message:       "批量插入题目失败",
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, err
	}

	utils.Info("ImportService", "题库导入完成", map[string]interface{}{
		"total": total,
	})

	// 统计信息
	stats := calculateImportStats(questions)

	// 将 map[string]int 转换为 map[string]interface{} 以适配 logrus.Fields
	statsInterface := make(map[string]interface{})
	for k, v := range stats {
		statsInterface[k] = v
	}
	utils.Info("ImportService", "统计信息", statsInterface)

	return &ImportResult{
		Success:       true,
		Message:       "导入成功",
		ImportedCount: total,
		TotalCount:    total,
		Stats:         stats,
	}, nil
}

// calculateImportStats 计算导入统计信息
func calculateImportStats(questions []*models.Question) map[string]int {
	stats := make(map[string]int)

	// 总计
	stats["total"] = len(questions)

	// A 类统计
	aTotal := 0
	aSingle := 0
	aMultiple := 0
	aWithImage := 0
	// B 类统计
	bTotal := 0
	bSingle := 0
	bMultiple := 0
	bWithImage := 0
	// C 类统计
	cTotal := 0
	cSingle := 0
	cMultiple := 0
	cWithImage := 0

	for _, q := range questions {
		isSingle := len(q.T) == 1
		isMultiple := len(q.T) > 1
		hasImage := q.F != ""

		// A 类
		if q.LA == 1 {
			aTotal++
			if isSingle {
				aSingle++
			}
			if isMultiple {
				aMultiple++
			}
			if hasImage {
				aWithImage++
			}
		}

		// B 类
		if q.LB == 1 {
			bTotal++
			if isSingle {
				bSingle++
			}
			if isMultiple {
				bMultiple++
			}
			if hasImage {
				bWithImage++
			}
		}

		// C 类
		if q.LC == 1 {
			cTotal++
			if isSingle {
				cSingle++
			}
			if isMultiple {
				cMultiple++
			}
			if hasImage {
				cWithImage++
			}
		}
	}

	stats["a_total"] = aTotal
	stats["a_single"] = aSingle
	stats["a_multiple"] = aMultiple
	stats["a_with_image"] = aWithImage

	stats["b_total"] = bTotal
	stats["b_single"] = bSingle
	stats["b_multiple"] = bMultiple
	stats["b_with_image"] = bWithImage

	stats["c_total"] = cTotal
	stats["c_single"] = cSingle
	stats["c_multiple"] = cMultiple
	stats["c_with_image"] = cWithImage

	return stats
}

// ProcessUnifiedData 根据文件类型自动处理数据并导入数据库（通过文件路径）
// Python 原版：process_unified_data(file_path, ...)
func (s *ImportService) ProcessUnifiedData(filePath string) (*ImportResult, error) {
	utils.Info("ImportService", "开始处理文件", map[string]interface{}{
		"file_path": filePath,
	})

	// 验证文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		utils.Error("ImportService", "获取文件信息失败", err, nil)
		return nil, fmt.Errorf("获取文件信息失败：%v", err)
	}
	if fileInfo.Size() > 100*1024*1024 { // 100MB限制
		utils.Error("ImportService", "文件过大", nil, map[string]interface{}{
			"size": fileInfo.Size(),
		})
		return &ImportResult{
			Success:       false,
			Message:       "文件过大，最大支持 100MB",
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	// 获取文件扩展名
	fileExt := filepath.Ext(filePath)
	fileExt = filepath.Base(fileExt) // 转为小写

	utils.Debug("ImportService", "文件类型", map[string]interface{}{
		"extension": fileExt,
	})

	// 根据文件类型选择对应的处理方法
	switch fileExt {
	case ".json":
		return s.processJSONData(filePath)
	case ".xlsx", ".xls":
		return s.processExcelData(filePath)
	case ".csv":
		return s.processCSVData(filePath)
	case ".zip":
		return s.processPDFData(filePath)
	default:
		utils.Error("ImportService", "不支持的文件类型", nil, map[string]interface{}{
			"extension": fileExt,
		})
		return &ImportResult{
			Success:       false,
			Message:       fmt.Sprintf("不支持的文件类型：%s", fileExt),
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}
}

// ProcessFileContent 处理文件内容导入（通过文件内容和类型）
func (s *ImportService) ProcessFileContent(content string, fileType string) (*ImportResult, error) {
	utils.Info("ImportService", "开始处理文件内容", map[string]interface{}{
		"type": fileType,
	})

	var questions []*models.Question
	var err error

	switch fileType {
	case "json":
		// JSON 可能包含字符串类型的 type 字段，需要特殊处理
		questions, err = s.parseJSONWithFlexibleType(content)
		if err != nil {
			return nil, fmt.Errorf("解析 JSON 失败：%v", err)
		}
	case "csv":
		importer := NewCSVImporter()
		questions, err = importer.ProcessCSVDataFromContent(content)
		if err != nil {
			return nil, err
		}
	case "xlsx", "xls":
		// Excel 需要保存为临时文件
		return s.importFromBase64Content(content, fileType)
	case "zip":
		// ZIP 需要保存为临时文件
		return s.importFromBase64Content(content, fileType)
	default:
		return &ImportResult{
			Success:       false,
			Message:       fmt.Sprintf("不支持的文件类型：%s", fileType),
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	return s.ImportQuestions(questions)
}

// importFromBase64Content 从 base64 编码的内容导入（用于 Excel 和 ZIP）
func (s *ImportService) importFromBase64Content(base64Content string, fileType string) (*ImportResult, error) {
	// 解码 base64 内容
	decoded, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return nil, fmt.Errorf("解码 base64 失败：%v", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "import_*."+fileType)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败：%v", err)
	}
	tempFilePath := tempFile.Name()
	defer tempFile.Close()
	defer os.Remove(tempFilePath)

	// 写入临时文件
	_, err = tempFile.Write(decoded)
	if err != nil {
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("写入临时文件失败：%v", err)
	}

	// 根据文件类型直接调用对应的处理方法
	switch fileType {
	case "xlsx", "xls":
		return s.processExcelData(tempFilePath)
	case "zip":
		return s.processPDFData(tempFilePath)
	default:
		return nil, fmt.Errorf("不支持的文件类型：%s", fileType)
	}
}

// processJSONData 处理 JSON 文件导入
func (s *ImportService) processJSONData(filePath string) (*ImportResult, error) {
	utils.Debug("ImportService", "处理 JSON 文件", map[string]interface{}{
		"file_path": filePath,
	})

	// 读取 JSON 文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		utils.Error("ImportService", "读取 JSON 文件失败", err, nil)
		return nil, err
	}

	// 解析 JSON 数据
	var questions []*models.Question
	err = json.Unmarshal(data, &questions)
	if err != nil {
		utils.Error("ImportService", "解析 JSON 数据失败", err, nil)
		return nil, err
	}

	// 导入题目
	return s.ImportQuestions(questions)
}

// processExcelData 处理 Excel 文件导入（Go 原生实现）
func (s *ImportService) processExcelData(filePath string) (*ImportResult, error) {
	utils.Debug("ImportService", "处理 Excel 文件（Go 原生）", map[string]interface{}{
		"file_path": filePath,
	})

	// 使用 Go 原生实现
	importer := NewExcelImporter()
	questions, err := importer.ProcessExcelData(filePath)
	if err != nil {
		return nil, err
	}

	return s.ImportQuestions(questions)
}

// processCSVData 处理 CSV 文件导入（Go 原生实现）
func (s *ImportService) processCSVData(filePath string) (*ImportResult, error) {
	utils.Debug("ImportService", "处理 CSV 文件（Go 原生）", map[string]interface{}{
		"file_path": filePath,
	})

	// 使用 Go 原生实现
	importer := NewCSVImporter()
	questions, err := importer.ProcessCSVData(filePath)
	if err != nil {
		return nil, err
	}

	return s.ImportQuestions(questions)
}

// processPDFData 处理 PDF 文件导入（Go 原生实现）
func (s *ImportService) processPDFData(filePath string) (*ImportResult, error) {
	utils.Debug("ImportService", "处理 PDF 文件（Go 原生）", map[string]interface{}{
		"file_path": filePath,
	})

	// 使用 Go 原生实现
	importer := NewPDFImporter()
	questions, err := importer.ProcessPDFData(filePath, nil)
	if err != nil {
		return nil, err
	}

	return s.ImportQuestions(questions)
}

// callPythonScript 调用 Python 脚本处理数据
func (s *ImportService) callPythonScript(scriptType string, filePath string) (*ImportResult, error) {
	// 确定 Python 脚本路径
	scriptPath := filepath.Join(s.scriptDir, fmt.Sprintf("import_%s.py", scriptType))

	// 检查 Python 脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		utils.Error("ImportService", "Python 脚本不存在", err, map[string]interface{}{
			"script_path": scriptPath,
		})
		return nil, fmt.Errorf("Python 脚本不存在：%s", scriptPath)
	}

	// 创建临时文件存储结果
	tempFile, err := os.CreateTemp("", "import_result_*.json")
	if err != nil {
		utils.Error("ImportService", "创建临时文件失败", err, nil)
		return nil, err
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)

	// 构建 Python 命令
	pythonScript := fmt.Sprintf(`
import sys
sys.path.insert(0, r'%s')
from import_%s import process_pdf_data
import json

result = process_pdf_data(r'%s')
with open(r'%s', 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False)
`, s.scriptDir, scriptType, filePath, tempFilePath)

	utils.Debug("ImportService", "执行 Python 脚本", map[string]interface{}{
		"script": scriptType,
		"file":   filePath,
	})

	// 执行 Python 脚本
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(s.pythonPath, "-c", pythonScript)
	} else {
		cmd = exec.Command("python3", "-c", pythonScript)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils.Error("ImportService", "执行 Python 脚本失败", err, map[string]interface{}{
			"stderr": stderr.String(),
		})
		return nil, fmt.Errorf("执行 Python 脚本失败：%v", err)
	}

	// 读取结果文件
	resultData, err := os.ReadFile(tempFilePath)
	if err != nil {
		utils.Error("ImportService", "读取结果文件失败", err, nil)
		return nil, err
	}

	// 解析结果
	var result map[string]interface{}
	err = json.Unmarshal(resultData, &result)
	if err != nil {
		utils.Error("ImportService", "解析结果数据失败", err, nil)
		return nil, err
	}

	// 检查是否成功
	success, _ := result["success"].(bool)
	if !success {
		message, _ := result["message"].(string)
		return &ImportResult{
			Success:       false,
			Message:       message,
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	// 提取题目数据并导入
	questionsData, ok := result["questions"].([]interface{})
	if !ok {
		return &ImportResult{
			Success:       false,
			Message:       "无法解析题目数据",
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	// 将 interface{} 转换为 []*models.Question
	questions := make([]*models.Question, 0)
	for _, q := range questionsData {
		qMap, ok := q.(map[string]interface{})
		if !ok {
			continue
		}

		question := &models.Question{}
		if v, ok := qMap["J"].(string); ok {
			question.J = v
		}
		if v, ok := qMap["P"].(string); ok {
			question.P = v
		}
		if v, ok := qMap["I"].(string); ok {
			question.I = v
		}
		if v, ok := qMap["Q"].(string); ok {
			question.Q = v
		}
		if v, ok := qMap["T"].(string); ok {
			question.T = v
		}
		if v, ok := qMap["A"].(string); ok {
			question.A = v
		}
		if v, ok := qMap["B"].(string); ok {
			question.B = v
		}
		if v, ok := qMap["C"].(string); ok {
			question.C = v
		}
		if v, ok := qMap["D"].(string); ok {
			question.D = v
		}
		if v, ok := qMap["F"].(string); ok {
			question.F = v
		}
		if v, ok := qMap["LA"].(int); ok {
			question.LA = v
		} else if v, ok := qMap["LA"].(float64); ok {
			question.LA = int(v)
		}
		if v, ok := qMap["LB"].(int); ok {
			question.LB = v
		} else if v, ok := qMap["LB"].(float64); ok {
			question.LB = int(v)
		}
		if v, ok := qMap["LC"].(int); ok {
			question.LC = v
		} else if v, ok := qMap["LC"].(float64); ok {
			question.LC = int(v)
		}
		if v, ok := qMap["type"].(float64); ok {
			question.Type = int(v)
		}

		questions = append(questions, question)
	}

	// 导入题目到数据库
	importResult, err := s.ImportQuestions(questions)
	if err != nil {
		return nil, err
	}

	return importResult, nil
}

// GetImportStats 获取导入统计信息
func (s *ImportService) GetImportStats() (map[string]interface{}, error) {
	// TODO: 实现统计信息获取
	return make(map[string]interface{}), nil
}

// ImportFromFileContent 从文件内容导入题目（仅用于 JSON 和 CSV）
func (s *ImportService) ImportFromFileContent(content string, fileType string) (*ImportResult, error) {
	utils.Info("ImportService", "开始从文件内容导入", map[string]interface{}{
		"type": fileType,
	})

	var questions []*models.Question
	var err error

	switch fileType {
	case "json":
		// JSON 可能包含字符串类型的 type 字段，需要特殊处理
		questions, err = s.parseJSONWithFlexibleType(content)
		if err != nil {
			return nil, fmt.Errorf("解析 JSON 失败：%v", err)
		}
	case "csv":
		importer := NewCSVImporter()
		questions, err = importer.ProcessCSVDataFromContent(content)
		if err != nil {
			return nil, err
		}
	default:
		return &ImportResult{
			Success:       false,
			Message:       fmt.Sprintf("不支持的文件类型：%s", fileType),
			ImportedCount: 0,
			TotalCount:    0,
			Stats:         make(map[string]int),
		}, nil
	}

	return s.ImportQuestions(questions)
}

// parseJSONWithFlexibleType 解析 JSON 数据，支持 type 字段为字符串或数字
func (s *ImportService) parseJSONWithFlexibleType(content string) ([]*models.Question, error) {
	// 先解析为通用 map 类型
	var rawData []map[string]interface{}
	err := json.Unmarshal([]byte(content), &rawData)
	if err != nil {
		return nil, err
	}

	questions := make([]*models.Question, 0, len(rawData))

	for _, item := range rawData {
		question := &models.Question{}

		// 解析基本字段
		if v, ok := item["J"].(string); ok {
			question.J = v
		}
		if v, ok := item["P"].(string); ok {
			question.P = v
		}
		if v, ok := item["I"].(string); ok {
			question.I = v
		}
		if v, ok := item["Q"].(string); ok {
			question.Q = v
		}
		if v, ok := item["T"].(string); ok {
			question.T = v
		}
		if v, ok := item["A"].(string); ok {
			question.A = v
		}
		if v, ok := item["B"].(string); ok {
			question.B = v
		}
		if v, ok := item["C"].(string); ok {
			question.C = v
		}
		if v, ok := item["D"].(string); ok {
			question.D = v
		}
		if v, ok := item["F"].(string); ok {
			question.F = v
		}
		// 解析 LA/LB/LC，支持 int 和字符串类型
		if v, ok := item["LA"].(int); ok {
			question.LA = v
		} else if v, ok := item["LA"].(string); ok {
			if val, err := strconv.Atoi(v); err == nil {
				question.LA = val
			}
		}
		if v, ok := item["LB"].(int); ok {
			question.LB = v
		} else if v, ok := item["LB"].(string); ok {
			if val, err := strconv.Atoi(v); err == nil {
				question.LB = val
			}
		}
		if v, ok := item["LC"].(int); ok {
			question.LC = v
		} else if v, ok := item["LC"].(string); ok {
			if val, err := strconv.Atoi(v); err == nil {
				question.LC = val
			}
		}

		// 解析 type 字段，支持字符串和数字两种类型
		if v, ok := item["type"]; ok {
			switch val := v.(type) {
			case float64:
				// JSON 数字在 Go 中默认为 float64
				question.Type = int(val)
			case string:
				// 字符串类型，尝试转换为数字
				typeVal, err := strconv.Atoi(val)
				if err == nil {
					question.Type = typeVal
				}
			case int:
				question.Type = val
			}
		}

		// 如果 Type 为 0，根据答案长度自动判断题型
		if question.Type == 0 && question.T != "" {
			if len(question.T) == 1 {
				question.Type = 1 // 单选题
			} else if len(question.T) > 1 {
				question.Type = 2 // 多选题
			}
		}

		questions = append(questions, question)
	}

	return questions, nil
}
