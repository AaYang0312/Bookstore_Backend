package controller

import (
	"bookstore-manager/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoriteController struct {
	FavoriteService *service.FavoriteService
}

func NewFavoriteController() *FavoriteController {
	return &FavoriteController{
		FavoriteService: service.NewFavoriteService(),
	}
}
func getUserID(ctx *gin.Context) int {
	userID, exist := ctx.Get("userID")
	if !exist {
		return 0
	}
	return userID.(int)
}
func (f *FavoriteController) AddFavorite(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	bookID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的书籍ID",
		})
		return
	}
	err = f.FavoriteService.AddFavorite(userID, bookID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "添加收藏失败",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "添加收藏成功",
	})
}
func (f *FavoriteController) DelFavorite(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	bookID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的书籍ID",
		})
		return
	}
	err = f.FavoriteService.DelFavorite(userID, bookID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "移除收藏失败",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "移除收藏成功",
	})
}
