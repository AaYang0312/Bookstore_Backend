package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"

	"github.com/cloudwego/base64x"
)

type UserService struct {
	UserDB *repository.UserDAO
}

// service --> repository --> 调用 db 方法(操作 model 里的模型)
func NewUserService() *UserService {
	return &UserService{
		UserDB: repository.NewUserDAO(),
	}
}

// 服务层面，service层面方法
func (u *UserService) UserRegister(username, password, phone, email string) error {
	// 1.检查用户名，邮箱，手机号唯一
	exists, err := u.UserDB.CheckUserExists(username, phone, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户名，邮箱或手机号已存在")
	}

	// 2.密码加密（base64编码）
	encodedPassword := u.encodePassword(password)

	// 3.创建用户
	err = u.createUser(username, encodedPassword, phone, email)
	if err != nil {
		return err
	}
	return nil
}

// 小写开头，对外不可见
func (u *UserService) encodePassword(password string) string {
	return base64x.StdEncoding.EncodeToString([]byte(password))
}
func (u *UserService) createUser(username, passwordHash, phone, email string) error {
	user := &model.User{
		Username: username,
		Password: passwordHash,
		Phone:    phone,
		Email:    email,
	}
	return u.UserDB.CreateUser(user)
}
