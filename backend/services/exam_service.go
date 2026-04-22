package services

import (
	"crac_exam_go/backend/config"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// ExamService 考试服务
type ExamService struct {
	examRecordDAO      *dao.ExamRecordDAO
	examQuestionDetail *dao.ExamQuestionDetailDAO
	questionDAO        *dao.QuestionDAO
	errorQuestionDAO   *dao.ErrorQuestionDAO
	examConfig         map[string]config.ExamConfig
}

// ExamResult 考试结果
type ExamResult struct {
	ExamID          int64     `json:"exam_id"`
	Category        string    `json:"category"`
	ExamDate        time.Time `json:"exam_date"`
	DurationSeconds float64   `json:"duration_seconds"`
	Score           int       `json:"score"`
	CorrectCount    int       `json:"correct_count"`
	TotalCount      int       `json:"total_count"`
	PassExam        bool      `json:"pass_exam"`
	PassScore       int       `json:"pass_score"`
}

// ExamStartResponse 考试开始响应
type ExamStartResponse struct {
	ExamID    int64              `json:"exam_id"`
	Questions []*models.Question `json:"questions"`
	Config    config.ExamConfig  `json:"config"`
}

// QuestionDetail 题目详情
type QuestionDetail struct {
	ID     int64  `json:"id"`
	J      string `json:"J"`
	P      string `json:"P"`
	I      string `json:"I"`
	Q      string `json:"Q"`
	T      string `json:"T"`
	A      string `json:"A"`
	B      string `json:"B"`
	C      string `json:"C"`
	D      string `json:"D"`
	F      string `json:"F"`
	Type   int    `json:"type"`
	LA     int    `json:"LA"`
	LB     int    `json:"LB"`
	LC     int    `json:"LC"`
	UserID int64  `json:"user_id"`
}

// UserAnswer 用户答案
type UserAnswer struct {
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
}

// NewExamService 创建 ExamService 实例
func NewExamService(db *gorm.DB) *ExamService {
	return &ExamService{
		examRecordDAO:      dao.NewExamRecordDAO(db),
		examQuestionDetail: dao.NewExamQuestionDetailDAO(db),
		questionDAO:        dao.NewQuestionDAO(db),
		errorQuestionDAO:   dao.NewErrorQuestionDAO(db),
		examConfig:         config.EXAM_CONFIG,
	}
}

// CreateExam 创建一场新考试
// Python 原版：create_exam(user_id, category) -> (exam_id, questions, config)
func (s *ExamService) CreateExam(userID int64, category string) (*ExamStartResponse, error) {
	// 验证用户 ID
	if userID <= 0 {
		utils.Warn("ExamService", "无效的用户 ID", map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("无效的用户 ID")
	}

	// 验证类别参数
	if category == "" {
		utils.Warn("ExamService", "考试类别不能为空", nil)
		return nil, fmt.Errorf("考试类别不能为空")
	}
	if category != "A" && category != "B" && category != "C" {
		utils.Warn("ExamService", "无效的考试类别", map[string]interface{}{
			"category": category,
		})
		return nil, fmt.Errorf("无效的考试类别：%s（仅支持 A/B/C）", category)
	}
	// 输入验证
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户 ID：%d", userID)
	}

	// 验证 category 是否有效
	validCategories := map[string]bool{"A": true, "B": true, "C": true}
	if !validCategories[category] {
		return nil, fmt.Errorf("无效的考试类别：%s，支持的类别：A、B、C", category)
	}

	// 获取配置
	examConfig := s.examConfig[category]
	if examConfig.Total == 0 {
		examConfig = s.examConfig["A"]
	}

	utils.Info("ExamService", "开始创建考试", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"config":   examConfig,
	})

	// 直接从数据库随机获取题目
	selectedSingle, err := s.questionDAO.GetRandomByCategoryAndType(category, 1, examConfig.Single)
	if err != nil {
		utils.Error("ExamService", "获取单选题失败", err, nil)
		return nil, err
	}

	selectedMultiple, err := s.questionDAO.GetRandomByCategoryAndType(category, 2, examConfig.Multiple)
	if err != nil {
		utils.Error("ExamService", "获取多选题失败", err, nil)
		return nil, err
	}

	// 计算实际可用的题目数量
	actualSingle := len(selectedSingle)
	actualMultiple := len(selectedMultiple)

	utils.Debug("ExamService", "获取题目数量", map[string]interface{}{
		"single_count":   actualSingle,
		"multiple_count": actualMultiple,
	})

	// 检查是否有足够的题目
	if actualSingle+actualMultiple == 0 {
		utils.Error("ExamService", "题库中没有足够的题目", fmt.Errorf("类别 %s 题目数量为0", category), nil)
		return nil, fmt.Errorf("题库中没有足够的题目（类别: %s）", category)
	}

	// 更新配置
	examConfig.Single = actualSingle
	examConfig.Multiple = actualMultiple
	examConfig.Total = actualSingle + actualMultiple

	// 合并题目并打乱顺序
	examQuestions := append(selectedSingle, selectedMultiple...)
	s.shuffleQuestions(examQuestions)

	// 打乱每个题目的选项顺序并更新正确答案
	shuffledQuestions := s.shuffleOptions(examQuestions)

	// 创建考试记录
	examRecord := &models.ExamRecord{
		Category:        category,
		ExamDate:        time.Now(),
		UserID:          userID,
		DurationSeconds: 0,
		Score:           0,
		TotalQuestions:  examConfig.Total,
		CorrectCount:    0,
	}

	examID, err := s.examRecordDAO.Create(examRecord)
	if err != nil {
		utils.Error("ExamService", "创建考试记录失败", err, nil)
		return nil, err
	}

	// 检查考试记录是否创建成功
	if examID == 0 {
		utils.Error("ExamService", "创建考试记录返回ID为0", fmt.Errorf("数据库插入失败"), nil)
		return nil, fmt.Errorf("创建考试记录失败")
	}

	// 保存考试题目详情
	err = s.saveExamQuestionDetails(examID, shuffledQuestions)
	if err != nil {
		utils.Error("ExamService", "保存考试题目详情失败", err, nil)
		return nil, err
	}

	utils.Info("ExamService", "创建考试成功", map[string]interface{}{
		"exam_id": examID,
		"total":   examConfig.Total,
	})

	// 返回结构体
	return &ExamStartResponse{
		ExamID:    examID,
		Questions: shuffledQuestions,
		Config:    examConfig,
	}, nil
}

