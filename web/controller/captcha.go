package controller

import (
	"bookstore-manager/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 生成验证码
func GenerateCaptcha(ctx *gin.Context) {
	// 生成验证码
	captchaSvc := service.NewCaptchaService()
	// 生成响应
	response, err := captchaSvc.GenerateCaptcha()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "生成验证码失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "验证码生成成功",
		"data":    response,
	})
}
