package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
)

// FavoriteQuestionDAO 收藏题目数据访问对象
type FavoriteQuestionDAO struct {
	*BaseDAO
}

// NewFavoriteQuestionDAO 创建 FavoriteQuestionDAO 实例
func NewFavoriteQuestionDAO(db *sql.DB) *FavoriteQuestionDAO {
	return &FavoriteQuestionDAO{
		BaseDAO: NewBaseDAO(db, "favorite_questions"),
	}
}

// Create 创建收藏题目
func (dao *FavoriteQuestionDAO) Create(favorite *models.FavoriteQuestion) (int64, error) {
	query := `INSERT INTO favorite_questions (question_id, category, user_id) VALUES (?, ?, ?)`

	result, err := dao.ExecuteUpdate(query, favorite.QuestionID, favorite.Category, favorite.UserID)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	favorite.ID = id
	utils.Debug("FavoriteQuestionDAO", "创建收藏题目成功", map[string]interface{}{
		"favorite_id": favorite.ID,
		"question_id": favorite.QuestionID,
		"user_id":     favorite.UserID,
		"category":    favorite.Category,
	})

	return favorite.ID, nil
}

// GetByID 根据 ID 获取收藏
func (dao *FavoriteQuestionDAO) GetByID(id int64) (*models.FavoriteQuestion, error) {
	query := `SELECT id, question_id, category, user_id FROM favorite_questions WHERE id = ?`

	row := dao.QueryRow(query, id)
	favorite, err := dao.scanFavorite(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return favorite, nil
}

// GetByUserID 根据用户 ID 获取所有收藏
func (dao *FavoriteQuestionDAO) GetByUserID(userID int64) ([]*models.FavoriteQuestion, error) {
	query := `SELECT id, question_id, category, user_id FROM favorite_questions WHERE user_id = ?`

	rows, err := dao.ExecuteQuery(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*models.FavoriteQuestion
	for rows.Next() {
		favorite, err := dao.scanFavoriteRow(rows)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	return favorites, nil
}

// GetByUserAndQuestion 根据用户 ID 和题目 ID 获取收藏
func (dao *FavoriteQuestionDAO) GetByUserAndQuestion(userID int64, questionID int64) (*models.FavoriteQuestion, error) {
	query := `SELECT id, question_id, category, user_id FROM favorite_questions WHERE user_id = ? AND question_id = ?`

	row := dao.QueryRow(query, userID, questionID)
	favorite, err := dao.scanFavorite(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return favorite, nil
}

// GetByUserAndCategory 根据用户 ID 和类别获取收藏
func (dao *FavoriteQuestionDAO) GetByUserAndCategory(userID int64, category string) ([]*models.FavoriteQuestion, error) {
	query := `SELECT id, question_id, category, user_id FROM favorite_questions WHERE user_id = ? AND category = ?`

	rows, err := dao.ExecuteQuery(query, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*models.FavoriteQuestion
	for rows.Next() {
		favorite, err := dao.scanFavoriteRow(rows)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户类别收藏成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(favorites),
	})

	return favorites, nil
}

// Delete 删除收藏
func (dao *FavoriteQuestionDAO) Delete(id int64) error {
	query := `DELETE FROM favorite_questions WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, id)
	if err != nil {
		return err
	}

	utils.Info("FavoriteQuestionDAO", "删除收藏成功", map[string]interface{}{
		"favorite_id": id,
	})

	return nil
}

// DeleteByUserAndQuestion 根据用户 ID 和题目 ID 删除收藏
func (dao *FavoriteQuestionDAO) DeleteByUserAndQuestion(userID int64, questionID int64) error {
	query := `DELETE FROM favorite_questions WHERE user_id = ? AND question_id = ?`

	_, err := dao.ExecuteUpdate(query, userID, questionID)
	if err != nil {
		return err
	}

	utils.Debug("FavoriteQuestionDAO", "删除用户收藏成功", map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	})

	return nil
}

// GetCountByUser 获取用户收藏数量
func (dao *FavoriteQuestionDAO) GetCountByUser(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM favorite_questions WHERE user_id = ?`

	row := dao.QueryRow(query, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户收藏数量成功", map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})

	return count, nil
}

// GetCountByUserAndCategory 获取用户指定类别的收藏数量
func (dao *FavoriteQuestionDAO) GetCountByUserAndCategory(userID int64, category string) (int, error) {
	query := `SELECT COUNT(*) FROM favorite_questions WHERE user_id = ? AND category = ?`

	row := dao.QueryRow(query, userID, category)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("FavoriteQuestionDAO", "获取用户类别收藏数量成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    count,
	})

	return count, nil
}

// GetCount 获取收藏总数
func (dao *FavoriteQuestionDAO) GetCount() (int, error) {
	query := `SELECT COUNT(*) FROM favorite_questions`

	row := dao.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	utils.Debug("FavoriteQuestionDAO", "获取收藏总数成功", map[string]interface{}{
		"count": count,
	})

	return count, nil
}

// DeleteByUserAndCategory 根据用户 ID 和类别删除收藏
func (dao *FavoriteQuestionDAO) DeleteByUserAndCategory(userID int64, category string) error {
	query := `DELETE FROM favorite_questions WHERE user_id = ? AND category = ?`

	_, err := dao.ExecuteUpdate(query, userID, category)
	if err != nil {
		return err
	}

	utils.Info("FavoriteQuestionDAO", "删除用户类别收藏成功", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	return nil
}

// ClearByUser 清空用户的收藏
func (dao *FavoriteQuestionDAO) ClearByUser(userID int64) error {
	query := `DELETE FROM favorite_questions WHERE user_id = ?`

	_, err := dao.ExecuteUpdate(query, userID)
	if err != nil {
		return err
	}

	utils.Info("FavoriteQuestionDAO", "清空用户收藏成功", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// scanFavorite 扫描单行数据
func (dao *FavoriteQuestionDAO) scanFavorite(row Scanner) (*models.FavoriteQuestion, error) {
	favorite := &models.FavoriteQuestion{}
	err := row.Scan(&favorite.ID, &favorite.QuestionID, &favorite.Category, &favorite.UserID)
	if err != nil {
		return nil, err
	}
	return favorite, nil
}

// scanFavoriteRow 扫描多行数据中的一行
func (dao *FavoriteQuestionDAO) scanFavoriteRow(rows *sql.Rows) (*models.FavoriteQuestion, error) {
	favorite := &models.FavoriteQuestion{}
	err := rows.Scan(&favorite.ID, &favorite.QuestionID, &favorite.Category, &favorite.UserID)
	if err != nil {
		return nil, err
	}
	return favorite, nil
}