// shuffleQuestions 打乱题目顺序
func (s *ExamService) shuffleQuestions(questions []*models.Question) {
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
}

// shuffleOptions 打乱题目选项并更新正确答案
// Python 原版：_shuffle_options(questions)
func (s *ExamService) shuffleOptions(questions []*models.Question) []*models.Question {
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
		s.shuffleOptionsList(options)

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

	return questions
}

// shuffleOptionsList 打乱选项列表
func (s *ExamService) shuffleOptionsList(options []map[string]string) {
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})
}

// saveExamQuestionDetails 保存考试题目详情
// Python 原版：_save_exam_question_details(exam_id, questions)
func (s *ExamService) saveExamQuestionDetails(examID int64, questions []*models.Question) error {
	if len(questions) == 0 {
		return nil
	}

	// 创建考试题目详情列表
	examQuestionDetails := make([]*models.ExamQuestionDetail, 0)
	for _, question := range questions {
		examQuestionDetail := &models.ExamQuestionDetail{
			ExamID:        examID,
			QuestionID:    question.ID,
			QuestionText:  question.Q,
			OptionA:       question.A,
			OptionB:       question.B,
			OptionC:       question.C,
			OptionD:       question.D,
			CorrectAnswer: question.T,
			Type:          question.Type,
			ImageData:     question.F,
			UserAnswer:    "",
			IsCorrect:     false,
		}
		examQuestionDetails = append(examQuestionDetails, examQuestionDetail)
	}

	// 使用批量插入
	err := s.examQuestionDetail.BulkCreate(examQuestionDetails)
	if err != nil {
		return err
	}

	utils.Debug("ExamService", "保存考试题目详情成功", map[string]interface{}{
		"exam_id": examID,
		"count":   len(examQuestionDetails),
	})

	return nil
}

// GetExamRecordByID 根据考试 ID 获取考试记录
// Python 原版：get_exam_record_by_id(exam_id)
func (s *ExamService) GetExamRecordByID(examID int64) (*models.ExamRecord, error) {
	return s.examRecordDAO.GetByID(examID)
}

// GetUserExamRecords 获取用户考试记录
// Python 原版：get_user_exam_records(user_id, category, limit)
func (s *ExamService) GetUserExamRecords(userID int64, category string, limit int) ([]*models.ExamRecord, error) {
	if category != "" {
		return s.examRecordDAO.GetByUserAndCategory(userID, category)
	}
	return s.examRecordDAO.GetRecentExams(userID, limit)
}

