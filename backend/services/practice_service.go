package services

import (
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PracticeService 练习服务
type PracticeService struct {
	questionDAO         *dao.QuestionDAO
	practiceProgressDAO *dao.PracticeProgressDAO
	errorQuestionDAO    *dao.ErrorQuestionDAO
}

// NewPracticeService 创建 PracticeService 实例
func NewPracticeService(db *gorm.DB) *PracticeService {
	return &PracticeService{
		questionDAO:         dao.NewQuestionDAO(db),
		practiceProgressDAO: dao.NewPracticeProgressDAO(db),
		errorQuestionDAO:    dao.NewErrorQuestionDAO(db),
	}
}

// GetQuestionsByCategory 根据类别获取题目
// Python 原版：get_questions_by_category(category)
func (s *PracticeService) GetQuestionsByCategory(category string) ([]*models.Question, error) {
	utils.Info("PracticeService", "获取类别题目", map[string]interface{}{
		"category": category,
	})

	questions, err := s.questionDAO.GetByCategory(category)
	if err != nil {
		utils.Error("PracticeService", "获取类别题目失败", err, nil)
		return nil, err
	}

	utils.Debug("PracticeService", "获取类别题目成功", map[string]interface{}{
		"category": category,
		"count":    len(questions),
	})

	return questions, nil
}

// GetErrorQuestions 根据用户 ID 和类别获取错题，并关联查询题目详情
// Python 原版：get_error_questions(user_id, category) -> List[Question]
func (s *PracticeService) GetErrorQuestions(userID int64, category string) ([]*models.Question, error) {
	// 输入验证
	if userID <= 0 {
		utils.Warn("PracticeService", "无效的用户 ID", map[string]interface{}{"user_id": userID})
		return nil, fmt.Errorf("无效的用户 ID")
	}
	validCategories := map[string]bool{"A": true, "B": true, "C": true}
	if !validCategories[category] {
		utils.Warn("PracticeService", "无效的错题类别", map[string]interface{}{"category": category})
		return nil, fmt.Errorf("无效的类别：%s（仅支持 A/B/C）", category)
	}

	utils.Info("PracticeService", "获取用户错题", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	errorQuestions, err := s.errorQuestionDAO.GetErrorQuestionsWithDetails(userID, category)
	if err != nil {
		utils.Error("PracticeService", "获取错题失败", err, map[string]interface{}{
			"user_id":  userID,
			"category": category,
		})
		return nil, err
	}

	utils.Info("PracticeService", "获取错题记录数量", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(errorQuestions),
	})

	if len(errorQuestions) == 0 {
		utils.Info("PracticeService", "用户在该类别没有错题记录", map[string]interface{}{
			"user_id":  userID,
			"category": category,
		})
		return []*models.Question{}, nil
	}

	// 将 ErrorQuestion 转换为 Question
	questions := make([]*models.Question, 0, len(errorQuestions))
	for _, eq := range errorQuestions {
		question := &models.Question{
			ID:   eq.QuestionID, // 使用题目 ID 而不是错题记录 ID
			J:    eq.J,
			P:    eq.P,
			I:    eq.I,
			Q:    eq.Q,
			T:    eq.T,
			A:    eq.A,
			B:    eq.B,
			C:    eq.C,
			D:    eq.D,
			F:    eq.F,
			LA:   eq.LA,
			LB:   eq.LB,
			LC:   eq.LC,
			Type: eq.Type,
		}
		questions = append(questions, question)
	}

	utils.Debug("PracticeService", "获取错题成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(questions),
	})

	// 打印第一条错题的详细信息用于调试
	if len(questions) > 0 {
		first := questions[0]
		// 安全截取题干前 50 个字符
		qText := first.Q
		if len(qText) > 50 {
			qText = qText[:50]
		}
		utils.Info("PracticeService", "第一条错题", map[string]interface{}{
			"id":   first.ID,
			"type": first.Type,
			"Q":    qText,
			"T":    first.T,
		})
	}

	return questions, nil
}

// ShuffleOptions 打乱题目选项并更新正确答案
// Python 原版：shuffle_options(questions)
func (s *PracticeService) ShuffleOptions(questions []*models.Question) []*models.Question {
	for _, question := range questions {
		// 保存原始选项和正确答案
		originalOptions := map[string]string{
			"A": question.A,
			"B": question.B,
			"C": question.C,
			"D": question.D,
		}
		correctAnswer := question.T

		// 创建选项列表（保留键值对）并打乱
		options := make([]map[string]string, 0)
		for k, v := range originalOptions {
			options = append(options, map[string]string{"key": k, "value": v})
		}
		utils.ShuffleOptions(options)

		// 重新赋值打乱后的选项
		if len(options) >= 4 {
			question.A = options[0]["value"]
			question.B = options[1]["value"]
			question.C = options[2]["value"]
			question.D = options[3]["value"]

			// 创建原始选项到新选项的映射
			originalToNew := make(map[string]string)
			for i, opt := range options {
				originalToNew[opt["key"]] = string(rune(65 + i)) // A=65, B=66, C=67, D=68
			}

			// 更新正确答案
			newCorrect := ""
			for _, char := range correctAnswer {
				// 过滤掉非字母字符（如空格）
				if char >= 'A' && char <= 'Z' {
					newCorrect += originalToNew[string(char)]
				}
			}

			// 对新的正确答案进行排序（多选题）
			if len(newCorrect) > 1 {
				newCorrect = utils.SortString(newCorrect)
			}

			question.T = newCorrect
		}
	}

	utils.Debug("PracticeService", "打乱题目选项", map[string]interface{}{
		"count": len(questions),
	})

	return questions
}

