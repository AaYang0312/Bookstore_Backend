package service

import (
	"bookstore-manager/jwt"
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"encoding/base64"
	"errors"
)

type UserService struct {
	UserDB *repository.UserDAO
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpireIn     int64     `json:"expire_in"`
	Userinfo     *UserInfo `json:"user_info"`
}
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
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

func (u *UserService) UserLogin(username, password string) (*LoginResponse, error) {
	// 查询有没有这个用户
	user, err := u.UserDB.CheckLoginUserExists(username)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	// 如果用户存在，验证密码
	if !u.verifyPassword(user.Password, password) {
		return nil, errors.New("密码错误")
	}
	// 返回 JWT
	token, err_ := jwt.GenerateTokenPair(uint(user.ID), username)
	if err_ != nil {
		return nil, errors.New("生成token失败")
	}
	response := &LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpireIn:     token.ExpiresIn,
		Userinfo: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
		},
	}
	return response, nil
}

// 小写开头，对外不可见
func (u *UserService) encodePassword(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
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

func (u *UserService) verifyPassword(storedPassword, password string) bool {
	return u.encodePassword(password) == storedPassword
}

func (u *UserService) GetUserByID(userID int) (*model.User, error) {
	user, err := u.UserDB.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

func (u *UserService) UpdateUserInfo(user *model.User) error {
	// 用户是否存在
	existingUser, err := u.UserDB.GetUserByID(user.ID)
	if err != nil {
		return errors.New("用户不存在")
	}
	existingUser.Phone = user.Phone
	existingUser.Email = user.Email
	existingUser.Username = user.Username
	existingUser.Avatar = user.Avatar

	// 调用 DAO 层更新信息
	return u.UserDB.UpdateUser(existingUser)
}

func (u *UserService) ChangePassword(userID int, oldPassword, newPassword string) error {
	// 获取对应用户
	user, err := u.UserDB.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	// 验证旧密码
	if !u.verifyPassword(user.Password, oldPassword) {
		return errors.New("密码错误")
	}
	// 操作数据库修改密码
	user.Password = u.encodePassword(newPassword)
	return u.UserDB.ChangePassword(user)
}
