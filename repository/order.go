package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type OrderDAO struct {
	db *gorm.DB
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

func (o *OrderDAO) PayOrder(order *model.Order) error {
	// 订单号、销量的更新、库存的减少、金额的更新、订单的状态（0/1）
	// 使用事务处理支付和库存更新
	err := o.db.Transaction(func(tx *gorm.DB) error {
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
			Where("id = ?", order.ID).
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
