package dao

import (
	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"
	"time"

	"gorm.io/gorm"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	*BaseDAO
}

// NewUserDAO 创建 UserDAO 实例
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		BaseDAO: NewBaseDAO(db, "users"),
	}
}

// Create 创建用户
func (dao *UserDAO) Create(user *models.User) (int64, error) {
	result := dao.db.Create(user)
	if result.Error != nil {
		return 0, result.Error
	}

	utils.Info("UserDAO", "创建用户成功", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	return user.ID, nil
}

// GetByID 根据 ID 获取用户
func (dao *UserDAO) GetByID(id int64) (*models.User, error) {
	user := &models.User{}
	result := dao.db.First(user, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	result := dao.db.Where("username = ?", username).First(user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return user, nil
}

// GetByIDCard 根据身份证号获取用户
func (dao *UserDAO) GetByIDCard(idCard string) (*models.User, error) {
	user := &models.User{}
	result := dao.db.Where("id_card = ?", idCard).First(user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return user, nil
}

// UpdateLastLogin 更新最后登录时间
func (dao *UserDAO) UpdateLastLogin(id int64, lastLogin time.Time) error {
	result := dao.db.Model(&models.User{}).Where("id = ?", id).Update("last_login", lastLogin)
	if result.Error != nil {
		return result.Error
	}

	utils.Debug("UserDAO", "更新用户最后登录时间成功", map[string]interface{}{
		"user_id":    id,
		"last_login": lastLogin,
	})

	return nil
}

// GetAll 获取所有用户
func (dao *UserDAO) GetAll() ([]*models.User, error) {
	var users []*models.User
	result := dao.db.Order("id").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
