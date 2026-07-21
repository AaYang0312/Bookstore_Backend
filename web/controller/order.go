package controller

import (
	"bookstore-manager/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct {
	OrderService *service.OrderService
}

func NewOrderController() *OrderController {
	return &OrderController{
		OrderService: service.NewOrderService(),
	}
}

type CreateOrderRequest struct {
	UserID int          `json:"user_id"`
	Items  []OrderItems `json:"items"`
}
type OrderItems struct {
	BookID   int `json:"book_id"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"`
}

func (o *OrderController) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	req.UserID = userID.(int)

	order, err := o.OrderService.CreateOrder(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "创建订单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    order,
		"message": "创建订单成功",
	})
}
func (o *OrderController) GetUserOrders(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "用户未登录",
		})
		return
	}
	orders, total, err := o.OrderService.GetUserOrders(userID, page, pageSize)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    -1,
			"message": "获取订单列表失败",
			"error":   err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "获取订单信息成功",
		"data": gin.H{
			"orders":      orders,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize-1)) / int64(pageSize),
		},
	})
}
func (o *OrderController) GetOrderDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的订单ID"})
		return
	}
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "用户未登录"})
		return
	}
	order, err := o.OrderService.GetUserOrderByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "订单不存在"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取订单成功", "data": order})
}
func (o *OrderController) PayOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的订单ID",
		})
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "用户未登录"})
		return
	}

	err = o.OrderService.PayOrder(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "支付失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "支付成功",
	})
}

func (o *OrderController) AdminListOrders(ctx *gin.Context) {
	page := parseAdminPositiveQuery(ctx, "page", 1)
	pageSize := parseAdminPositiveQuery(ctx, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	var status *int
	if raw := strings.TrimSpace(ctx.Query("status")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 2 {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "订单状态只能是0、1或2"})
			return
		}
		status = &value
	}
	orders, total, err := o.OrderService.AdminGetOrders(ctx.Query("keyword"), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取订单列表失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取订单列表成功", "data": gin.H{
		"orders": orders, "total": total, "page": page, "page_size": pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	}})
}

func (o *OrderController) AdminGetOrder(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "订单")
	if !ok {
		return
	}
	order, err := o.OrderService.AdminGetOrderByID(id)
	if err != nil {
		writeAdminOrderError(ctx, "获取订单详情失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "获取订单详情成功", "data": order})
}

func (o *OrderController) AdminUpdateOrderStatus(ctx *gin.Context) {
	id, ok := parseAdminResourceID(ctx, "订单")
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
	order, err := o.OrderService.AdminUpdateOrderStatus(id, *req.Status)
	if err != nil {
		writeAdminOrderError(ctx, "更新订单状态失败", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新订单状态成功", "data": order})
}

func writeAdminOrderError(ctx *gin.Context, message string, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "订单不存在"})
		return
	}
	if err.Error() == "订单状态只能是0、1或2" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	if errors.Is(err, service.ErrPaidOrderStatusLocked) {
		ctx.JSON(http.StatusConflict, gin.H{"code": -1, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": message, "error": err.Error()})
}
