package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"fmt"
	"math/rand"

	"gorm.io/gorm"
)

// QuestionDAO 题目数据访问对象
type QuestionDAO struct {
	*BaseDAO
}

// NewQuestionDAO 创建 QuestionDAO 实例
func NewQuestionDAO(db *gorm.DB) *QuestionDAO {
	return &QuestionDAO{
		BaseDAO: NewBaseDAO(db, "questions"),
	}
}

// Create 创建题目
func (dao *QuestionDAO) Create(question *models.Question) (int64, error) {
	result := dao.db.Create(question)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Info("QuestionDAO", "创建题目成功", map[string]interface{}{
		"question_id": question.ID,
		"type":        question.Type,
	})

	return question.ID, nil
}

// GetByID 根据 ID 获取题目
func (dao *QuestionDAO) GetByID(id int64) (*models.Question, error) {
	question := &models.Question{}
	result := dao.db.First(question, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return question, nil
}

// GetByType 根据题型获取题目
func (dao *QuestionDAO) GetByType(typeValue int) ([]*models.Question, error) {
	var questions []*models.Question
	result := dao.db.Where("type = ?", typeValue).Find(&questions)
	if result.Error != nil {
		return nil, result.Error
	}

	return questions, nil
}

// GetRandomQuestions 随机获取指定数量的题目
func (dao *QuestionDAO) GetRandomQuestions(typeValue int, count int) ([]*models.Question, error) {
	var questions []*models.Question
	result := dao.db.Where("type = ?", typeValue).Find(&questions)
	if result.Error != nil {
		return nil, result.Error
	}

	// Shuffle and take limited number
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	if len(questions) > count {
		questions = questions[:count]
	}

	return questions, nil
}

// Search 搜索题目
func (dao *QuestionDAO) Search(keyword string) ([]*models.Question, error) {
	var questions []*models.Question
	likeKeyword := "%" + keyword + "%"
	result := dao.db.Where("Q LIKE ? OR A LIKE ? OR B LIKE ? OR C LIKE ? OR D LIKE ?",
		likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword).Find(&questions)
	if result.Error != nil {
		return nil, result.Error
	}

	return questions, nil
}

// 类别字段白名单
var validCategoryFields = map[string]bool{
	"LA": true,
	"LB": true,
	"LC": true,
}

// GetByCategory 根据类别获取题目
func (dao *QuestionDAO) GetByCategory(category string) ([]*models.Question, error) {
	var questions []*models.Question
	var result *gorm.DB

	if category == "all" {
		result = dao.db.Find(&questions)
	} else {
		categoryField := dao.getCategoryField(category)
		// 白名单验证，防止 SQL 注入
		if !validCategoryFields[categoryField] {
			return nil, fmt.Errorf("无效的类别字段：%s", categoryField)
		}
		// 使用 GORM map 条件构建，避免 SQL 注入
		result = dao.db.Where(map[string]interface{}{categoryField: 1}).Find(&questions)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return questions, nil
}

// GetRandomByCategoryAndType 根据类别和题型随机获取题目
func (dao *QuestionDAO) GetRandomByCategoryAndType(category string, typeValue int, count int) ([]*models.Question, error) {
	categoryField := dao.getCategoryField(category)
	// 白名单验证，防止 SQL 注入
	if !validCategoryFields[categoryField] {
		return nil, fmt.Errorf("无效的类别字段：%s", categoryField)
	}

	var questions []*models.Question
	// 使用 GORM map 条件构建，避免 SQL 注入
	result := dao.db.Where(map[string]interface{}{categoryField: 1, "type": typeValue}).Find(&questions)
	if result.Error != nil {
		return nil, result.Error
	}

	// Shuffle and take limited number
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	if len(questions) > count {
		questions = questions[:count]
	}

	return questions, nil
}

// Update 更新题目
func (dao *QuestionDAO) Update(question *models.Question) error {
	result := dao.db.Save(question)
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("QuestionDAO", "更新题目成功", map[string]interface{}{
		"question_id": question.ID,
	})

	return nil
}

// Delete 删除题目
func (dao *QuestionDAO) Delete(id int64) error {
	result := dao.db.Delete(&models.Question{}, id)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("QuestionDAO", "删除题目成功", map[string]interface{}{
		"question_id": id,
	})

	return nil
}

// BatchInsert 批量插入题目
func (dao *QuestionDAO) BatchInsert(questions []*models.Question) error {
	if len(questions) == 0 {
		return nil
	}

	result := dao.db.CreateInBatches(questions, 100)
	if result.Error != nil {
		utils.Error("QuestionDAO", "批量插入题目失败", result.Error, map[string]interface{}{
			"count": len(questions),
		})
		return result.Error
	}

	utils.Info("QuestionDAO", "批量插入题目成功", map[string]interface{}{
		"count": len(questions),
	})

	return nil
}

// ResetTable 清空题库表
func (dao *QuestionDAO) ResetTable() error {
	result := dao.db.Exec("DELETE FROM questions")
	if result.Error != nil {
		utils.Error("QuestionDAO", "清空题库表失败", result.Error, nil)
		return result.Error
	}

	utils.Info("QuestionDAO", "清空题库表成功", nil)
	return nil
}

// ClearAll 清空所有题目数据
func (dao *QuestionDAO) ClearAll() error {
	return dao.ResetTable()
}

// GetCount 获取题目总数
func (dao *QuestionDAO) GetCount() (int64, error) {
	var count int64
	result := dao.db.Model(&models.Question{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetCountByCategory 根据类别获取题目总数
func (dao *QuestionDAO) GetCountByCategory(category string) (int64, error) {
	categoryField := dao.getCategoryField(category)
	// 白名单验证，防止 SQL 注入
	if !validCategoryFields[categoryField] {
		return 0, fmt.Errorf("无效的类别字段：%s", categoryField)
	}

	var count int64
	// 使用 GORM map 条件构建，避免 SQL 注入
	result := dao.db.Model(&models.Question{}).Where(map[string]interface{}{categoryField: 1}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// getCategoryField 获取类别字段
func (dao *QuestionDAO) getCategoryField(category string) string {
	switch category {
	case "A":
		return "LA"
	case "B":
		return "LB"
	case "C":
		return "LC"
	default:
		return "LA"
	}
}
