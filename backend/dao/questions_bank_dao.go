package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// QuestionsBankDAO 题库数据访问对象
type QuestionsBankDAO struct {
	*BaseDAO
	columns []string
}

// NewQuestionsBankDAO 创建 QuestionsBankDAO 实例
func NewQuestionsBankDAO(db *gorm.DB) *QuestionsBankDAO {
	return &QuestionsBankDAO{
		BaseDAO: NewBaseDAO(db, "questions"),
		columns: []string{"id", "J", "P", "I", "Q", "T", "A", "B", "C", "D", "F", "LA", "LB", "LC", "type", "user_id"},
	}
}

// GetTotalRecords 获取符合条件的总记录数
func (dao *QuestionsBankDAO) GetTotalRecords(searchQuery string, filterLA, filterLB, filterLC bool) (int64, error) {
	query := dao.db.Model(&models.Question{})

	// 添加搜索条件
	if searchQuery != "" {
		searchTerm := "%" + searchQuery + "%"
		query = query.Where("J LIKE ? OR P LIKE ? OR I LIKE ? OR Q LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// 添加分类筛选条件
	categoryConditions := []string{}
	if filterLA {
		categoryConditions = append(categoryConditions, "LA = 1")
	}
	if filterLB {
		categoryConditions = append(categoryConditions, "LB = 1")
	}
	if filterLC {
		categoryConditions = append(categoryConditions, "LC = 1")
	}

	if len(categoryConditions) > 0 {
		query = query.Where(strings.Join(categoryConditions, " OR "))
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// PageResult 分页结果
type PageResult struct {
	Data  []*models.Question
	Total int64
}

// GetPageData 获取分页数据
func (dao *QuestionsBankDAO) GetPageData(pageNum, pageSize int, searchQuery string, filterLA, filterLB, filterLC bool) (*PageResult, error) {
	query := dao.db.Model(&models.Question{})

	// 添加搜索条件
	if searchQuery != "" {
		searchTerm := "%" + searchQuery + "%"
		query = query.Where("J LIKE ? OR P LIKE ? OR I LIKE ? OR Q LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// 添加分类筛选条件
	categoryConditions := []string{}
	if filterLA {
		categoryConditions = append(categoryConditions, "LA = 1")
	}
	if filterLB {
		categoryConditions = append(categoryConditions, "LB = 1")
	}
	if filterLC {
		categoryConditions = append(categoryConditions, "LC = 1")
	}

	if len(categoryConditions) > 0 {
		query = query.Where(strings.Join(categoryConditions, " OR "))
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	var questions []*models.Question
	offset := (pageNum - 1) * pageSize
	err := query.Order("id").Offset(offset).Limit(pageSize).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return &PageResult{
		Data:  questions,
		Total: total,
	}, nil
}

// UpdateQuestion 更新题目数据
func (dao *QuestionsBankDAO) UpdateQuestion(questionID int64, updatedData map[string]interface{}) error {
	if len(updatedData) == 0 {
		return fmt.Errorf("更新数据不能为空")
	}

	utils.Info("QuestionsBankDAO", "更新题目", map[string]interface{}{
		"question_id":  questionID,
		"updated_data": updatedData,
	})

	return dao.db.Model(&models.Question{}).Where("id = ?", questionID).Updates(updatedData).Error
}

// GetQuestionByID 根据 ID 获取题目
func (dao *QuestionsBankDAO) GetQuestionByID(questionID int64) (*models.Question, error) {
	question := &models.Question{}
	result := dao.db.First(question, questionID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return question, nil
}

// DeleteQuestion 删除题目
func (dao *QuestionsBankDAO) DeleteQuestion(questionID int64) error {
	result := dao.db.Delete(&models.Question{}, questionID)
	if result.Error != nil {
		return result.Error
	}

	utils.Info("QuestionsBankDAO", "删除题目成功", map[string]interface{}{
		"question_id": questionID,
	})

	return nil
}

// ResetTable 重置题库表
func (dao *QuestionsBankDAO) ResetTable() error {
	utils.Info("QuestionsBankDAO", "重置题库表", nil)

	err := dao.db.Migrator().DropTable(&models.Question{})
	if err != nil {
		return err
	}

	err = dao.db.AutoMigrate(&models.Question{})
	if err != nil {
		return err
	}

	utils.Info("QuestionsBankDAO", "题库表重置完成", nil)

	return nil
}

// GetAllQuestions 获取所有题目
func (dao *QuestionsBankDAO) GetAllQuestions() ([]*models.Question, error) {
	var questions []*models.Question
	err := dao.db.Order("id").Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}
