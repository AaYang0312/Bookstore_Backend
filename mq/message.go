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

// 收藏事件动作
const (
	FavoriteActionAdd    = "add"
	FavoriteActionRemove = "remove"
)

// FavoriteEventMessage 收藏事件消息。只携带身份与动作，
// 不携带价格等任何可被客户端篡改的业务数据。
type FavoriteEventMessage struct {
	EventID   string    `json:"event_id"` // uuid，用于日志追踪
	UserID    int       `json:"user_id"`
	BookID    int       `json:"book_id"`
	Action    string    `json:"action"` // "add" | "remove"
	CreatedAt time.Time `json:"created_at"`
}
