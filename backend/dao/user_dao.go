package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	*BaseDAO
}

// NewUserDAO 创建 UserDAO 实例
func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{
		BaseDAO: NewBaseDAO(db, "users"),
	}
}

// Create 创建用户
func (dao *UserDAO) Create(user *models.User) (int64, error) {
	query := `INSERT INTO users (username, id_card, last_login) VALUES (?, ?, ?)`

	result, err := dao.ExecuteUpdate(query, user.Username, user.IDCard, user.LastLogin)
	if err != nil {
		return 0, err
	}

	id, err := dao.GetLastInsertID(result)
	if err != nil {
		return 0, err
	}

	user.ID = id
	utils.Info("UserDAO", "创建用户成功", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	return user.ID, nil
}

// GetByID 根据 ID 获取用户
func (dao *UserDAO) GetByID(id int64) (*models.User, error) {
	query := `SELECT id, username, id_card, last_login FROM users WHERE id = ?`

	row := dao.QueryRow(query, id)
	user, err := dao.scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, id_card, last_login FROM users WHERE username = ?`

	row := dao.QueryRow(query, username)
	user, err := dao.scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByIDCard 根据身份证号获取用户
func (dao *UserDAO) GetByIDCard(idCard string) (*models.User, error) {
	query := `SELECT id, username, id_card, last_login FROM users WHERE id_card = ?`

	row := dao.QueryRow(query, idCard)
	user, err := dao.scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateLastLogin 更新最后登录时间
func (dao *UserDAO) UpdateLastLogin(id int64, lastLogin time.Time) error {
	query := `UPDATE users SET last_login = ? WHERE id = ?`

	_, err := dao.ExecuteUpdate(query, lastLogin, id)
	if err != nil {
		return err
	}

	utils.Debug("UserDAO", "更新用户最后登录时间成功", map[string]interface{}{
		"user_id":    id,
		"last_login": lastLogin,
	})

	return nil
}

// GetAll 获取所有用户
func (dao *UserDAO) GetAll() ([]*models.User, error) {
	query := `SELECT id, username, id_card, last_login FROM users ORDER BY id`

	rows, err := dao.ExecuteQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := dao.scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// scanUser 扫描单行数据
func (dao *UserDAO) scanUser(row Scanner) (*models.User, error) {
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.IDCard, &user.LastLogin)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// scanUserRow 扫描多行数据中的一行
func (dao *UserDAO) scanUserRow(rows *sql.Rows) (*models.User, error) {
	user := &models.User{}
	err := rows.Scan(&user.ID, &user.Username, &user.IDCard, &user.LastLogin)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Scanner 扫描接口（兼容 sql.Row 和 sql.Rows）
type Scanner interface {
	Scan(dest ...interface{}) error
}
