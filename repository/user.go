package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
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
