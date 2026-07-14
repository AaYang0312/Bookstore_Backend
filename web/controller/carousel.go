package controller

import (
	"bookstore-manager/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CarouselController struct {
	CarouselService *service.CarouselService
}

func NewCarouselController() *CarouselController {
	return &CarouselController{
		CarouselService: service.NewCarouselService(),
	}
}

func (c *CarouselController) GetCarouselList(ctx *gin.Context) {
	carousels, err := c.CarouselService.GetCarouselList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取轮播图失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取轮播图成功",
		"data":    carousels,
	})
}
