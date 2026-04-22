package services

import (
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"fmt"

	"gorm.io/gorm"
)

// FavoriteService 收藏题目服务
type FavoriteService struct {
	favoriteDAO *dao.FavoriteQuestionDAO
	questionDAO *dao.QuestionDAO
}

// NewFavoriteService 创建 FavoriteService 实例
func NewFavoriteService(db *gorm.DB) *FavoriteService {
	return &FavoriteService{
		favoriteDAO: dao.NewFavoriteQuestionDAO(db),
		questionDAO: dao.NewQuestionDAO(db),
	}
}

// GetFavoriteQuestions 根据用户 ID 和分类获取收藏题目
// Python 原版：get_favorite_questions(user_id, category) -> List[Question]
func (s *FavoriteService) GetFavoriteQuestions(userID int64, category string) ([]*models.Question, error) {
	// 输入验证
	if userID <= 0 {
		utils.Warn("FavoriteService", "无效的用户 ID", map[string]interface{}{"user_id": userID})
		return nil, fmt.Errorf("无效的用户 ID")
	}
	validCategories := map[string]bool{"A": true, "B": true, "C": true}
	if !validCategories[category] {
		utils.Warn("FavoriteService", "无效的收藏类别", map[string]interface{}{"category": category})
		return nil, fmt.Errorf("无效的类别：%s（仅支持 A/B/C）", category)
	}

	utils.Info("FavoriteService", "获取收藏题目", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	// 获取收藏记录
	favoriteRecords, err := s.favoriteDAO.GetByUserAndCategory(userID, category)
	if err != nil {
		utils.Error("FavoriteService", "获取收藏记录失败", err, nil)
		return nil, err
	}

	utils.Info("FavoriteService", "获取收藏记录数量", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(favoriteRecords),
	})

	if len(favoriteRecords) == 0 {
		utils.Info("FavoriteService", "用户在该类别没有收藏记录", map[string]interface{}{
			"user_id":  userID,
			"category": category,
		})
		return []*models.Question{}, nil
	}

	// 获取题目 ID 列表
	questionIDs := make([]int64, 0, len(favoriteRecords))
	for _, record := range favoriteRecords {
		questionIDs = append(questionIDs, record.QuestionID)
	}

	// 根据 ID 列表获取题目对象
	questions := make([]*models.Question, 0, len(questionIDs))
	notFoundCount := 0
	for _, qID := range questionIDs {
		question, err := s.questionDAO.GetByID(qID)
		if err != nil {
			utils.Error("FavoriteService", "获取题目失败", err, map[string]interface{}{
				"question_id": qID,
			})
			continue
		}
		if question == nil {
			notFoundCount++
			utils.Info("FavoriteService", "题目不存在（创建占位记录）", map[string]interface{}{
				"question_id": qID,
			})
			// 创建占位题目，让前端显示记录存在但题目已删除
			placeholder := &models.Question{
				ID:   qID,
				J:    "",
				P:    "",
				I:    "",
				Q:    "[题目已被删除]",
				T:    "",
				A:    "",
				B:    "",
				C:    "",
				D:    "",
				F:    "",
				LA:   0,
				LB:   0,
				LC:   0,
				Type: 0,
			}
			questions = append(questions, placeholder)
			continue
		}
		questions = append(questions, question)
	}

	utils.Debug("FavoriteService", "获取收藏题目成功", map[string]interface{}{
		"user_id":      userID,
		"category":     category,
		"total_count":  len(favoriteRecords),
		"result_count": len(questions),
		"not_found":    notFoundCount,
	})

	return questions, nil
}

