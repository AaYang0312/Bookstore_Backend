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

	// 异步下单：消息进入 Kafka 后立即返回，前端轮询 /order/create/result 获取结果
	result, err := o.OrderService.SubmitOrder(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrMQUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if result.Status == service.OrderCreateCreated {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"data":    result,
			"message": "订单已创建",
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"data":    result,
		"message": "订单已受理，正在处理",
	})
}

// GetOrderCreateResult 查询异步下单结果，前端在提交订单后轮询该接口。
func (o *OrderController) GetOrderCreateResult(c *gin.Context) {
	key := strings.TrimSpace(c.Query("idempotency_key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "缺少幂等键"})
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "用户未登录"})
		return
	}
	result, err := o.OrderService.GetCreateResult(c.Request.Context(), userID, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data":    result,
	})
}

// CancelOrder 用户取消订单（仅待支付）。
func (o *OrderController) CancelOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的订单ID"})
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "用户未登录"})
		return
	}
	if err := o.OrderService.UserCancelOrder(id, userID); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "订单已取消"})
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
