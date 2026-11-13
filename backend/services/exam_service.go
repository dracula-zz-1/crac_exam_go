package services

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

type ExamService struct {
	db *sql.DB
}

func NewExamService(db *sql.DB) *ExamService {
	return &ExamService{db: db}
}

func (s *ExamService) ServiceName() string { return "ExamService" }

func (s *ExamService) GetConfig(category string) (config.ExamConfig, error) {
	cfg, ok := config.ExamConfigFor(strings.ToUpper(category))
	if !ok {
		return config.ExamConfig{}, fmt.Errorf("无效的考试类别")
	}
	return cfg, nil
}

// CreateExam 创建考试：服务端记时、随机抽题、打乱选项、保存快照。
func (s *ExamService) CreateExam(userID int64, category string) (*models.ExamStartResponse, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("请先登录")
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	cfg, ok := config.ExamConfigFor(category)
	if !ok {
		return nil, fmt.Errorf("无效的考试类别")
	}

	singles, err := store.RandomQuestions(s.db, category, 1, cfg.Single)
	if err != nil {
		return nil, fmt.Errorf("抽取单选题失败：%w", err)
	}
	if len(singles) < cfg.Single {
		return nil, fmt.Errorf("%s 类单选题不足：需要 %d 题，当前题库只有 %d 题，请先导入题库", category, cfg.Single, len(singles))
	}
	multiples, err := store.RandomQuestions(s.db, category, 2, cfg.Multiple)
	if err != nil {
		return nil, fmt.Errorf("抽取多选题失败：%w", err)
	}
	if len(multiples) < cfg.Multiple {
		return nil, fmt.Errorf("%s 类多选题不足：需要 %d 题，当前题库只有 %d 题，请先导入题库", category, cfg.Multiple, len(multiples))
	}

	all := append(singles, multiples...)
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })

	startedAt := time.Now()
	examQuestions := make([]models.ExamQuestion, 0, len(all))
	details := make([]models.ExamDetail, 0, len(all))
	for _, q := range all {
		opts, correct := shuffleQuestionOptions(q)
		examQuestions = append(examQuestions, models.ExamQuestion{
			ID: q.ID, J: q.J, Q: q.Q,
			A: opts[0], B: opts[1], C: opts[2], D: opts[3],
			F: q.F, Type: q.Type,
		})
		details = append(details, models.ExamDetail{
			QuestionID: q.ID, QuestionText: q.Q,
			OptionA: opts[0], OptionB: opts[1], OptionC: opts[2], OptionD: opts[3],
			CorrectAnswer: correct, Type: q.Type, ImageData: q.F,
		})
	}

	examID, err := store.CreateExam(s.db, category, userID, cfg.Total, cfg.PassScore, startedAt, details)
	if err != nil {
		utils.Error("ExamService", "创建考试失败", err, "user_id", userID)
		return nil, fmt.Errorf("创建考试失败，请重试")
	}
	utils.Info("ExamService", "创建考试", "exam_id", examID, "category", category, "total", cfg.Total)

	return &models.ExamStartResponse{
		ExamID:           examID,
		Category:         category,
		Questions:        examQuestions,
		Total:            cfg.Total,
		Single:           cfg.Single,
		Multiple:         cfg.Multiple,
		TimeLimitSeconds: cfg.TimeMinutes * 60,
		PassScore:        cfg.PassScore,
		StartedAt:        startedAt,
	}, nil
}