// SubmitExam 提交考试
// Python 原版：submit_exam(exam_id, user_answers, start_time)
func (s *ExamService) SubmitExam(examID int64, userAnswers map[int64]UserAnswer, startTime time.Time) (*ExamResult, error) {
	// 计算考试时长
	durationSeconds := int64(time.Since(startTime).Seconds())

	// 获取考试记录
	examRecord, err := s.examRecordDAO.GetByID(examID)
	if err != nil {
		utils.Error("ExamService", "获取考试记录失败", err, nil)
		return nil, err
	}

	if examRecord == nil {
		return nil, fmt.Errorf("考试记录不存在 (ID: %d)", examID)
	}

	// 更新考试记录时长
	examRecord.DurationSeconds = float64(durationSeconds)
	err = s.examRecordDAO.Update(examRecord)
	if err != nil {
		utils.Error("ExamService", "更新考试记录时长失败", err, nil)
		return nil, err
	}

	// 更新题目详情中的用户答案和正确性
	correctCount := 0
	totalCount := 0
	userID := examRecord.UserID
	category := examRecord.Category

	examQuestionDetails, err := s.examQuestionDetail.GetByExamID(examID)
	if err != nil {
		utils.Error("ExamService", "获取考试题目详情失败", err, nil)
		return nil, err
	}

	// 收集需要更新的题目详情和错题
	var updatedDetails []*models.ExamQuestionDetail
	var newErrorQuestions []*models.ErrorQuestion
	var wrongQuestionIDs []int64

	// 第一次遍历：判断答案正确性，收集错误题目 ID
	for _, detail := range examQuestionDetails {
		questionID := detail.QuestionID
		totalCount++

		if userAnswer, exists := userAnswers[questionID]; exists {
			detail.UserAnswer = userAnswer.Answer

			// ✅ 修复 BUG-001: 后端重新判断答案正确性，不依赖前端传入的 IsCorrect
			isCorrect := s.isAnswerCorrect(userAnswer.Answer, detail.CorrectAnswer, detail.Type)
			detail.IsCorrect = isCorrect

			if isCorrect {
				correctCount++
			} else {
				// 回答错误，收集题目 ID
				wrongQuestionIDs = append(wrongQuestionIDs, questionID)
			}
		} else {
			// 未作答，标记为错误并收集题目 ID
			detail.UserAnswer = ""
			detail.IsCorrect = false
			wrongQuestionIDs = append(wrongQuestionIDs, questionID)
		}

		updatedDetails = append(updatedDetails, detail)
	}

	// 批量查询错题本，减少 N+1 查询问题
	existingErrors, err := s.errorQuestionDAO.BatchGetByUserQuestionAndCategory(userID, category, wrongQuestionIDs)
	if err != nil {
		utils.Error("ExamService", "批量查询错题失败", err, nil)
		return nil, err
	}

	// 第二次遍历：根据查询结果构建错题列表
	for _, detail := range updatedDetails {
		questionID := detail.QuestionID
		if !detail.IsCorrect && !existingErrors[questionID] {
			// 添加到错题本
			newErrorQuestions = append(newErrorQuestions, &models.ErrorQuestion{
				QuestionID: questionID,
				Category:   category,
				UserID:     userID,
			})
		}
	}

	// 批量更新题目详情
	err = s.examQuestionDetail.BulkUpdate(updatedDetails)
	if err != nil {
		utils.Error("ExamService", "批量更新题目详情失败", err, nil)
		return nil, err
	}

	// 批量添加错题
	if len(newErrorQuestions) > 0 {
		err = s.errorQuestionDAO.BulkCreate(newErrorQuestions)
		if err != nil {
			utils.Error("ExamService", "批量添加错题失败", err, nil)
		} else {
			utils.Debug("ExamService", "批量添加错题成功", map[string]interface{}{
				"count": len(newErrorQuestions),
			})
		}
	}

	// 更新考试记录的分数
	examRecord.Score = correctCount
	examRecord.CorrectCount = correctCount
	examRecord.TotalQuestions = totalCount
	err = s.examRecordDAO.Update(examRecord)
	if err != nil {
		utils.Error("ExamService", "更新考试记录分数失败", err, nil)
		return nil, err
	}

	// 获取配置判断是否通过
	examConfig := s.examConfig[category]
	if examConfig.Total == 0 {
		examConfig = s.examConfig["A"]
	}
	passExam := correctCount >= examConfig.PassScore

	utils.Info("ExamService", "考试提交成功", map[string]interface{}{
		"exam_id":       examID,
		"correct_count": correctCount,
		"total_count":   totalCount,
		"pass_exam":     passExam,
		"duration":      durationSeconds,
	})

	// 返回考试结果
	return &ExamResult{
		ExamID:          examID,
		Category:        category,
		ExamDate:        examRecord.ExamDate,
		DurationSeconds: float64(durationSeconds),
		Score:           correctCount,
		CorrectCount:    correctCount,
		TotalCount:      totalCount,
		PassExam:        passExam,
		PassScore:       examConfig.PassScore,
	}, nil
}

