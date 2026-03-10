package model

import "time"

type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`       // 书名
	Author      string    `json:"author"`      // 作者
	Price       int       `json:"price"`       // 价格
	Discount    int       `json:"discount"`    // 折扣
	Type        string    `json:"type"`        // 类型
	Stock       int       `json:"stock"`       // 库存
	Status      int       `json:"status"`      // 状态
	Description string    `json:"description"` // 描述
	CoverURL    string    `json:"cover_url"`   // 封面号
	ISBN        string    `json:"isbn"`        // ISBN
	Publisher   string    `json:"publisher"`   // 出版社
	Pages       int       `json:"pages"`       // 页数
	Language    string    `json:"language"`    // 语言
	Format      string    `json:"format"`      // 装帧格式
	CategoryID  uint      `json:"category_id"` // 分类id
	CreatedAt   time.Time `json:"created_at"`
	UpdateAt    time.Time `json:"update_at"`
}
