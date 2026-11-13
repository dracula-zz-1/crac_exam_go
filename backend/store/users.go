package store

import (
	"database/sql"
	"fmt"

	"crac_exam_v3/backend/models"
)

// EnsureUser 按身份证号查找用户，不存在则创建，并更新最后登录时间。
func EnsureUser(db *sql.DB, username, idCard string) (*models.User, error) {
	now := nowMS()
	row := db.QueryRow(`
		INSERT INTO users (username, id_card, last_login) VALUES (?, ?, ?)
		ON CONFLICT(id_card) DO UPDATE SET username = excluded.username, last_login = excluded.last_login
		RETURNING id, username, id_card, last_login`,
		username, idCard, now)

	var u models.User
	var lastLogin int64
	if err := row.Scan(&u.ID, &u.Username, &u.IDCard, &lastLogin); err != nil {
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}
	u.LastLogin = toTime(lastLogin)
	return &u, nil
}

func GetUser(db *sql.DB, id int64) (*models.User, error) {
	row := db.QueryRow(`SELECT id, username, id_card, last_login FROM users WHERE id = ?`, id)
	var u models.User
	var lastLogin int64
	if err := row.Scan(&u.ID, &u.Username, &u.IDCard, &lastLogin); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.LastLogin = toTime(lastLogin)
	return &u, nil
}

func CountUsers(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
