package controller

import (
	"net/http"
	"strconv"

	"bookstore-manager/service"

	"github.com/gin-gonic/gin"
)

type BrowseLogController struct {
	BrowseLogService *service.BrowseLogService
}

func NewBrowseLogController() *BrowseLogController {
	return &BrowseLogController{
		BrowseLogService: service.NewBrowseLogService(),
	}
}

// RecordBookView 显式浏览埋点（JWT），供前端后续接入
func (b *BrowseLogController) RecordBookView(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	bookID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || bookID < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的书籍ID",
		})
		return
	}
	b.BrowseLogService.RecordView(userID, bookID)
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "记录浏览成功",
	})
}

// GetBrowseHistory 当前用户浏览记录（JWT，分页）
func (b *BrowseLogController) GetBrowseHistory(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	records, total, err := b.BrowseLogService.GetBrowseHistory(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取浏览记录失败",
			"error":   err.Error(),
		})
		return
	}
	totalPages := (int(total) + pageSize - 1) / pageSize
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取浏览记录成功",
		"data": gin.H{
			"records":     records,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// GetAgentProfile 聚合画像（JWT）：近 10 笔订单、收藏分类分布、浏览 Top5 分类、最近浏览
func (b *BrowseLogController) GetAgentProfile(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	profile, err := b.BrowseLogService.GetAgentProfile(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取用户画像失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取用户画像成功",
		"data":    profile,
	})
}
