package controller

import (
	"bookstore-manager/service"
	"net/http"
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
func (b *BookController) GetBookList(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "12"))

	books, total, err := b.BookService.GetBooksByPage(page, pageSize)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "获取书籍列表失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "获取书籍列表成功",
		"data": gin.H{
			"books":      books,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_size": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (b *BookController) SearchBooks(ctx *gin.Context) {
	keyword := ctx.Query("q")
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "搜索关键词不能为空",
		})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "12"))
	books, total, err := b.BookService.SearchBooksWithPage(keyword, page, pageSize)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "搜索书籍失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "搜索书籍成功",
		"data": gin.H{
			"books":      books,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_size": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
