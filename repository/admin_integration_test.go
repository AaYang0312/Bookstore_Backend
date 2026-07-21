package repository

import (
	"bookstore-manager/config"
	"bookstore-manager/global"
	"os"
	"testing"
	"time"
)

func TestAdminReadQueriesAgainstLocalDatabase(t *testing.T) {
	if os.Getenv("BOOKSTORE_INTEGRATION") != "1" {
		t.Skip("设置 BOOKSTORE_INTEGRATION=1 后运行本地数据库集成测试")
	}
	config.InitConfig("../conf/config.yaml")
	global.InitMysql()
	if global.GetDB() == nil {
		t.Fatal("无法连接本地 MySQL")
	}
	t.Cleanup(global.CloseDB)

	adminDAO := NewAdminDAO()
	if _, err := adminDAO.GetDashboardCounts(); err != nil {
		t.Fatalf("查询仪表盘统计失败: %v", err)
	}
	if _, err := adminDAO.GetRecentOrders(5); err != nil {
		t.Fatalf("查询最近订单失败: %v", err)
	}
	if _, err := adminDAO.GetTopBooks(5); err != nil {
		t.Fatalf("查询热销图书失败: %v", err)
	}
	if _, err := adminDAO.GetSalesSince(time.Now().AddDate(0, 0, -6)); err != nil {
		t.Fatalf("查询销售趋势失败: %v", err)
	}
	if _, _, err := NewBookDAO().GetAdminBooks("", nil, 1, 20); err != nil {
		t.Fatalf("查询管理端图书失败: %v", err)
	}
	if _, err := NewCategoryDAO().GetAdminCategories(); err != nil {
		t.Fatalf("查询管理端分类失败: %v", err)
	}
	if _, _, err := NewOrderDAO().GetAdminOrders("", nil, 1, 20); err != nil {
		t.Fatalf("查询管理端订单失败: %v", err)
	}
	if _, _, err := NewUserDAO().GetAdminUsers("", 1, 20); err != nil {
		t.Fatalf("查询管理端用户失败: %v", err)
	}
	if _, err := NewCarouselDAO().GetAdminCarousels(); err != nil {
		t.Fatalf("查询管理端轮播图失败: %v", err)
	}
}
