package middleware

import (
	"bookstore-manager/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 约定：从请求头获取信息
// header key为Authorization value是Bearer Authorization xxxxx
// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "请求头中缺少Authorization字段",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "Authorization格式错误，应为：Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 解析并验证token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "无效的token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 检查token类型，只允许access token访问API
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "token类型错误，请使用access token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", int(claims.UserID))
		c.Set("username", claims.Username)

		// 继续处理请求
		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（用于可选登录的接口）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果没有token，继续处理请求
			c.Next()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// token格式错误，但不中断请求
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		// 解析并验证token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			// token无效，但不中断请求
			c.Next()
			return
		}

		// 检查token类型
		if claims.TokenType == "access" {
			// 将用户信息存储到上下文中
			c.Set("userID", int(claims.UserID))
			c.Set("username", claims.Username)
			c.Set("authenticated", true)
		}

		// 继续处理请求
		c.Next()
	}
}
