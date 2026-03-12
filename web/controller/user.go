package controller

import (
	"bookstore-manager/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 给 controller 对象做归属
type UserController struct {
	UserService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		UserService: service.NewUserService(),
	}
}

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

func (u *UserController) UserRegister(ctx *gin.Context) {
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
	//svc := service.NewUserService()
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
	err = u.UserService.UserRegister(req.Username, req.Password, req.Phone, req.Email)
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
func (u *UserController) UserLogin(ctx *gin.Context) {
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
	//userSvc := service.NewUserService()
	response, err := u.UserService.UserLogin(req.Username, req.Password)
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

func (u *UserController) GetUserProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
	}
	// 调用服务层的获取用户信息
	user, err := u.UserService.GetUserByID(userID.(int))
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	response := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"phone":      user.Phone,
		"avatar":     user.Avatar,
		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"data":    response,
		"message": "获取用户信息成功",
	})
}
