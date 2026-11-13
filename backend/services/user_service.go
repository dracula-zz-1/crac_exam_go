package services

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode/utf8"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) ServiceName() string { return "UserService" }

// Login 本地档案登录：身份证号首次使用即创建档案。
func (s *UserService) Login(username, idCard string) (*models.User, error) {
	username = strings.TrimSpace(username)
	idCard = strings.TrimSpace(idCard)

	if n := utf8.RuneCountInString(username); n == 0 || n > 32 {
		return nil, fmt.Errorf("用户名长度需在 1-32 个字符之间")
	}
	if n := utf8.RuneCountInString(idCard); n < 3 || n > 32 {
		return nil, fmt.Errorf("证件号长度需在 3-32 个字符之间")
	}

	user, err := store.EnsureUser(s.db, username, idCard)
	if err != nil {
		utils.Error("UserService", "登录失败", err)
		return nil, fmt.Errorf("登录失败，请重试")
	}
	utils.Info("UserService", "用户登录", "user_id", user.ID)
	return user, nil
}

func (s *UserService) GetUser(id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的用户")
	}
	return store.GetUser(s.db, id)
}