// GetPracticeProgress 获取练习进度
// Python 原版：get_practice_progress(user_id, category) -> int
func (s *PracticeService) GetPracticeProgress(userID int64, category string) (int, error) {
	utils.Debug("PracticeService", "获取练习进度", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	progress, err := s.practiceProgressDAO.GetByUserAndCategory(userID, category)
	if err != nil {
		utils.Error("PracticeService", "获取练习进度失败", err, nil)
		return 0, err
	}

	if progress != nil {
		utils.Debug("PracticeService", "获取练习进度成功", map[string]interface{}{
			"user_id":       userID,
			"category":      category,
			"current_index": progress.CurrentIndex,
		})
		return progress.CurrentIndex, nil
	}

	utils.Debug("PracticeService", "练习进度不存在，返回 0", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})
	return 0, nil
}

// SavePracticeProgress 保存练习进度
// Python 原版：save_practice_progress(user_id, category, index)
func (s *PracticeService) SavePracticeProgress(userID int64, category string, index int) error {
	utils.Debug("PracticeService", "保存练习进度", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"index":    index,
	})

	// 使用 DAO 的 UpdateByUserAndCategory 方法，它会自动判断是更新还是创建
	now := time.Now()
	err := s.practiceProgressDAO.UpdateByUserAndCategory(userID, category, index, now)
	if err != nil {
		utils.Error("PracticeService", "保存练习进度失败", err, nil)
		return err
	}

	utils.Debug("PracticeService", "保存练习进度成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"index":    index,
	})

	return nil
}

// AddErrorQuestion 添加错题到错题本
// Python 原版：add_error_question(question_id, category, user_id)
func (s *PracticeService) AddErrorQuestion(questionID int64, category string, userID int64) (bool, error) {
	utils.Debug("PracticeService", "添加错题", map[string]interface{}{
		"question_id": questionID,
		"user_id":     userID,
		"category":    category,
	})

	// 检查题目是否已存在于错题本
	existingError, err := s.errorQuestionDAO.GetByUserQuestionAndCategory(userID, questionID, category)
	if err != nil {
		utils.Error("PracticeService", "检查错题是否存在失败", err, nil)
		return false, err
	}

	if existingError != nil {
		utils.Info("PracticeService", "题目已存在于错题本", map[string]interface{}{
			"question_id": questionID,
			"user_id":     userID,
			"category":    category,
		})
		return false, nil
	}

	// 创建错题实体（只存储引用信息，不再复制题目详情）
	errorQuestion := &models.ErrorQuestion{
		QuestionID: questionID,
		Category:   category,
		UserID:     userID,
	}

	// 添加新错题
	id, err := s.errorQuestionDAO.Create(errorQuestion)
	if err != nil {
		utils.Error("PracticeService", "添加错题失败", err, nil)
		return false, err
	}

	utils.Debug("PracticeService", "添加错题成功", map[string]interface{}{
		"question_id": questionID,
		"user_id":     userID,
		"category":    category,
		"error_id":    id,
	})

	return true, nil
}

// GetRandomQuestions 随机获取指定类别和题型的题目
// Python 原版：questions_dao.get_random_questions_by_category_and_type(category, type, count)
func (s *PracticeService) GetRandomQuestions(category string, questionType int, count int) ([]*models.Question, error) {
	utils.Debug("PracticeService", "随机获取题目", map[string]interface{}{
		"category":      category,
		"question_type": questionType,
		"count":         count,
	})

	questions, err := s.questionDAO.GetRandomByCategoryAndType(category, questionType, count)
	if err != nil {
		utils.Error("PracticeService", "随机获取题目失败", err, nil)
		return nil, err
	}

	utils.Debug("PracticeService", "随机获取题目成功", map[string]interface{}{
		"category":      category,
		"question_type": questionType,
		"count":         len(questions),
	})

	return questions, nil
}

// ResetProgress 重置用户练习进度
// Python 原版：reset_progress(user_id, category)
func (s *PracticeService) ResetProgress(userID int64, category string) error {
	utils.Info("PracticeService", "重置用户练习进度", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	if category == "all" {
		// 重置所有类别
		err := s.practiceProgressDAO.ClearByUser(userID)
		if err != nil {
			utils.Error("PracticeService", "重置所有类别进度失败", err, nil)
			return err
		}
	} else {
		// 重置指定类别
		err := s.practiceProgressDAO.DeleteByUserAndCategory(userID, category)
		if err != nil {
			utils.Error("PracticeService", "重置指定类别进度失败", err, nil)
			return err
		}
	}

	utils.Debug("PracticeService", "重置用户练习进度成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})
	return nil
}
