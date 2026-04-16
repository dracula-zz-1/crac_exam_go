package config

// ExamConfig 考试配置结构
type ExamConfig struct {
	Total       int `json:"total"`        // 总题数
	Single      int `json:"single"`       // 单选题数量
	Multiple    int `json:"multiple"`     // 多选题数量
	TimeMinutes int `json:"time_minutes"` // 考试时间（分钟）
	PassScore   int `json:"pass_score"`   // 通过分数（答对题数）
}

// EXAM_CONFIG 考试配置（与 Python 版本完全一致）
// 来源：src/configs/config.py 第 40-44 行
var EXAM_CONFIG = map[string]ExamConfig{
	"A": {
		Total:       40,
		Single:      32,
		Multiple:    8,
		TimeMinutes: 40,
		PassScore:   30,
	},
	"B": {
		Total:       60,
		Single:      45,
		Multiple:    15,
		TimeMinutes: 60,
		PassScore:   45,
	},
	"C": {
		Total:       90,
		Single:      70,
		Multiple:    20,
		TimeMinutes: 90,
		PassScore:   70,
	},
}

// GetExamConfig 获取指定类别的考试配置
func GetExamConfig(category string) ExamConfig {
	if config, exists := EXAM_CONFIG[category]; exists {
		return config
	}
	// 默认返回 A 类配置
	return EXAM_CONFIG["A"]
}

// QuestionConfig 题目相关配置
var QuestionConfig = struct {
	// PageSize 题库管理每页显示题数
	PageSize int
}{
	PageSize: 20, // 与 Python 版本一致
}