// AddFavoriteQuestion 添加收藏题目
// Python 原版：add_favorite_question(user_id, question_id, category) -> bool
func (s *FavoriteService) AddFavoriteQuestion(userID int64, questionID int64, category string) (bool, error) {
	utils.Debug("FavoriteService", "添加收藏题目", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
		"category":    category,
	})

	// 检查是否已经收藏
	existing, err := s.favoriteDAO.GetByUserAndQuestion(userID, questionID)
	if err != nil {
		utils.Error("FavoriteService", "检查收藏状态失败", err, nil)
		return false, err
	}

	if existing != nil {
		utils.Debug("FavoriteService", "用户已经收藏了该题目", map[string]interface{}{
			"user_id":     userID,
			"question_id": questionID,
		})
		return false, nil
	}

	// 创建新的收藏记录
	newFavorite := &models.FavoriteQuestion{
		UserID:     userID,
		QuestionID: questionID,
		Category:   category,
	}

	id, err := s.favoriteDAO.Create(newFavorite)
	if err != nil {
		utils.Error("FavoriteService", "添加收藏失败", err, nil)
		return false, err
	}

	utils.Debug("FavoriteService", "添加收藏题目成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
		"category":    category,
		"favorite_id": id,
	})

	return true, nil
}

// RemoveFavoriteQuestion 移除收藏题目
// Python 原版：remove_favorite_question(user_id, question_id) -> bool
func (s *FavoriteService) RemoveFavoriteQuestion(userID int64, questionID int64) (bool, error) {
	utils.Debug("FavoriteService", "移除收藏题目", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	err := s.favoriteDAO.DeleteByUserAndQuestion(userID, questionID)
	if err != nil {
		utils.Error("FavoriteService", "移除收藏失败", err, nil)
		return false, err
	}

	utils.Debug("FavoriteService", "移除收藏题目成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	return true, nil
}

// IsQuestionFavorited 检查题目是否已收藏
// Python 原版：is_question_favorited(user_id, question_id) -> bool
func (s *FavoriteService) IsQuestionFavorited(userID int64, questionID int64) (bool, error) {
	utils.Debug("FavoriteService", "检查收藏状态", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	favorite, err := s.favoriteDAO.GetByUserAndQuestion(userID, questionID)
	if err != nil {
		utils.Error("FavoriteService", "检查收藏状态失败", err, nil)
		return false, err
	}

	isFavorited := favorite != nil

	utils.Debug("FavoriteService", "收藏状态检查结果", map[string]interface{}{
		"user_id":      userID,
		"question_id":  questionID,
		"is_favorited": isFavorited,
	})

	return isFavorited, nil
}

// GetFavoriteCount 获取用户收藏总数
func (s *FavoriteService) GetFavoriteCount(userID int64, category string) (int, error) {
	utils.Debug("FavoriteService", "获取收藏数量", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	favorites, err := s.favoriteDAO.GetByUserAndCategory(userID, category)
	if err != nil {
		utils.Error("FavoriteService", "获取收藏数量失败", err, nil)
		return 0, err
	}

	count := len(favorites)

	utils.Debug("FavoriteService", "获取收藏数量成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    count,
	})

	return count, nil
}

// ClearUserFavorites 清空用户的收藏
func (s *FavoriteService) ClearUserFavorites(userID int64) error {
	utils.Info("FavoriteService", "清空用户收藏", map[string]interface{}{
		"user_id": userID,
	})

	err := s.favoriteDAO.ClearByUser(userID)
	if err != nil {
		utils.Error("FavoriteService", "清空用户收藏失败", err, nil)
		return err
	}

	utils.Info("FavoriteService", "清空用户收藏成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// ClearUserFavoritesByCategory 清空用户指定类别的收藏
func (s *FavoriteService) ClearUserFavoritesByCategory(userID int64, category string) error {
	utils.Info("FavoriteService", "清空用户类别收藏", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	err := s.favoriteDAO.DeleteByUserAndCategory(userID, category)
	if err != nil {
		utils.Error("FavoriteService", "清空用户类别收藏失败", err, nil)
		return err
	}

	utils.Info("FavoriteService", "清空用户类别收藏成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}
