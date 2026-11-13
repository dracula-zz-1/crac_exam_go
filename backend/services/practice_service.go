package services

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"strings"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
)

type PracticeService struct {
	db *sql.DB
}

func NewPracticeService(db *sql.DB) *PracticeService {
	return &PracticeService{db: db}
}

func (s *PracticeService) ServiceName() string { return "PracticeService" }

func (s *PracticeService) ensureUser(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("请先登录")
	}
	u, err := store.GetUser(s.db, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

func validCategory(category string) bool {
	switch strings.ToUpper(category) {
	case "A", "B", "C":
		return true
	}
	return false
}

// GetPracticeData 一次返回练习页所需的全部数据（题目、收藏集合、进度）。
func (s *PracticeService) GetPracticeData(userID int64, category, mode string) (*models.PracticeData, error) {
	if err := s.ensureUser(userID); err != nil {
		return nil, err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "practice"
	}

	if mode == "random" {
		if !validCategory(category) {
			return nil, fmt.Errorf("无效的类别")
		}
		return s.randomPracticeData(userID, category)
	}

	var questions []models.Question
	var err error
	switch mode {
	case "practice":
		if !validCategory(category) {
			return nil, fmt.Errorf("无效的类别")
		}
		questions, err = store.QuestionsByCategory(s.db, category)
	case "error":
		if !validCategory(category) {
			return nil, fmt.Errorf("无效的类别")
		}
		questions, err = store.ErrorQuestions(s.db, userID, category)
	case "favorite":
		questions, err = store.FavoriteQuestions(s.db, userID, category)
	default:
		return nil, fmt.Errorf("无效的练习模式")
	}
	if err != nil {
		return nil, err
	}

	favoriteIDs, err := store.FavoriteIDs(s.db, userID, "")
	if err != nil {
		return nil, err
	}

	progress := 0
	if mode == "practice" {
		progress, _, err = store.GetProgress(s.db, userID, category)
		if err != nil {
			return nil, err
		}
		if progress < 0 || progress >= len(questions) {
			progress = 0
		}
	}

	return &models.PracticeData{
		Mode:        mode,
		Category:    category,
		Questions:   questions,
		FavoriteIDs: favoriteIDs,
		Progress:    progress,
		Total:       len(questions),
	}, nil
}

// randomPracticeData 随机练习：
// 每轮包含全部题目至少一次；错题按 wrong_count 追加副本（最多 5 份），
// 因此错得越多出现频率越高，但不会挤掉未出现的题目。
func (s *PracticeService) randomPracticeData(userID int64, category string) (*models.PracticeData, error) {
	questions, err := store.QuestionsByCategory(s.db, category)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]models.Question, len(questions))
	ids := make([]int64, 0, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
		ids = append(ids, q.ID)
	}

	state, err := store.GetRandomCycle(s.db, userID, category)
	if err != nil {
		return nil, err
	}

	var cycle *store.RandomCycle
	if state != nil && len(state.IDs) > 0 && state.Position < len(state.IDs) {
		valid := true
		for _, id := range state.IDs {
			if _, ok := byID[id]; !ok {
				valid = false
				break
			}
		}
		if valid {
			cycle = state
		}
	}

	if cycle == nil && len(ids) > 0 {
		wrong, err := store.WrongCounts(s.db, userID, category)
		if err != nil {
			return nil, err
		}
		queue := append([]int64(nil), ids...)
		for _, id := range ids {
			extra := wrong[id]
			if extra > 10 {
				extra = 10
			}
			for i := 0; i < extra; i++ {
				queue = append(queue, id)
			}
		}
		rand.Shuffle(len(queue), func(i, j int) {
			queue[i], queue[j] = queue[j], queue[i]
		})
		cycleNo := 1
		if state != nil {
			cycleNo = state.CycleNo + 1
		}
		cycle = &store.RandomCycle{CycleNo: cycleNo, IDs: queue, Position: 0}
		if err := store.SaveRandomCycle(s.db, userID, category, cycle); err != nil {
			return nil, err
		}
	}

	ordered := make([]models.Question, 0, len(ids))
	progress := 0
	if cycle != nil {
		for _, id := range cycle.IDs {
			if q, ok := byID[id]; ok {
				ordered = append(ordered, q)
			}
		}
		progress = cycle.Position
		if progress < 0 || progress >= len(ordered) {
			progress = 0
		}
	}

	favIDs, err := store.FavoriteIDs(s.db, userID, "")
	if err != nil {
		return nil, err
	}
	return &models.PracticeData{
		Mode:        "random",
		Category:    category,
		Questions:   ordered,
		FavoriteIDs: favIDs,
		Progress:    progress,
		Total:       len(ordered),
	}, nil
}

