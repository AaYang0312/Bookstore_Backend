package service

import (
	"bookstore-manager/global"
	"bookstore-manager/mq"
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrPaidOrderStatusLocked = errors.New("已支付订单不能直接改为待支付或已取消")
	ErrOrderNotFound         = errors.New("订单不存在")
	ErrMQUnavailable         = errors.New("消息队列暂不可用，请稍后重试")
)

// 下单受理/处理结果状态
const (
	OrderCreatePending = "pending" // 已进入消息队列，处理中
	OrderCreateCreated = "created" // 订单已创建成功
	OrderCreateFailed  = "failed"  // 业务校验失败（库存不足等）
)

const (
	// 超过该时长未支付的订单将被定时任务自动取消
	orderPayTimeout = 30 * time.Minute
	// 超时关单扫描间隔
	orderTimeoutScanInterval = time.Minute
	// 下单失败原因在 Redis 中的保留时长
	orderFailResultTTL = 10 * time.Minute
)

type CreateOrderRequest struct {
	UserID         int          `json:"user_id"`
	IdempotencyKey string       `json:"idempotency_key"`
	Items          []OrderItems `json:"items"`
}
type OrderItems struct {
	BookID   int `json:"book_id"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"` // 已废弃：金额一律由服务端按数据库价格计算
}

// SubmitOrderResult POST /order/create 的受理结果
type SubmitOrderResult struct {
	Status         string `json:"status"` // pending / created
	OrderID        int    `json:"order_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key"`
}

