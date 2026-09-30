package repository

import (
	"bookstore-manager/config"
	"bookstore-manager/global"
	"bookstore-manager/model"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestBrowseLogAndProfileAgainstLocalDatabase 集成测试：浏览记录去重（唯一索引 upsert）、
// 浏览历史分页与画像聚合。设置 BOOKSTORE_INTEGRATION=1 后对本地数据库运行。
func TestBrowseLogAndProfileAgainstLocalDatabase(t *testing.T) {
	if os.Getenv("BOOKSTORE_INTEGRATION") != "1" {
		t.Skip("设置 BOOKSTORE_INTEGRATION=1 后运行本地数据库集成测试")
	}
	config.InitConfig("../conf/config.yaml")
	global.InitMysql()
	if global.GetDB() == nil {
		t.Fatal("无法连接本地 MySQL")
	}
	t.Cleanup(global.CloseDB)

	db := global.GetDB()

	// 准备：临时用户 + 任取一本在架图书
	suffix := time.Now().UnixNano()
	user := &model.User{
		Username: fmt.Sprintf("it_browse_%d", suffix),
		Password: "test-password",
		Email:    fmt.Sprintf("it_browse_%d@test.local", suffix),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建临时用户失败: %v", err)
	}
	cleanup := func() {
		db.Where("user_id = ?", user.ID).Delete(&model.BrowseLog{})
		db.Where("user_id = ?", user.ID).Delete(&model.Favorite{})
		db.Delete(user)
	}
	t.Cleanup(cleanup)

	books, _, err := NewBookDAO().GetBooksByPage(1, 1)
	if err != nil || len(books) == 0 {
		t.Fatalf("获取测试图书失败: %v", err)
	}
	book := books[0]

	browseDAO := NewBrowseLogDAO()
	profileDAO := NewProfileDAO()

	// 1. 埋点去重：同一 user+book 写两次只保留一条，且刷新 viewed_at
	if err := browseDAO.RecordView(user.ID, book.ID); err != nil {
		t.Fatalf("首次记录浏览失败: %v", err)
	}
	time.Sleep(1100 * time.Millisecond) // 保证 viewed_at 可被刷新（秒级精度）
	if err := browseDAO.RecordView(user.ID, book.ID); err != nil {
		t.Fatalf("重复记录浏览失败: %v", err)
	}

	records, total, err := browseDAO.GetByUser(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("查询浏览历史失败: %v", err)
	}
	if total != 1 || len(records) != 1 {
		t.Fatalf("去重失败：期望 1 条记录，实际 total=%d len=%d", total, len(records))
	}
	if records[0].BookID != book.ID || records[0].Title != book.Title {
		t.Fatalf("浏览记录关联图书字段异常: %+v", records[0])
	}

	// 2. 收藏一条数据后校验画像聚合
	if err := NewFavoriteDAO().AddFavorite(user.ID, book.ID); err != nil {
		t.Fatalf("添加临时收藏失败: %v", err)
	}

	favoriteCategories, err := profileDAO.GetFavoriteCategories(user.ID, 5)
	if err != nil {
		t.Fatalf("查询收藏分类分布失败: %v", err)
	}
	if len(favoriteCategories) == 0 {
		t.Fatal("收藏分类分布为空，期望至少 1 项")
	}

	browseCategories, err := profileDAO.GetBrowseTopCategories(user.ID, 5)
	if err != nil {
		t.Fatalf("查询浏览分类分布失败: %v", err)
	}
	if len(browseCategories) == 0 {
		t.Fatal("浏览分类分布为空，期望至少 1 项")
	}

	recentBrowsed, err := profileDAO.GetRecentBrowsed(user.ID, 10)
	if err != nil {
		t.Fatalf("查询最近浏览失败: %v", err)
	}
	if len(recentBrowsed) != 1 || recentBrowsed[0].BookID != book.ID {
		t.Fatalf("最近浏览结果异常: %+v", recentBrowsed)
	}

	// 3. 无订单用户：画像订单部分应为空数组而非报错
	orders, err := profileDAO.GetRecentOrders(user.ID, 10)
	if err != nil {
		t.Fatalf("查询近订单失败: %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("无订单用户期望空订单列表，实际 %d 条", len(orders))
	}

	// 4. 内部同步查询：水位与分页
	syncBooks, syncTotal, nextSince, err := NewBookDAO().GetBooksForSync(nil, 1, 2)
	if err != nil {
		t.Fatalf("查询内部同步数据失败: %v", err)
	}
	if syncTotal < int64(len(syncBooks)) {
		t.Fatalf("同步总数异常: total=%d len=%d", syncTotal, len(syncBooks))
	}
	if len(syncBooks) > 0 && nextSince.IsZero() {
		t.Fatal("next_since 不应为零值")
	}
	if len(syncBooks) > 0 {
		first := syncBooks[0]
		if first.UpdatedAt.After(nextSince) {
			t.Fatalf("next_since(%v) 小于数据内最大 updated_at(%v)", nextSince, first.UpdatedAt)
		}
	}
}
