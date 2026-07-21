package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderDAO struct {
	db *gorm.DB
}

// AdminOrder 是管理端订单列表使用的扁平结构。
type AdminOrder struct {
	ID          int        `json:"id"`
	OrderNo     string     `json:"order_no"`
	UserID      int        `json:"user_id"`
	Username    string     `json:"username"`
	TotalAmount int        `json:"total_amount"`
	Status      int        `json:"status"`
	IsPaid      bool       `json:"is_paid"`
	ItemCount   int        `json:"item_count"`
	PaymentTime *time.Time `json:"payment_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewOrderDAO() *OrderDAO {
	return &OrderDAO{
		db: global.GetDB(),
	}
}

// CreateOrder 创建订单
func (o *OrderDAO) CreateOrder(order *model.Order) error {
	err := o.db.Create(order).Error
	if err != nil {
	} else {
	}
	return err
}

// GetOrderByID 根据ID获取订单
func (o *OrderDAO) GetOrderByID(id int) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("OrderItems.Book").First(&order, id).Error
	if err != nil {
	} else {
	}
	return &order, err
}

func (o *OrderDAO) GetOrderByUserAndID(id, userID int) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("OrderItems.Book").
		Where("id = ? AND user_id = ?", id, userID).
		First(&order).Error
	return &order, err
}

func (o *OrderDAO) GetOrderByIdempotencyKey(userID int, key string) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("OrderItems.Book").
		Where("user_id = ? AND idempotency_key = ?", userID, key).
		First(&order).Error
	return &order, err
}

// GetOrderByOrderNo 根据订单号获取订单
func (o *OrderDAO) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("OrderItems.Book").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
	} else {
	}
	return &order, err
}

// GetUserOrders 获取用户的订单列表
func (o *OrderDAO) GetUserOrders(userID int, page, pageSize int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	// 获取总数
	err := o.db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = o.db.Preload("OrderItems.Book").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (o *OrderDAO) GetAdminOrders(keyword string, status *int, page, pageSize int) ([]*AdminOrder, int64, error) {
	var orders []*AdminOrder
	var total int64

	query := o.db.Model(&model.Order{}).Joins("LEFT JOIN users ON users.id = orders.user_id")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("orders.order_no LIKE ? OR users.username LIKE ?", like, like)
	}
	if status != nil {
		query = query.Where("orders.status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Select(`orders.id, orders.order_no, orders.user_id, users.username,
			orders.total_amount, orders.status, orders.is_paid, orders.payment_time,
			orders.created_at, orders.updated_at,
			(SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE order_id = orders.id) AS item_count`).
		Order("orders.created_at DESC").Offset(offset).Limit(pageSize).Scan(&orders).Error
	return orders, total, err
}

func (o *OrderDAO) GetAdminOrderByID(id int) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("User").Preload("OrderItems.Book").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	if order.User != nil {
		order.User.Password = ""
	}
	return &order, nil
}

func (o *OrderDAO) UpdateAdminOrderStatus(id, status int) error {
	updates := map[string]any{"status": status, "is_paid": status == 1}
	if status == 1 {
		updates["payment_time"] = gorm.Expr("COALESCE(payment_time, NOW())")
	} else {
		updates["payment_time"] = nil
	}
	result := o.db.Model(&model.Order{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := o.db.Model(&model.Order{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (o *OrderDAO) PayOrder(orderID, userID int) error {
	// 订单号、销量的更新、库存的减少、金额的更新、订单的状态（0/1）
	// 使用事务处理支付和库存更新
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("OrderItems").
			Where("id = ? AND user_id = ?", orderID, userID).
			First(&order).Error; err != nil {
			return err
		}
		if order.IsPaid {
			return nil
		}

		// 再次检查库存（防止并发问题）
		for _, item := range order.OrderItems {
			var book model.Book
			if err := tx.First(&book, item.BookID).Error; err != nil {
				return errors.New("图书不存在")
			}
			if book.Stock < item.Quantity {
				return errors.New("库存不足")
			}
		}

		// 标记订单为已支付
		if err := tx.Model(&model.Order{}).
			Where("id = ? AND user_id = ? AND is_paid = ?", order.ID, userID, false).
			Updates(map[string]interface{}{
				"status":       1,
				"is_paid":      true,
				"payment_time": gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}

		// 更新图书库存和销售量
		for _, item := range order.OrderItems {
			if err := tx.Model(&model.Book{}).
				Where("id = ?", item.BookID).
				Updates(map[string]interface{}{
					"stock": gorm.Expr("stock - ?", item.Quantity),
					"sale":  gorm.Expr("sale + ?", item.Quantity),
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// GenerateOrderNo 生成订单号
func (o *OrderDAO) GenerateOrderNo() string {
	orderNo := fmt.Sprintf("ORD%d", time.Now().UnixNano())
	return orderNo
}

// CreateOrderWithItems 创建订单和订单项
func (o *OrderDAO) CreateOrderWithItems(order *model.Order, items []*model.OrderItem) error {
	err := o.db.Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 创建订单项
		for _, item := range items {
			item.OrderID = order.ID
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
	} else {
	}
	return err
}