// OrderCreateResult GET /order/create/result 的轮询结果
type OrderCreateResult struct {
	Status  string `json:"status"` // pending / created / failed
	OrderID int    `json:"order_id,omitempty"`
	OrderNo string `json:"order_no,omitempty"`
	Message string `json:"message,omitempty"`
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

// SubmitOrder 受理下单请求：参数校验 + 幂等快查后发送 Kafka 消息，
// 真正的建单逻辑由消费者异步执行。
func (o *OrderService) SubmitOrder(ctx context.Context, req *CreateOrderRequest) (*SubmitOrderResult, error) {
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
	for _, item := range req.Items {
		if item.BookID <= 0 {
			return nil, errors.New("图书ID不合法")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("购买数量必须大于0")
		}
	}

	// 幂等快查：该幂等键的订单已创建则直接返回，前端可立即跳转支付
	if existing, err := o.OrderDAO.GetOrderByIdempotencyKey(req.UserID, req.IdempotencyKey); err == nil {
		return &SubmitOrderResult{
			Status:         OrderCreateCreated,
			OrderID:        existing.ID,
			IdempotencyKey: req.IdempotencyKey,
		}, nil
	}

	if mq.Producer == nil {
		return nil, ErrMQUnavailable
	}
	msg := &mq.OrderCreateMessage{
		UserID:         req.UserID,
		IdempotencyKey: req.IdempotencyKey,
		Items:          make([]mq.OrderCreateMessageItem, 0, len(req.Items)),
		CreatedAt:      time.Now(),
	}
	for _, item := range req.Items {
		msg.Items = append(msg.Items, mq.OrderCreateMessageItem{
			BookID:   item.BookID,
			Quantity: item.Quantity,
		})
	}
	if err := mq.Producer.SendOrderCreate(ctx, msg); err != nil {
		log.Printf("发送下单消息失败: user=%d key=%s error=%v", req.UserID, req.IdempotencyKey, err)
		return nil, ErrMQUnavailable
	}
	return &SubmitOrderResult{
		Status:         OrderCreatePending,
		IdempotencyKey: req.IdempotencyKey,
	}, nil
}

// HandleOrderCreate Kafka 消费者回调：异步执行建单。
// 业务失败（图书不存在/下架/库存不足）写入 Redis 结果标记并返回 nil（消息正常提交）；
// 返回 error 视为系统故障，消费者不提交位点、消息重投。
func (o *OrderService) HandleOrderCreate(ctx context.Context, msg *mq.OrderCreateMessage) error {
	// 幂等复查：消息重投或重复发送时直接跳过
	if existing, err := o.OrderDAO.GetOrderByIdempotencyKey(msg.UserID, msg.IdempotencyKey); err == nil {
		log.Printf("下单消息命中幂等键（订单 %d 已存在），跳过", existing.ID)
		return nil
	}

	// 一次性查出订单涉及的全部图书
	bookIDs := make([]int, 0, len(msg.Items))
	for _, item := range msg.Items {
		bookIDs = append(bookIDs, item.BookID)
	}
	books, err := o.BookDAO.GetBooksByIDs(bookIDs)
	if err != nil {
		return fmt.Errorf("查询图书失败: %w", err)
	}
	bookMap := make(map[int]*model.Book, len(books))
	for _, book := range books {
		bookMap[book.ID] = book
	}

	// 校验并按数据库价格计算金额
	var totalAmount int
	var orderItems []*model.OrderItem
	for _, item := range msg.Items {
		book, ok := bookMap[item.BookID]
		if !ok {
			return o.rejectOrderCreate(ctx, msg, "图书不存在")
		}
		if book.Status != 1 {
			return o.rejectOrderCreate(ctx, msg, fmt.Sprintf("《%s》已下架", book.Title))
		}
		if book.Stock < item.Quantity {
			return o.rejectOrderCreate(ctx, msg, fmt.Sprintf("《%s》库存不足", book.Title))
		}
		price := effectivePrice(book.Price, book.Discount)
		subtotal := price * item.Quantity
		totalAmount += subtotal
		orderItems = append(orderItems, &model.OrderItem{
			BookID:   item.BookID,
			Quantity: item.Quantity,
			Price:    price,
			Subtotal: subtotal,
		})
	}

	order := &model.Order{
		UserID:         msg.UserID,
		OrderNo:        o.OrderDAO.GenerateOrderNo(),
		IdempotencyKey: msg.IdempotencyKey,
		TotalAmount:    totalAmount,
		Status:         0, // 待支付
		IsPaid:         false,
	}
	if err := o.OrderDAO.CreateOrderWithItems(order, orderItems); err != nil {
		// 并发消息命中唯一索引时按幂等成功处理
		if existing, lookupErr := o.OrderDAO.GetOrderByIdempotencyKey(msg.UserID, msg.IdempotencyKey); lookupErr == nil {
			log.Printf("订单落库命中幂等键（订单 %d 已存在），跳过", existing.ID)
			return nil
		}
		return fmt.Errorf("订单落库失败: %w", err)
	}
	log.Printf("异步建单成功: order_id=%d order_no=%s user=%d total=%d", order.ID, order.OrderNo, msg.UserID, totalAmount)
	return nil
}

// effectivePrice 服务端计算现价，口径与前端 bookPrice.js 一致：
// 折扣在 (0,100] 时按 price*(100-discount)/100 向下取整，否则取原价。
func effectivePrice(price, discount int) int {
	if discount > 0 && discount <= 100 {
		return price * (100 - discount) / 100
	}
	return price
}

// rejectOrderCreate 记录业务失败原因并放行消息（正常提交位点）
func (o *OrderService) rejectOrderCreate(ctx context.Context, msg *mq.OrderCreateMessage, reason string) error {
	log.Printf("下单业务失败: user=%d key=%s reason=%s", msg.UserID, msg.IdempotencyKey, reason)
	if global.RedisClient != nil {
		key := fmt.Sprintf("order:create:fail:%d:%s", msg.UserID, msg.IdempotencyKey)
		if err := global.RedisClient.Set(ctx, key, reason, orderFailResultTTL).Err(); err != nil {
			log.Println("写入下单失败标记失败：", err)
		}
	}
	return nil
}

// GetCreateResult 查询下单处理结果：先查订单（created），再查 Redis 失败标记（failed），
// 都没有则仍在处理中（pending）。
func (o *OrderService) GetCreateResult(ctx context.Context, userID int, key string) (*OrderCreateResult, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("缺少幂等键")
	}
	if existing, err := o.OrderDAO.GetOrderByIdempotencyKey(userID, key); err == nil {
		return &OrderCreateResult{
			Status:  OrderCreateCreated,
			OrderID: existing.ID,
			OrderNo: existing.OrderNo,
		}, nil
	}
	if global.RedisClient != nil {
		redisKey := fmt.Sprintf("order:create:fail:%d:%s", userID, key)
		if reason, err := global.RedisClient.Get(ctx, redisKey).Result(); err == nil {
			return &OrderCreateResult{Status: OrderCreateFailed, Message: reason}, nil
		}
	}
	return &OrderCreateResult{Status: OrderCreatePending}, nil
}