// SubmitExam 交卷：服务端计时、重新判分、事务写入、幂等处理。
func (s *ExamService) SubmitExam(examID int64, answers map[string]string) (*models.ExamResult, error) {
	if examID <= 0 {
		return nil, fmt.Errorf("无效的考试")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	record, err := store.GetExamRecordTx(tx, examID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("考试记录不存在")
	}
	if record.Status == store.ExamSubmitted {
		return examResultFromRecord(record), nil
	}

	now := time.Now()
	duration := int(now.Sub(record.StartedAt).Seconds())
	if duration < 0 {
		duration = 0
	}
	if duration > 24*60*60 {
		duration = 24 * 60 * 60
	}

	details, err := store.GetExamDetailsTx(tx, examID)
	if err != nil {
		return nil, err
	}

	correctCount := 0
	totalCount := 0
	for i := range details {
		d := &details[i]
		totalCount++
		raw := answers[fmt.Sprintf("%d", d.QuestionID)]
		answer := normalizeAnswerLetters(raw)
		correct := answer != "" && answer == d.CorrectAnswer
		if correct {
			correctCount++
		} else {
			if err := store.RecordWrongTx(tx, record.UserID, d.QuestionID, record.Category); err != nil {
				return nil, err
			}
		}
		if err := store.UpdateDetailAnswer(tx, d.ID, answer, correct); err != nil {
			return nil, err
		}
	}

	if err := store.FinishExam(tx, examID, correctCount, correctCount, totalCount, duration, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	record.Score = correctCount
	record.CorrectCount = correctCount
	record.TotalQuestions = totalCount
	record.DurationSeconds = duration
	record.SubmittedAt = now
	record.Status = store.ExamSubmitted

	utils.Info("ExamService", "交卷", "exam_id", examID, "score", correctCount, "total", totalCount, "duration", duration)
	return examResultFromRecord(record), nil
}

func (s *ExamService) GetResult(examID int64) (*models.ExamResult, error) {
	record, err := store.GetExamRecord(s.db, examID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("考试记录不存在")
	}
	if record.Status != store.ExamSubmitted {
		return nil, fmt.Errorf("考试尚未交卷")
	}
	return examResultFromRecord(record), nil
}

// GetDetails 返回考试题目快照，用于结果页复盘。
func (s *ExamService) GetDetails(examID int64) ([]models.ExamDetail, error) {
	record, err := store.GetExamRecord(s.db, examID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("考试记录不存在")
	}
	return store.GetExamDetails(s.db, examID)
}

// InvalidateExam 主动放弃未交卷的考试。
func (s *ExamService) InvalidateExam(examID int64) error {
	record, err := store.GetExamRecord(s.db, examID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}
	if record.Status == store.ExamSubmitted {
		return fmt.Errorf("考试已交卷，无法作废")
	}
	if err := store.DeleteExam(s.db, examID); err != nil {
		return err
	}
	utils.Info("ExamService", "作废考试", "exam_id", examID)
	return nil
}

func (s *ExamService) GetUserExamRecords(userID int64, limit int) ([]models.ExamSummary, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("请先登录")
	}
	return store.ListExamSummaries(s.db, userID, limit)
}

func examResultFromRecord(r *store.ExamRecord) *models.ExamResult {
	return &models.ExamResult{
		ExamID:          r.ID,
		Category:        r.Category,
		ExamDate:        r.ExamDate,
		DurationSeconds: r.DurationSeconds,
		Score:           r.Score,
		CorrectCount:    r.CorrectCount,
		TotalCount:      r.TotalQuestions,
		PassScore:       r.PassScore,
		Passed:          r.Score >= r.PassScore,
	}
}

// shuffleQuestionOptions 打乱四个选项，并同步重排正确答案。
func shuffleQuestionOptions(q models.Question) ([4]string, string) {
	orig := [4]string{q.A, q.B, q.C, q.D}
	perm := [4]int{0, 1, 2, 3}
	rand.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })

	var opts [4]string
	var newAnswer []rune
	for newIdx, oldIdx := range perm {
		opts[newIdx] = orig[oldIdx]
		oldLetter := rune('A' + oldIdx)
		if strings.ContainsRune(q.T, oldLetter) {
			newAnswer = append(newAnswer, rune('A'+newIdx))
		}
	}
	sort.Slice(newAnswer, func(i, j int) bool { return newAnswer[i] < newAnswer[j] })
	return opts, string(newAnswer)
}

// normalizeAnswerLetters 去重、过滤 A-D 并排序。
func normalizeAnswerLetters(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	seen := map[rune]bool{}
	var letters []rune
	for _, r := range s {
		if r >= 'A' && r <= 'D' && !seen[r] {
			seen[r] = true
			letters = append(letters, r)
		}
	}
	sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
	return string(letters)
}
