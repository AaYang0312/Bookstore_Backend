package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"time"

	"gorm.io/gorm"
)

type AdminDAO struct {
	db *gorm.DB
}

type AdminDashboardCounts struct {
	BookCount         int64 `json:"book_count"`
	UserCount         int64 `json:"user_count"`
	OrderCount        int64 `json:"order_count"`
	Revenue           int64 `json:"revenue"`
	LowStockCount     int64 `json:"low_stock_count"`
	PendingOrderCount int64 `json:"pending_order_count"`
}

type DailySales struct {
	Day    time.Time
	Amount int64
}

func NewAdminDAO() *AdminDAO {
	return &AdminDAO{db: global.GetDB()}
}

func (a *AdminDAO) GetDashboardCounts() (*AdminDashboardCounts, error) {
	var result AdminDashboardCounts
	queries := []struct {
		target *int64
		query  *gorm.DB
	}{
		{&result.BookCount, a.db.Model(&model.Book{})},
		{&result.UserCount, a.db.Model(&model.User{})},
		{&result.OrderCount, a.db.Model(&model.Order{})},
		{&result.LowStockCount, a.db.Model(&model.Book{}).Where("stock < ?", 10)},
		{&result.PendingOrderCount, a.db.Model(&model.Order{}).Where("status = ?", 0)},
	}
	for _, item := range queries {
		if err := item.query.Count(item.target).Error; err != nil {
			return nil, err
		}
	}
	if err := a.db.Model(&model.Order{}).Where("status = ?", 1).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&result.Revenue).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AdminDAO) GetRecentOrders(limit int) ([]*AdminOrder, error) {
	var orders []*AdminOrder
	err := a.db.Model(&model.Order{}).
		Select(`orders.id, orders.order_no, orders.user_id, users.username,
			orders.total_amount, orders.status, orders.is_paid, orders.payment_time,
			orders.created_at, orders.updated_at,
			(SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE order_id = orders.id) AS item_count`).
		Joins("LEFT JOIN users ON users.id = orders.user_id").
		Order("orders.created_at DESC").Limit(limit).Scan(&orders).Error
	return orders, err
}

func (a *AdminDAO) GetTopBooks(limit int) ([]*model.Book, error) {
	var books []*model.Book
	err := a.db.Order("sale DESC, id ASC").Limit(limit).Find(&books).Error
	return books, err
}

func (a *AdminDAO) GetSalesSince(start time.Time) ([]*DailySales, error) {
	var rows []*DailySales
	err := a.db.Model(&model.Order{}).
		Select("DATE(created_at) AS day, COALESCE(SUM(total_amount), 0) AS amount").
		Where("status = ? AND created_at >= ?", 1, start).
		Group("DATE(created_at)").Order("day ASC").Scan(&rows).Error
	return rows, err
}
