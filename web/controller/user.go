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
	CaptchaID       string `json:"captcha_id"`
	CaptchaValue    string `json:"captcha_value"`
}
type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	CaptchaID    string `json:"captcha_id"`
	CaptchaValue string `json:"captcha_value"`
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
	captSvc := service.NewCaptchaService()
	if !captSvc.VerifyCaptcha(req.CaptchaID, req.CaptchaValue) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码错误",
		})
	}

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
	// JWT: 用于验证用户身份的一段 Hash 值，服务端根据 JWT 获取对应用户信息
	// 1. 验证图片验证码
	// 2. 校验用户信息，是否有这个用户，校验密码是否正确
	// 3. 返回 JWT 信息给用户，后面发送就知道是哪个用户
	var req LoginRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数绑定错误",
			"error":   err.Error(),
		})
	}
	captSvc := service.NewCaptchaService()
	if !captSvc.VerifyCaptcha(req.CaptchaID, req.CaptchaValue) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码错误",
		})
	}
	userSvc := service.NewUserService()
	response, err := userSvc.UserLogin(req.Username, req.Password)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"data":    response,
		"message": "登陆成功",
	})
}
