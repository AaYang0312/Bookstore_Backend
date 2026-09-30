package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"time"

	"gorm.io/gorm"
)

// ProfileDAO 聚合画像查询：订单 / 收藏分布 / 浏览分布，供 /user/agent-profile 使用
type ProfileDAO struct {
	db *gorm.DB
}

func NewProfileDAO() *ProfileDAO {
	return &ProfileDAO{db: global.GetDB()}
}

// CategoryCount 分类计数
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ProfileOrderItem 画像订单项（书名 + 分类）
type ProfileOrderItem struct {
	OrderID  int    `json:"-"`
	BookID   int    `json:"book_id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

// ProfileOrder 画像订单摘要
type ProfileOrder struct {
	OrderID     int               `json:"order_id"`
	OrderNo     string            `json:"order_no"`
	TotalAmount int               `json:"total_amount"`
	Status      int               `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	Items       []ProfileOrderItem `json:"items"`
}

// ProfileBrowsedBook 最近浏览图书
type ProfileBrowsedBook struct {
	BookID   int       `json:"book_id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Type     string    `json:"type"`
	ViewedAt time.Time `json:"viewed_at"`
}

// GetRecentOrders 近 N 笔订单（含订单项与图书分类）
func (p *ProfileDAO) GetRecentOrders(userID, limit int) ([]*ProfileOrder, error) {
	var orders []*ProfileOrder
	err := p.db.Model(&model.Order{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).
		Scan(&orders).Error
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return orders, nil
	}

	ids := make([]int, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.OrderID)
	}

	var items []ProfileOrderItem
	err = p.db.Table("order_items").
		Select(`order_items.order_id, order_items.book_id, order_items.quantity, order_items.price,
			books.title, COALESCE(categories.name, '未分类') AS category`).
		Joins("JOIN books ON books.id = order_items.book_id").
		Joins("LEFT JOIN categories ON categories.id = books.category_id").
		Where("order_items.order_id IN ?", ids).
		Scan(&items).Error
	if err != nil {
		return nil, err
	}

	itemByOrder := make(map[int][]ProfileOrderItem)
	for _, item := range items {
		itemByOrder[item.OrderID] = append(itemByOrder[item.OrderID], item)
	}
	for _, o := range orders {
		o.Items = itemByOrder[o.OrderID]
		if o.Items == nil {
			o.Items = []ProfileOrderItem{}
		}
	}
	return orders, nil
}

// GetFavoriteCategories 收藏分类分布（Top N）
func (p *ProfileDAO) GetFavoriteCategories(userID, limit int) ([]CategoryCount, error) {
	var rows []CategoryCount
	err := p.db.Table("favorites").
		Select("COALESCE(categories.name, '未分类') AS category, COUNT(*) AS count").
		Joins("JOIN books ON books.id = favorites.book_id").
		Joins("LEFT JOIN categories ON categories.id = books.category_id").
		Where("favorites.user_id = ?", userID).
		Group("categories.name").
		Order("count DESC, category ASC").Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetBrowseTopCategories 浏览 Top N 分类
func (p *ProfileDAO) GetBrowseTopCategories(userID, limit int) ([]CategoryCount, error) {
	var rows []CategoryCount
	err := p.db.Table("browse_logs").
		Select("COALESCE(categories.name, '未分类') AS category, COUNT(*) AS count").
		Joins("JOIN books ON books.id = browse_logs.book_id").
		Joins("LEFT JOIN categories ON categories.id = books.category_id").
		Where("browse_logs.user_id = ?", userID).
		Group("categories.name").
		Order("count DESC, category ASC").Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetRecentBrowsed 最近浏览的 N 本书
func (p *ProfileDAO) GetRecentBrowsed(userID, limit int) ([]*ProfileBrowsedBook, error) {
	var rows []*ProfileBrowsedBook
	err := p.db.Table("browse_logs").
		Select(`browse_logs.book_id, books.title, books.author, books.type, browse_logs.viewed_at`).
		Joins("JOIN books ON books.id = browse_logs.book_id AND books.status = ?", 1).
		Where("browse_logs.user_id = ?", userID).
		Order("browse_logs.viewed_at DESC").Limit(limit).
		Scan(&rows).Error
	return rows, err
}
