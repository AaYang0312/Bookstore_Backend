package controller

import (
	"bookstore-manager/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 注册的请求参数
type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func UserRegister(ctx *gin.Context) {
	var req RegisterRequest
	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    -1,
			"message": "参数绑定失败",
			"error":   err,
		})
		return
	}
	svc := service.NewUserService()
	// 验证密码两次是否一致
	if req.Password != req.ConfirmPassword {
		// 用户请求错误
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "两次密码不一致",
		})
		return
	}
	err = svc.UserRegister(req.Username, req.Password, req.Phone, req.Email)
	if err != nil {
		// 服务器内部错误
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	// 成功
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "注册成功",
	})
}
func UserLogin(ctx *gin.Context) {

}
