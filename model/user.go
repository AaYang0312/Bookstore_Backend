package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"` // 用户名
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"` // 头像
	IsAdmin   string    `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}
