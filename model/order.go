package model

import (
	"time"
)

type Order struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	UserID      int       `json:"user_id"`
	OrderNo     string    `json:"order_no"` // 订单号
	TotalAmount int       `json:"total_amount"`
	Status      int       `json:"status"`  // 状态，0-待支付，1-已支付，2-已取消
	IsPaid      bool      `json:"is_paid"` // 支付状态
	PaymentTime time.Time `json:"payment_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联字段
	User       *User       `gorm:"foreignKey:UserID" json:"user"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
}

func (o *Order) TableName() string {
	return "orders"
}

// OrderItem 订单项模型
type OrderItem struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	OrderID   int       `gorm:"not null" json:"order_id"` // 订单ID
	BookID    int       `gorm:"not null" json:"book_id"`  // 图书ID
	Quantity  int       `gorm:"not null" json:"quantity"` // 数量
	Price     int       `gorm:"not null" json:"price"`    // 单价（分）
	Subtotal  int       `gorm:"not null" json:"subtotal"` // 小计（分）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联字段
	Book *Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

func (oi *OrderItem) TableName() string {
	return "order_items"
}