// GetUserOrders 获取用户的订单列表
func (o *OrderService) GetUserOrders(userID int, page, pageSize int) ([]*model.Order, int64, error) {
	return o.OrderDAO.GetUserOrders(userID, page, pageSize)
}

// PayOrder 支付订单
func (o *OrderService) PayOrder(orderID, userID int) error {
	return o.OrderDAO.PayOrder(orderID, userID)
}

// UserCancelOrder 用户取消订单，仅待支付订单可取消。
func (o *OrderService) UserCancelOrder(orderID, userID int) error {
	order, err := o.OrderDAO.GetOrderByUserAndID(orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if order.Status == 2 {
		return nil // 已取消视为幂等成功
	}
	if order.Status != 0 {
		return errors.New("仅待支付订单可以取消")
	}
	cancelled, err := o.OrderDAO.CancelOrder(orderID, userID)
	if err != nil {
		return err
	}
	if !cancelled {
		return errors.New("订单状态已变更，请刷新后查看")
	}
	return nil
}

// StartOrderTimeoutWorker 超时关单定时任务：扫描并取消超时未支付订单，阻塞直到 ctx 取消。
func (o *OrderService) StartOrderTimeoutWorker(ctx context.Context) {
	ticker := time.NewTicker(orderTimeoutScanInterval)
	defer ticker.Stop()
	log.Printf("超时关单任务启动: 扫描间隔 %s, 未支付超过 %s 自动取消", orderTimeoutScanInterval, orderPayTimeout)
	for {
		select {
		case <-ctx.Done():
			log.Println("超时关单任务退出")
			return
		case <-ticker.C:
			deadline := time.Now().Add(-orderPayTimeout)
			count, err := o.OrderDAO.CancelExpiredOrders(deadline)
			if err != nil {
				log.Println("超时关单扫描失败：", err)
				continue
			}
			if count > 0 {
				log.Printf("超时关单: %d 个待支付订单已自动取消", count)
			}
		}
	}
}

// GetOrderByID 根据ID获取订单
func (o *OrderService) GetOrderByID(id int) (*model.Order, error) {
	return o.OrderDAO.GetOrderByID(id)
}

func (o *OrderService) GetUserOrderByID(id, userID int) (*model.Order, error) {
	return o.OrderDAO.GetOrderByUserAndID(id, userID)
}

func (o *OrderService) AdminGetOrders(keyword string, status *int, page, pageSize int) ([]*repository.AdminOrder, int64, error) {
	return o.OrderDAO.GetAdminOrders(strings.TrimSpace(keyword), status, page, pageSize)
}

func (o *OrderService) AdminGetOrderByID(id int) (*model.Order, error) {
	return o.OrderDAO.GetAdminOrderByID(id)
}

func (o *OrderService) AdminUpdateOrderStatus(id, status int) (*model.Order, error) {
	if status < 0 || status > 2 {
		return nil, errors.New("订单状态只能是0、1或2")
	}
	order, err := o.OrderDAO.GetAdminOrderByID(id)
	if err != nil {
		return nil, err
	}
	if order.Status == status {
		return order, nil
	}
	if order.IsPaid && status != 1 {
		return nil, ErrPaidOrderStatusLocked
	}
	if status == 1 {
		if err := o.OrderDAO.PayOrder(id, order.UserID); err != nil {
			return nil, err
		}
		return o.OrderDAO.GetAdminOrderByID(id)
	}
	if err := o.OrderDAO.UpdateAdminOrderStatus(id, status); err != nil {
		return nil, err
	}
	return o.OrderDAO.GetAdminOrderByID(id)
}
