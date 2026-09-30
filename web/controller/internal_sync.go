package controller

import (
	"net/http"
	"strconv"
	"time"

	"bookstore-manager/repository"
	"bookstore-manager/service"

	"github.com/gin-gonic/gin"
)

// InternalSyncController 内部只读同步接口：供 Agent 向量库入库管道增量拉取图书。
// 通过共享密钥（X-Internal-Token）鉴权，不走管理端 JWT，也无限流。
type InternalSyncController struct {
	BookService *service.BookService
}

func NewInternalSyncController() *InternalSyncController {
	return &InternalSyncController{
		BookService: service.NewBookService(),
	}
}

// BooksSync GET /internal/books/sync?since=&page=&page_size=
func (i *InternalSyncController) BooksSync(ctx *gin.Context) {
	var since *time.Time
	if raw := ctx.Query("since"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    -1,
				"message": "since 参数格式错误，应为 RFC3339 时间",
			})
			return
		}
		since = &parsed
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "100"))
	if err != nil || pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	books, total, nextSince, err := i.BookService.GetBooksForSync(since, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "同步图书数据失败",
			"error":   err.Error(),
		})
		return
	}
	if books == nil {
		books = []*repository.AdminBook{} // 空结果返回 [] 而非 null
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	data := gin.H{
		"books":       books,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}
	// 空结果集时 MAX(updated_at) 为 NULL（零值时间），省略 next_since 而非返回 0001-01-01
	if !nextSince.IsZero() {
		data["next_since"] = nextSince.UTC().Format(time.RFC3339Nano)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "同步图书数据成功",
		"data":    data,
	})
}
