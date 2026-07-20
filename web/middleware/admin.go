package middleware

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminRequired 管理员权限校验中间件
// 必须放在 JWTAuthMiddleware 之后使用，依赖 context 中的 userID
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		var user model.User
		if err := global.GetDB().First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    -1,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    -1,
				"message": "权限不足，需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
