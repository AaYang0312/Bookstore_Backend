package mq

import "time"

// OrderCreateMessage 下单消息。金额由消费者按数据库价格计算，
// 消息中不携带任何价格信息，防止客户端篡改。
type OrderCreateMessage struct {
	UserID         int                  `json:"user_id"`
	IdempotencyKey string               `json:"idempotency_key"`
	Items          []OrderCreateMessageItem `json:"items"`
	CreatedAt      time.Time            `json:"created_at"`
}

type OrderCreateMessageItem struct {
	BookID   int `json:"book_id"`
	Quantity int `json:"quantity"`
}
