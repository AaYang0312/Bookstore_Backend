package controller

import (
	"bookstore-manager/model"
	"bookstore-manager/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	CategoryService *service.CategoryService
}

func (c *CategoryController) AdminListCategories(ctx *gin.Context) {
	categories, err := c.CategoryService.AdminGetCategories()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取分类列表失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取分类列表成功", "data": gin.H{"categories": categories}})
}

func (c *CategoryController) AdminCreateCategory(ctx *gin.Context) {
	var category model.Category
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误", "error": err.Error()})
		return
	}
	result, err := c.CategoryService.AdminCreateCategory(&category)
	if err != nil {
		writeAdminCategoryError(ctx, "新增分类失败", err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"code": 0, "message": "新增分类成功", "data": result})
}

func (c *CategoryController) AdminUpdateCategory(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "分类")
	if !ok {
		return
	}
	var category model.Category
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误", "error": err.Error()})
		return
	}
	category.ID = id
	result, err := c.CategoryService.AdminUpdateCategory(&category)
	if err != nil {
		writeAdminCategoryError(ctx, "更新分类失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新分类成功", "data": result})
}

func (c *CategoryController) AdminUpdateCategoryStatus(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "分类")
	if !ok {
		return
	}
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.IsActive == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误，缺少is_active"})
		return
	}
	result, err := c.CategoryService.AdminUpdateCategoryStatus(id, *req.IsActive)
	if err != nil {
		writeAdminCategoryError(ctx, "更新分类状态失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新分类状态成功", "data": result})
}

func writeAdminCategoryError(ctx *gin.Context, message string, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "分类不存在"})
	case errors.Is(err, service.ErrCategoryNameRequired):
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": message, "error": err.Error()})
	}
}

func NewCategoryController() *CategoryController {
	return &CategoryController{
		CategoryService: service.NewCategoryService(),
	}
}

func (c *CategoryController) GetCategoryList(ctx *gin.Context) {
	categories, err := c.CategoryService.GetCategoryList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取分类失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "获取分类列表成功",
		"data":    categories,
	})
}