// GetExamResult 获取考试结果
// Python 原版：get_exam_result(exam_id)
func (s *ExamService) GetExamResult(examID int64) (*ExamResult, error) {
	examRecord, err := s.examRecordDAO.GetByID(examID)
	if err != nil {
		utils.Error("ExamService", "获取考试记录失败", err, nil)
		return nil, err
	}

	if examRecord == nil {
		return nil, fmt.Errorf("考试记录不存在 (ID: %d)", examID)
	}

	examQuestionDetails, err := s.examQuestionDetail.GetByExamID(examID)
	if err != nil {
		utils.Error("ExamService", "获取考试题目详情失败", err, nil)
		return nil, err
	}

	correctCount := 0
	for _, detail := range examQuestionDetails {
		if detail.IsCorrect {
			correctCount++
		}
	}

	totalCount := len(examQuestionDetails)

	// 获取配置
	examConfig := s.examConfig[examRecord.Category]
	if examConfig.Total == 0 {
		examConfig = s.examConfig["A"]
	}
	passExam := correctCount >= examConfig.PassScore

	return &ExamResult{
		ExamID:          examID,
		Category:        examRecord.Category,
		ExamDate:        examRecord.ExamDate,
		DurationSeconds: examRecord.DurationSeconds,
		Score:           correctCount,
		CorrectCount:    correctCount,
		TotalCount:      totalCount,
		PassExam:        passExam,
		PassScore:       examConfig.PassScore,
	}, nil
}

// GetExamQuestions 获取考试题目
// Python 原版：get_exam_questions(exam_id)
func (s *ExamService) GetExamQuestions(examID int64) ([]*models.ExamQuestionDetail, error) {
	return s.examQuestionDetail.GetByExamID(examID)
}

// isAnswerCorrect 判断答案是否正确
// ✅ 修复 BUG-001: 后端重新判断答案正确性，不依赖前端传入的 IsCorrect
func (s *ExamService) isAnswerCorrect(userAnswer, correctAnswer string, questionType int) bool {
	if userAnswer == "" || correctAnswer == "" {
		return false
	}

	// 排序后比较（确保答案顺序不影响判断）
	return utils.SortString(userAnswer) == utils.SortString(correctAnswer)
}

// InvalidateExam 作废考试记录（用于用户退出未完成的考试）
// ✅ 修复 BUG-002: 添加考试退出作废功能
func (s *ExamService) InvalidateExam(examID int64) error {
	utils.Info("ExamService", "开始作废考试记录", map[string]interface{}{
		"exam_id": examID,
	})

	// 开启事务
	tx := s.examRecordDAO.GetDB().Begin()
	if tx.Error != nil {
		utils.Error("ExamService", "开启事务失败", tx.Error, map[string]interface{}{
			"exam_id": examID,
		})
		return tx.Error
	}

	// 添加 panic 恢复，确保事务回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.Error("ExamService", "发生 panic，事务已回滚", fmt.Errorf("%v", r), map[string]interface{}{
				"exam_id": examID,
			})
			panic(r)
		}
	}()

	// 1. 删除考试题目详情
	err := s.examQuestionDetail.DeleteByExamIDWithTx(examID, tx)
	if err != nil {
		tx.Rollback()
		utils.Error("ExamService", "删除考试题目详情失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	// 2. 删除考试记录
	err = s.examRecordDAO.DeleteWithTx(examID, tx)
	if err != nil {
		tx.Rollback()
		utils.Error("ExamService", "删除考试记录失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.Error("ExamService", "提交事务失败", err, map[string]interface{}{
			"exam_id": examID,
		})
		return err
	}

	utils.Info("ExamService", "考试记录作废成功", map[string]interface{}{
		"exam_id": examID,
	})
	return nil
}

// GetExamConfig 获取考试配置
// Python 原版：get_exam_config(category)
func (s *ExamService) GetExamConfig(category string) *config.ExamConfig {
	examConfig, exists := s.examConfig[category]
	if !exists {
		examConfig = s.examConfig["A"]
	}
	return &examConfig
}
