package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"fmt"
	"strings"
)

// QuestionsBankDAO 题库数据访问对象
type QuestionsBankDAO struct {
	*BaseDAO
	columns []string
}

// NewQuestionsBankDAO 创建 QuestionsBankDAO 实例
func NewQuestionsBankDAO(db *sql.DB) *QuestionsBankDAO {
	return &QuestionsBankDAO{
		BaseDAO: NewBaseDAO(db, "questions"),
		columns: []string{"id", "J", "P", "I", "Q", "T", "A", "B", "C", "D", "F", "LA", "LB", "LC", "type", "user_id"},
	}
}

// GetTotalRecords 获取符合条件的总记录数
// Python 原版：get_total_records(search_query, filter_la, filter_lb, filter_lc) -> int
func (dao *QuestionsBankDAO) GetTotalRecords(searchQuery string, filterLA, filterLB, filterLC bool) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", dao.tableName)
	params := []interface{}{}
	conditions := []string{}

	// 添加搜索条件
	if searchQuery != "" {
		searchTerm := "%" + searchQuery + "%"
		searchConditions := []string{
			"J LIKE ?", "P LIKE ?", "I LIKE ?", "Q LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		for i := 0; i < 4; i++ {
			params = append(params, searchTerm)
		}
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
		// ABC 类筛选是 OR 关系
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(categoryConditions, " OR ")))
	}

	// 组合条件
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	utils.Debug("QuestionsBankDAO", "获取总记录数", map[string]interface{}{
		"query":  query,
		"params": params,
	})

	row := dao.QueryRow(query, params...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// PageResult 分页结果
type PageResult struct {
	Data  []*models.Question
	Total int
}

// GetPageData 获取分页数据
func (dao *QuestionsBankDAO) GetPageData(pageNum, pageSize int, searchQuery string, filterLA, filterLB, filterLC bool) (*PageResult, error) {
	offset := (pageNum - 1) * pageSize

	columns := strings.Join(dao.columns, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s", columns, dao.tableName)
	params := []interface{}{}
	conditions := []string{}

	// 构建 WHERE 子句
	if searchQuery != "" {
		searchTerm := "%" + searchQuery + "%"
		searchConditions := []string{
			"J LIKE ?", "P LIKE ?", "I LIKE ?", "Q LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		for i := 0; i < 4; i++ {
			params = append(params, searchTerm)
		}
		utils.Info("QuestionsBankDAO", "搜索条件", map[string]interface{}{
			"search_query": searchQuery,
			"search_term":  searchTerm,
			"fields":       []string{"J", "P", "I", "Q"},
		})
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
		// ABC 类筛选是 OR 关系
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(categoryConditions, " OR ")))
	}

	// 组合条件
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// 计算总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", dao.tableName)
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := dao.QueryRow(countQuery, params...).Scan(&total)
	if err != nil {
		utils.Error("QuestionsBankDAO", "计算总数失败", err, nil)
		return nil, err
	}

	// 添加分页
	query += " LIMIT ? OFFSET ?"
	params = append(params, pageSize, offset)

	utils.Info("QuestionsBankDAO", "获取分页数据", map[string]interface{}{
		"query":     query,
		"params":    params,
		"page":      pageNum,
		"page_size": pageSize,
		"offset":    offset,
		"total":     total,
	})

	rows, err := dao.ExecuteQuery(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question, err := dao.scanQuestionRow(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	return &PageResult{
		Data:  questions,
		Total: total,
	}, nil
}

// UpdateQuestion 更新题目数据
// Python 原版：update_question(question_id, updated_data) -> bool
func (dao *QuestionsBankDAO) UpdateQuestion(questionID int64, updatedData map[string]interface{}) error {
	if len(updatedData) == 0 {
		return fmt.Errorf("更新数据不能为空")
	}

	// 构建 SET 子句
	setClauses := []string{}
	params := []interface{}{}

	for key, value := range updatedData {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		params = append(params, value)
	}

	setClause := strings.Join(setClauses, ", ")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", dao.tableName, setClause)
	params = append(params, questionID)

	utils.Info("QuestionsBankDAO", "更新题目", map[string]interface{}{
		"question_id":  questionID,
		"updated_data": updatedData,
		"query":        query,
		"params":       params,
	})

	_, err := dao.ExecuteUpdate(query, params...)
	if err != nil {
		return err
	}

	return nil
}

// GetQuestionByID 根据 ID 获取题目
// Python 原版：get_question_by_id(question_id) -> Optional[Dict]
func (dao *QuestionsBankDAO) GetQuestionByID(questionID int64) (*models.Question, error) {
	columns := strings.Join(dao.columns, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", columns, dao.tableName)

	utils.Debug("QuestionsBankDAO", "根据 ID 获取题目", map[string]interface{}{
		"question_id": questionID,
	})

	row := dao.QueryRow(query, questionID)
	question, err := dao.scanQuestion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return question, nil
}

// DeleteQuestion 删除题目
func (dao *QuestionsBankDAO) DeleteQuestion(questionID int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", dao.tableName)

	_, err := dao.ExecuteUpdate(query, questionID)
	if err != nil {
		return err
	}

	utils.Info("QuestionsBankDAO", "删除题目成功", map[string]interface{}{
		"question_id": questionID,
	})

	return nil
}

// ResetTable 重置题库表（删除并重新创建）
// Python 原版：reset_questions_table()
func (dao *QuestionsBankDAO) ResetTable() error {
	utils.Info("QuestionsBankDAO", "重置题库表", nil)

	// 删除原题库表
	dropQuery := "DROP TABLE IF EXISTS questions"
	_, err := dao.ExecuteUpdate(dropQuery)
	if err != nil {
		utils.Error("QuestionsBankDAO", "删除原题库表失败", err, nil)
		return err
	}
	utils.Debug("QuestionsBankDAO", "已删除原题库表", nil)

	// 重新创建题库表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS questions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		J TEXT NOT NULL,
		P TEXT NOT NULL,
		I TEXT,
		Q TEXT NOT NULL,
		T TEXT NOT NULL,
		A TEXT NOT NULL,
		B TEXT NOT NULL,
		C TEXT NOT NULL,
		D TEXT NOT NULL,
		F TEXT,
		LA INTEGER NOT NULL DEFAULT 0,
		LB INTEGER NOT NULL DEFAULT 0,
		LC INTEGER NOT NULL DEFAULT 0,
		type INTEGER NOT NULL DEFAULT 1,
		user_id INTEGER DEFAULT 0
	)`

	_, err = dao.ExecuteUpdate(createTableSQL)
	if err != nil {
		utils.Error("QuestionsBankDAO", "重新创建题库表失败", err, nil)
		return err
	}

	utils.Info("QuestionsBankDAO", "题库表重置完成", nil)

	return nil
}

// scanQuestion 扫描单行数据
func (dao *QuestionsBankDAO) scanQuestion(row Scanner) (*models.Question, error) {
	question := &models.Question{}
	err := row.Scan(
		&question.ID, &question.J, &question.P, &question.I, &question.Q, &question.T,
		&question.A, &question.B, &question.C, &question.D, &question.F,
		&question.LA, &question.LB, &question.LC, &question.Type, &question.UserID)
	if err != nil {
		return nil, err
	}
	return question, nil
}

// scanQuestionRow 扫描多行数据中的一行
func (dao *QuestionsBankDAO) scanQuestionRow(rows *sql.Rows) (*models.Question, error) {
	question := &models.Question{}
	err := rows.Scan(
		&question.ID, &question.J, &question.P, &question.I, &question.Q, &question.T,
		&question.A, &question.B, &question.C, &question.D, &question.F,
		&question.LA, &question.LB, &question.LC, &question.Type, &question.UserID)
	if err != nil {
		return nil, err
	}
	return question, nil
}

// GetAllQuestions 获取所有题目（用于导出）
// Python 原版：export_questions_to_json 中的 SELECT * FROM questions
func (dao *QuestionsBankDAO) GetAllQuestions() ([]*models.Question, error) {
	columns := []string{"id", "J", "P", "I", "Q", "T", "A", "B", "C", "D", "F", "LA", "LB", "LC", "type", "user_id"}
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY id", strings.Join(columns, ", "), dao.tableName)

	utils.Debug("QuestionsBankDAO", "获取所有题目", map[string]interface{}{
		"query": query,
	})

	rows, err := dao.ExecuteQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question := &models.Question{}
		err := rows.Scan(
			&question.ID, &question.J, &question.P, &question.I, &question.Q, &question.T,
			&question.A, &question.B, &question.C, &question.D, &question.F,
			&question.LA, &question.LB, &question.LC, &question.Type, &question.UserID)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
