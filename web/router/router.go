package router

import (
	"bookstore-manager/web/controller"
	"bookstore-manager/web/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	//r.GET("/test", func(context *gin.Context) {
	//	context.JSON(http.StatusOK, gin.H{
	//		"data":  "hello",
	//		"error": "none",
	//	})
	//})
	//return r

	// 添加CORS中间件，解决跨域问题和OPTIONS预检请求
	r.Use(func(c *gin.Context) {
		// 1. 设置允许的源
		c.Header("Access-Control-Allow-Origin", "*") // 允许所有来源

		// 2. 设置允许的 HTTP 方法
		c.Header("Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS")

		// 3. 设置允许的请求头
		c.Header("Access-Control-Allow-Headers",
			"Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Requested-With, Access-Control-Request-Method, Access-Control-Request-Headers")

		// 4. 设置暴露的头（前端可以访问这些头）
		c.Header("Access-Control-Expose-Headers",
			"Content-Length")

		// 5. 允许携带凭证
		c.Header("Access-Control-Allow-Credentials", "true")

		// 6. 设置预检请求缓存时间（秒）
		c.Header("Access-Control-Max-Age", "43200") // 12 小时

		// 7. 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			// 直接返回 200 OK，不执行后续 handler
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 8. 继续执行后续处理
		c.Next()
	})
	userController := controller.NewUserController()
	captchaController := controller.NewCaptchaController()
	bookController := controller.NewBookController()
	v1 := r.Group("/api/v1")
	{
		user := v1.Group("/user")
		{
			user.POST("/register", userController.UserRegister)
			user.POST("/login", userController.UserLogin)
		}
		auth := user.Group("")
		{
			auth.Use(middleware.JWTAuthMiddleware())
			{
				auth.GET("/profile", userController.GetUserProfile)
				auth.PUT("/profile", userController.UpdateUserProfile)
				auth.PUT("/password", userController.ChangePassword)
			}
		}
		book := v1.Group("/book")
		{
			book.GET("/hot", bookController.GetHotBooks)
			book.GET("/new", bookController.GetNewBooks)
			book.GET("/list", bookController.GetBookList)
			book.GET("/search", bookController.SearchBooks)
			book.GET("/detail/:id", bookController.GetBookDetail)
		}
	}
	captcha := v1.Group("/captcha")
	{
		captcha.GET("/generate", captchaController.GenerateCaptcha)
	}
	return r
}
