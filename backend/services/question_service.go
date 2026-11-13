package services

import (
	"database/sql"
	"fmt"
	"strings"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

type QuestionService struct {
	db *sql.DB
}

func NewQuestionService(db *sql.DB) *QuestionService {
	return &QuestionService{db: db}
}

func (s *QuestionService) ServiceName() string { return "QuestionService" }

// GetPage 题库分页。多个类别筛选为“或”语义。
func (s *QuestionService) GetPage(page, pageSize int, search string, filterA, filterB, filterC bool) (*models.QuestionPage, error) {
	return store.PageQuestions(s.db, page, pageSize, search, filterA, filterB, filterC)
}

func (s *QuestionService) GetQuestion(id int64) (*models.Question, error) {
	return store.GetQuestion(s.db, id)
}

func (s *QuestionService) GetCategoryCounts() ([]models.CategoryCount, error) {
	return store.CategoryCounts(s.db)
}

// UpdateQuestion 整题更新。答案由服务端规范化，题型按答案数量推导。
func (s *QuestionService) UpdateQuestion(q models.Question) error {
	q.J = strings.TrimSpace(q.J)
	q.P = strings.TrimSpace(q.P)
	q.I = strings.TrimSpace(q.I)
	q.Q = strings.TrimSpace(q.Q)
	if q.ID <= 0 {
		return fmt.Errorf("无效的题目")
	}
	if q.J == "" {
		return fmt.Errorf("题号不能为空")
	}
	if q.Q == "" {
		return fmt.Errorf("题干不能为空")
	}
	exists, err := store.QuestionExistsByJPI(s.db, q.J, q.P, q.I, q.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("题号（J/P/I 联合）已被其他题目使用：%s / %s / %s", q.J, q.P, q.I)
	}
	if len(q.F) > 8<<20 {
		return fmt.Errorf("图片过大（上限 8MB）")
	}
	q.T = normalizeAnswerLetters(q.T)
	if q.T == "" {
		return fmt.Errorf("答案必须是 A-D 的组合")
	}
	if len(q.T) > 4 {
		return fmt.Errorf("答案格式无效")
	}
	if len(q.T) == 1 {
		q.Type = 1
	} else {
		q.Type = 2
	}
	if q.LA != 0 {
		q.LA = 1
	}
	if q.LB != 0 {
		q.LB = 1
	}
	if q.LC != 0 {
		q.LC = 1
	}
	if err := store.UpdateQuestion(s.db, &q); err != nil {
		utils.Error("QuestionService", "更新题目失败", err, "question_id", q.ID)
		return err
	}
	return nil
}

func (s *QuestionService) DeleteQuestion(id int64) error {
	if id <= 0 {
		return fmt.Errorf("无效的题目")
	}
	if err := store.DeleteQuestion(s.db, id); err != nil {
		utils.Error("QuestionService", "删除题目失败", err, "question_id", id)
		return err
	}
	utils.Info("QuestionService", "删除题目", "question_id", id)
	return nil
}
