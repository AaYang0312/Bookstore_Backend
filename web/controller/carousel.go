package controller

import (
	"bookstore-manager/model"
	"bookstore-manager/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CarouselController struct {
	CarouselService *service.CarouselService
}

func (c *CarouselController) AdminListCarousels(ctx *gin.Context) {
	carousels, err := c.CarouselService.AdminGetCarousels()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取轮播图列表失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取轮播图列表成功", "data": gin.H{"carousel": carousels}})
}

func (c *CarouselController) AdminCreateCarousel(ctx *gin.Context) {
	var carousel model.Carousel
	if err := ctx.ShouldBindJSON(&carousel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误", "error": err.Error()})
		return
	}
	result, err := c.CarouselService.AdminCreateCarousel(&carousel)
	if err != nil {
		writeAdminCarouselError(ctx, "新增轮播图失败", err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"code": 0, "message": "新增轮播图成功", "data": result})
}

func (c *CarouselController) AdminUpdateCarousel(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "轮播图")
	if !ok {
		return
	}
	var carousel model.Carousel
	if err := ctx.ShouldBindJSON(&carousel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请求参数错误", "error": err.Error()})
		return
	}
	carousel.ID = id
	result, err := c.CarouselService.AdminUpdateCarousel(&carousel)
	if err != nil {
		writeAdminCarouselError(ctx, "更新轮播图失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新轮播图成功", "data": result})
}

func (c *CarouselController) AdminDeleteCarousel(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "轮播图")
	if !ok {
		return
	}
	if err := c.CarouselService.AdminDeleteCarousel(id); err != nil {
		writeAdminCarouselError(ctx, "删除轮播图失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除轮播图成功"})
}

func (c *CarouselController) AdminUpdateCarouselStatus(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "轮播图")
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
	result, err := c.CarouselService.AdminUpdateCarouselStatus(id, *req.IsActive)
	if err != nil {
		writeAdminCarouselError(ctx, "更新轮播图状态失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新轮播图状态成功", "data": result})
}

func writeAdminCarouselError(ctx *gin.Context, message string, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "轮播图不存在"})
	case errors.Is(err, service.ErrCarouselTitleRequired):
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": message, "error": err.Error()})
	}
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