// SaveRandomProgress 保存随机练习在一轮队列中的位置；index >= 队列长度表示本轮完成。
func (s *PracticeService) SaveRandomProgress(userID int64, category string, index int) error {
	if err := s.ensureUser(userID); err != nil {
		return err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	if !validCategory(category) {
		return fmt.Errorf("无效的类别")
	}
	if index < 0 {
		index = 0
	}
	state, err := store.GetRandomCycle(s.db, userID, category)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}
	if index >= len(state.IDs) {
		return store.DeleteRandomCycle(s.db, userID, category)
	}
	state.Position = index
	return store.SaveRandomCycle(s.db, userID, category, state)
}

func (s *PracticeService) SaveProgress(userID int64, category string, index int) error {
	if err := s.ensureUser(userID); err != nil {
		return err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	if !validCategory(category) {
		return fmt.Errorf("无效的类别")
	}
	if index < 0 {
		index = 0
	}
	return store.SaveProgress(s.db, userID, category, index)
}

func (s *PracticeService) AddError(userID, questionID int64, category string) (bool, error) {
	if err := s.ensureUser(userID); err != nil {
		return false, err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	if !validCategory(category) {
		return false, fmt.Errorf("无效的类别")
	}
	q, err := store.GetQuestion(s.db, questionID)
	if err != nil {
		return false, err
	}
	if q == nil {
		return false, fmt.Errorf("题目不存在")
	}
	return store.AddError(s.db, userID, questionID, category)
}

func (s *PracticeService) RemoveError(userID, questionID int64) error {
	if err := s.ensureUser(userID); err != nil {
		return err
	}
	return store.RemoveError(s.db, userID, questionID)
}

func (s *PracticeService) GetErrorQuestions(userID int64, category string) ([]models.Question, error) {
	if err := s.ensureUser(userID); err != nil {
		return nil, err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	return store.ErrorQuestions(s.db, userID, category)
}

func (s *PracticeService) AddFavorite(userID, questionID int64, category string) (bool, error) {
	if err := s.ensureUser(userID); err != nil {
		return false, err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	if !validCategory(category) {
		return false, fmt.Errorf("无效的类别")
	}
	q, err := store.GetQuestion(s.db, questionID)
	if err != nil {
		return false, err
	}
	if q == nil {
		return false, fmt.Errorf("题目不存在")
	}
	return store.AddFavorite(s.db, userID, questionID, category)
}

func (s *PracticeService) RemoveFavorite(userID, questionID int64) error {
	if err := s.ensureUser(userID); err != nil {
		return err
	}
	return store.RemoveFavorite(s.db, userID, questionID)
}

func (s *PracticeService) GetFavoriteQuestions(userID int64, category string) ([]models.Question, error) {
	if err := s.ensureUser(userID); err != nil {
		return nil, err
	}
	category = strings.ToUpper(strings.TrimSpace(category))
	return store.FavoriteQuestions(s.db, userID, category)
}

func (s *PracticeService) GetFavoriteIDs(userID int64) ([]int64, error) {
	if err := s.ensureUser(userID); err != nil {
		return nil, err
	}
	return store.FavoriteIDs(s.db, userID, "")
}
