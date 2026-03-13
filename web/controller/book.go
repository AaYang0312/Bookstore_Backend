package controller

import (
	"bookstore-manager/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookController struct {
	BookService *service.BookService
}

func NewBookController() *BookController {
	return &BookController{
		BookService: service.NewBookService(),
	}
}

func (b *BookController) GetHotBooks(ctx *gin.Context) {
	// 根据sale排序
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))

	books, err := b.BookService.GetHotBooks(limit)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "获取热门书籍失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "获取热门书籍成功",
		"data":    books,
	})
}
func (b *BookController) GetNewBooks(ctx *gin.Context) {
	// 根据created_at排序
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))

	books, err := b.BookService.GetNewBooks(limit)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "获取最新书籍失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "获取最新书籍成功",
		"data":    books,
	})
}
