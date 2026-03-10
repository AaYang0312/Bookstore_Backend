package model

import "time"

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderNo     string    `json:"order_no"` // 订单号
	TotalAmount int       `json:"total_amount"`
	Status      int       `json:"status"`  // 状态
	IsPaid      bool      `json:"is_paid"` // 支付状态
	PaymentTime time.Time `json:"payment_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
