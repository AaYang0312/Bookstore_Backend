package controller

import (
	"bookstore-manager/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	AdminService *service.AdminService
}

func NewAdminController() *AdminController {
	return &AdminController{AdminService: service.NewAdminService()}
}

func (a *AdminController) GetDashboard(ctx *gin.Context) {
	dashboard, err := a.AdminService.GetDashboard()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": -1, "message": "获取管理后台概览失败", "error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取管理后台概览成功", "data": dashboard})
}

func parseAdminResourceID(ctx *gin.Context, resourceName string) (int, bool) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的" + resourceName + "ID"})
		return 0, false
	}
	return id, true
}
