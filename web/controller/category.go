package controller

import (
	"bookstore-manager/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	CategoryService *service.CategoryService
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
