package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"time"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

// AdminUser 明确排除密码字段，避免管理端列表泄露凭据。
type AdminUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserDAO() *UserDAO {
	return &UserDAO{
		db: global.GetDB(),
	}
}

func (u *UserDAO) CreateUser(user *model.User) error {
	return u.db.Debug().Create(user).Error
}

func (u *UserDAO) CheckUserExists(username, phone, email string) (bool, error) {
	var total int64

	// 使用 OR 条件一次性检查三个字段
	err := u.db.Model(&model.User{}).
		Where("username = ? OR phone = ? OR email = ?", username, phone, email).
		Count(&total).Error

	if err != nil {
		return false, err
	}

	return total > 0, nil
}

func (u *UserDAO) CheckLoginUserExists(username string) (*model.User, error) {
	var user model.User

	err := u.db.Model(&model.User{}).Debug().
		Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserDAO) GetUserByID(userID int) (*model.User, error) {
	var user model.User

	err := u.db.Debug().First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserDAO) UpdateUser(user *model.User) error { return u.db.Debug().Save(user).Error }

func (u *UserDAO) ChangePassword(user *model.User) error {
	return u.db.Debug().Updates(user).Error
}

func (u *UserDAO) GetAdminUsers(keyword string, page, pageSize int) ([]*AdminUser, int64, error) {
	var users []*AdminUser
	var total int64
	query := u.db.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Select("id, username, email, phone, avatar, is_admin, created_at, updated_at").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Scan(&users).Error
	return users, total, err
}

func (u *UserDAO) UpdateAdminRole(id int, isAdmin bool) error {
	result := u.db.Model(&model.User{}).Where("id = ?", id).Update("is_admin", isAdmin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := u.db.Model(&model.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (u *UserDAO) GetAdminUserByID(id int) (*AdminUser, error) {
	var user AdminUser
	result := u.db.Model(&model.User{}).
		Select("id, username, email, phone, avatar, is_admin, created_at, updated_at").
		Where("id = ?", id).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, nil
}
