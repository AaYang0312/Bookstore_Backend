package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"strings"
)

type CreateOrderRequest struct {
	UserID         int          `json:"user_id"`
	IdempotencyKey string       `json:"idempotency_key"`
	Items          []OrderItems `json:"items"`
}
type OrderItems struct {
	BookID   int `json:"book_id"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"`
}
type OrderService struct {
	OrderDAO *repository.OrderDAO
	BookDAO  *repository.BookDAO
}

func NewOrderService() *OrderService {
	return &OrderService{
		OrderDAO: repository.NewOrderDAO(),
		BookDAO:  repository.NewBookDAO(),
	}
}
func (o *OrderService) CreateOrder(req *CreateOrderRequest) (*model.Order, error) {
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return nil, errors.New("缺少幂等键")
	}
	if len(req.IdempotencyKey) > 64 {
		return nil, errors.New("幂等键过长")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("订单项不能为空")
	}
	if existing, err := o.OrderDAO.GetOrderByIdempotencyKey(req.UserID, req.IdempotencyKey); err == nil {
		return existing, nil
	}

	// 检查库存
	err := o.checkStockAvailability(req.Items)
	if err != nil {
		return nil, err
	}

	// 生成订单号
	orderNo := o.OrderDAO.GenerateOrderNo()

	// 计算总金额
	var totalAmount int
	var orderItems []*model.OrderItem

	for _, item := range req.Items {
		subtotal := item.Price * item.Quantity
		totalAmount += subtotal

		orderItems = append(orderItems, &model.OrderItem{
			BookID:   item.BookID,
			Quantity: item.Quantity,
			Price:    item.Price,
			Subtotal: subtotal,
		})
	}

	// 创建订单
	order := &model.Order{
		UserID:         req.UserID,
		OrderNo:        orderNo,
		IdempotencyKey: req.IdempotencyKey,
		TotalAmount:    totalAmount,
		Status:         0, // 待支付
		IsPaid:         false,
	}

	// 创建订单和订单项
	err = o.OrderDAO.CreateOrderWithItems(order, orderItems)
	if err != nil {
		if existing, lookupErr := o.OrderDAO.GetOrderByIdempotencyKey(req.UserID, req.IdempotencyKey); lookupErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return order, nil
}

// checkStockAvailability 检查库存是否充足
func (o *OrderService) checkStockAvailability(items []OrderItems) error {
	for _, item := range items {
		book, err := o.BookDAO.GetBookByID(item.BookID)
		if err != nil {
			return errors.New("图书不存在")
		}

		if book.Status != 1 {
			return errors.New("图书已下架")
		}

		if book.Stock < item.Quantity {
			return errors.New("库存不足")
		}
	}
	return nil
}

// GetUserOrders 获取用户的订单列表
func (o *OrderService) GetUserOrders(userID int, page, pageSize int) ([]*model.Order, int64, error) {
	return o.OrderDAO.GetUserOrders(userID, page, pageSize)
}

// PayOrder 支付订单
func (o *OrderService) PayOrder(orderID, userID int) error {
	return o.OrderDAO.PayOrder(orderID, userID)
}

// GetOrderByID 根据ID获取订单
func (o *OrderService) GetOrderByID(id int) (*model.Order, error) {
	return o.OrderDAO.GetOrderByID(id)
}

func (o *OrderService) GetUserOrderByID(id, userID int) (*model.Order, error) {
	return o.OrderDAO.GetOrderByUserAndID(id, userID)
}
