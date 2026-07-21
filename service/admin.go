package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"time"
)

type AdminService struct {
	AdminDB *repository.AdminDAO
}

type AdminDashboard struct {
	BookCount         int64                    `json:"book_count"`
	UserCount         int64                    `json:"user_count"`
	OrderCount        int64                    `json:"order_count"`
	Revenue           int64                    `json:"revenue"`
	LowStockCount     int64                    `json:"low_stock_count"`
	PendingOrderCount int64                    `json:"pending_order_count"`
	RecentOrders      []*repository.AdminOrder `json:"recent_orders"`
	TopBooks          []*model.Book            `json:"top_books"`
	SalesTrend        []int64                  `json:"sales_trend"`
}

func NewAdminService() *AdminService {
	return &AdminService{AdminDB: repository.NewAdminDAO()}
}

func (a *AdminService) GetDashboard() (*AdminDashboard, error) {
	counts, err := a.AdminDB.GetDashboardCounts()
	if err != nil {
		return nil, err
	}
	recentOrders, err := a.AdminDB.GetRecentOrders(5)
	if err != nil {
		return nil, err
	}
	topBooks, err := a.AdminDB.GetTopBooks(5)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := today.AddDate(0, 0, -6)
	dailySales, err := a.AdminDB.GetSalesSince(start)
	if err != nil {
		return nil, err
	}
	salesByDay := make(map[string]int64, len(dailySales))
	for _, row := range dailySales {
		salesByDay[row.Day.Format("2006-01-02")] = row.Amount
	}
	salesTrend := make([]int64, 7)
	var maxSales int64
	for index := range salesTrend {
		day := start.AddDate(0, 0, index)
		salesTrend[index] = salesByDay[day.Format("2006-01-02")]
		if salesTrend[index] > maxSales {
			maxSales = salesTrend[index]
		}
	}
	// 前端柱状图直接把该值作为高度百分比，因此按七天峰值归一化。
	if maxSales > 0 {
		for index, amount := range salesTrend {
			salesTrend[index] = amount * 100 / maxSales
		}
	}

	return &AdminDashboard{
		BookCount:         counts.BookCount,
		UserCount:         counts.UserCount,
		OrderCount:        counts.OrderCount,
		Revenue:           counts.Revenue,
		LowStockCount:     counts.LowStockCount,
		PendingOrderCount: counts.PendingOrderCount,
		RecentOrders:      recentOrders,
		TopBooks:          topBooks,
		SalesTrend:        salesTrend,
	}, nil
}
