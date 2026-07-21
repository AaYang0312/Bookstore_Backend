package controller

import (
	"bookstore-manager/model"
	"bookstore-manager/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
			"books":       books,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
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
			"books":       books,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (b *BookController) GetBookDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{
			"code":    -1,
			"message": "无效的书籍ID",
		})
		return
	}
	intid, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    -1,
			"message": "无效的书籍ID",
		})
		return
	}
	book, err := b.BookService.GetBookDetail(intid)
	if err != nil {
		ctx.JSON(404, gin.H{
			"code":    -1,
			"message": "书籍不存在",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "查找书籍详情成功",
		"data":    book,
	})
}

func (b *BookController) GetBooksByCategory(ctx *gin.Context) {
	categoryName := ctx.Param("category")
	if categoryName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "分类名称不能为空",
		})
		return
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "12"))
	if err != nil || pageSize < 1 {
		pageSize = 12
	}

	books, total, err := b.BookService.GetBooksByCategory(
		categoryName,
		page,
		pageSize,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取分类图书失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取分类图书成功",
		"data": gin.H{
			"books":       books,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (b *BookController) AdminListBooksHandler(ctx *gin.Context) {
	page := parseAdminPositiveQuery(ctx, "page", 1)
	pageSize := parseAdminPositiveQuery(ctx, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	var status *int
	if rawStatus := strings.TrimSpace(ctx.Query("status")); rawStatus != "" {
		value, err := strconv.Atoi(rawStatus)
		if err != nil || (value != 0 && value != 1) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "图书状态只能是0或1"})
			return
		}
		status = &value
	}

	books, total, err := b.BookService.AdminGetBooks(strings.TrimSpace(ctx.Query("keyword")), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取管理端图书列表失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取管理端图书列表成功",
		"data": gin.H{
			"books":       books,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (b *BookController) AdminCreateBookHandler(ctx *gin.Context) {
	book, categoryName, ok := bindBookForm(ctx, 0)
	if !ok {
		return
	}
	result, err := b.BookService.AdminCreateBook(book, categoryName)
	if err != nil {
		writeAdminBookHandlerError(ctx, "新增图书失败", err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"code": 0, "message": "新增图书成功", "data": result})
}

func (b *BookController) AdminUpdateBookHandler(ctx *gin.Context) {
	id, ok := parseAdminBookHandlerID(ctx)
	if !ok {
		return
	}
	book, categoryName, ok := bindBookForm(ctx, id)
	if !ok {
		return
	}
	result, err := b.BookService.AdminUpdateBook(book, categoryName)
	if err != nil {
		writeAdminBookHandlerError(ctx, "更新图书失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新图书成功", "data": result})
}

func (b *BookController) AdminUpdateBookStatusHandler(ctx *gin.Context) {
	id, ok := parseAdminBookHandlerID(ctx)
	if !ok {
		return
	}
	var req struct {
		Status *int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Status == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误，缺少status"})
		return
	}
	if *req.Status != 0 && *req.Status != 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "图书状态只能是0或1"})
		return
	}
	book, err := b.BookService.AdminUpdateBookStatus(id, *req.Status)
	if err != nil {
		writeAdminBookHandlerError(ctx, "更新图书状态失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新图书状态成功", "data": book})
}

func (b *BookController) AdminUpdateBookStockHandler(ctx *gin.Context) {
	id, ok := parseAdminBookHandlerID(ctx)
	if !ok {
		return
	}
	var req struct {
		Stock *int `json:"stock"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Stock == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误，缺少stock"})
		return
	}
	if *req.Stock < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "库存不能小于0"})
		return
	}
	book, err := b.BookService.AdminUpdateBookStock(id, *req.Stock)
	if err != nil {
		writeAdminBookHandlerError(ctx, "更新图书库存失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新图书库存成功", "data": book})
}

func bindBookForm(ctx *gin.Context, id int) (*model.Book, string, bool) {
	var req struct {
		model.Book
		CategoryName string `json:"category_name"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误", "error": err.Error()})
		return nil, "", false
	}

	req.Book.ID = id
	req.Book.Title = strings.TrimSpace(req.Book.Title)
	req.Book.Author = strings.TrimSpace(req.Book.Author)
	req.Book.Type = strings.TrimSpace(req.Book.Type)
	req.Book.Description = strings.TrimSpace(req.Book.Description)
	req.Book.CoverURL = strings.TrimSpace(req.Book.CoverURL)
	req.Book.ISBN = strings.TrimSpace(req.Book.ISBN)
	req.Book.Publisher = strings.TrimSpace(req.Book.Publisher)
	req.Book.PublishDate = strings.TrimSpace(req.Book.PublishDate)
	req.Book.Language = strings.TrimSpace(req.Book.Language)
	req.Book.Format = strings.TrimSpace(req.Book.Format)
	categoryName := strings.TrimSpace(req.CategoryName)

	switch {
	case req.Book.Title == "":
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "书名不能为空"})
	case req.Book.Price < 0 || req.Book.Stock < 0 || req.Book.Pages < 0:
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "价格、库存和页数不能小于0"})
	case req.Book.Discount < 0 || req.Book.Discount > 100:
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "折扣必须在0到100之间"})
	case req.Book.Status != 0 && req.Book.Status != 1:
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "图书状态只能是0或1"})
	default:
		return &req.Book, categoryName, true
	}
	return nil, "", false
}

func parseAdminBookHandlerID(ctx *gin.Context) (int, bool) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的图书ID"})
		return 0, false
	}
	return id, true
}

func parseAdminPositiveQuery(ctx *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(ctx.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func writeAdminBookHandlerError(ctx *gin.Context, message string, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "图书不存在"})
	case errors.Is(err, service.ErrBookCategoryNotFound):
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": message, "error": err.Error()})
	}
}
