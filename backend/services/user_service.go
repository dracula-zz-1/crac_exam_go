package services

import (
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"database/sql"
	"time"
)

// UserService 用户服务
type UserService struct {
	userDAO *dao.UserDAO
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Success  bool         `json:"success"`
	Message  string       `json:"message,omitempty"`
	UserInfo *models.User `json:"user_info,omitempty"`
}

// NewUserService 创建 UserService 实例
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		userDAO: dao.NewUserDAO(db),
	}
}

// Login 用户登录
// Python 原版：login(username, id_card) -> Dict[str, Any]
// 如果身份证号不存在，会自动创建新用户
func (s *UserService) Login(username string, idCard string) (*UserLoginResponse, error) {
	utils.Info("UserService", "用户登录", map[string]interface{}{
		"username": username,
		"id_card":  idCard,
	})

	// 通过身份证号查找用户
	user, err := s.userDAO.GetByIDCard(idCard)
	if err != nil {
		utils.Error("UserService", "查询用户失败", err, map[string]interface{}{
			"id_card": idCard,
		})
		return &UserLoginResponse{
			Success: false,
			Message: "登录失败，请重试",
		}, nil
	}

	// 如果用户不存在，自动创建新用户（Python 原版行为）
	if user == nil {
		utils.Info("UserService", "用户不存在，创建新用户", map[string]interface{}{
			"username": username,
			"id_card":  idCard,
		})

		// 创建新用户
		currentTime := time.Now()
		newUser := &models.User{
			Username:  username,
			IDCard:    idCard,
			LastLogin: currentTime,
		}

		userID, err := s.userDAO.Create(newUser)
		if err != nil {
			utils.Error("UserService", "创建用户失败", err, map[string]interface{}{
				"username": username,
				"id_card":  idCard,
			})
			return &UserLoginResponse{
				Success: false,
				Message: "创建用户失败",
			}, nil
		}

		// 获取刚创建的用户
		user, err = s.userDAO.GetByID(userID)
		if err != nil || user == nil {
			utils.Error("UserService", "获取新用户失败", err, nil)
			return &UserLoginResponse{
				Success: false,
				Message: "获取用户信息失败",
			}, nil
		}

		utils.Info("UserService", "新用户创建成功", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
		})
	} else {
		// 用户存在，更新最后登录时间
		utils.Info("UserService", "用户存在，更新最后登录时间", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
		})

		user.LastLogin = time.Now()
		err := s.userDAO.UpdateLastLogin(user.ID, user.LastLogin)
		if err != nil {
			utils.Error("UserService", "更新最后登录时间失败", err, map[string]interface{}{
				"user_id": user.ID,
			})
		}
	}

	// 返回成功响应
	return &UserLoginResponse{
		Success: true,
		UserInfo: &models.User{
			ID:        user.ID,
			Username:  user.Username,
			IDCard:    user.IDCard,
			LastLogin: user.LastLogin,
		},
	}, nil
}

// GetByID 根据用户 ID 获取用户信息
// Python 原版：get_user_by_id(user_id) -> Optional[User]
func (s *UserService) GetByID(userID int64) (*models.User, error) {
	return s.userDAO.GetByID(userID)
}

// GetByUsername 根据用户名获取用户信息
// Python 原版：get_user_by_username(username) -> Optional[User]
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return s.userDAO.GetByUsername(username)
}

// GetByIDCard 根据身份证号获取用户信息
// Python 原版：get_user_by_id_card(id_card) -> Optional[User]
func (s *UserService) GetByIDCard(idCard string) (*models.User, error) {
	return s.userDAO.GetByIDCard(idCard)
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userDAO.GetAll()
}
