package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"bookstore-manager/global"
	"bookstore-manager/repository"
)

// BrowseDedupWindow 同一用户对同一本书的去重窗口，窗口内重复浏览不落库
const BrowseDedupWindow = 10 * time.Minute

type BrowseLogService struct {
	BrowseLogDAO *repository.BrowseLogDAO
	ProfileDAO   *repository.ProfileDAO
}

func NewBrowseLogService() *BrowseLogService {
	return &BrowseLogService{
		BrowseLogDAO: repository.NewBrowseLogDAO(),
		ProfileDAO:   repository.NewProfileDAO(),
	}
}

// dedupKey Redis SETNX 去重键：browse:dedup:{userID}:{bookID}
func dedupKey(userID, bookID int) string {
	return fmt.Sprintf("browse:dedup:%d:%d", userID, bookID)
}

// allowByDedup SETNX 抢占去重键；Redis 故障时放行（宁多记不漏记）
func (s *BrowseLogService) allowByDedup(userID, bookID int) bool {
	if global.RedisClient == nil {
		return true
	}
	ctx := context.Background()
	ok, err := global.RedisClient.SetNX(ctx, dedupKey(userID, bookID), 1, BrowseDedupWindow).Result()
	if err != nil {
		return true
	}
	return ok
}

// RecordView 记录浏览（去重 + upsert），失败只记日志不影响主流程
func (s *BrowseLogService) RecordView(userID, bookID int) {
	if userID <= 0 || bookID <= 0 {
		return
	}
	if !s.allowByDedup(userID, bookID) {
		return
	}
	if err := s.BrowseLogDAO.RecordView(userID, bookID); err != nil {
		log.Printf("记录浏览失败 user=%d book=%d: %v", userID, bookID, err)
	}
}

// GetBrowseHistory 用户浏览记录（分页）
func (s *BrowseLogService) GetBrowseHistory(userID, page, pageSize int) ([]*repository.BrowseRecord, int64, error) {
	return s.BrowseLogDAO.GetByUser(userID, page, pageSize)
}

// AgentProfile 聚合画像响应结构
type AgentProfile struct {
	RecentOrders        []*repository.ProfileOrder       `json:"recent_orders"`
	FavoriteCategories  []repository.CategoryCount       `json:"favorite_categories"`
	BrowseTopCategories []repository.CategoryCount       `json:"browse_top_categories"`
	RecentBrowsed       []*repository.ProfileBrowsedBook `json:"recent_browsed"`
}

// GetAgentProfile 聚合画像：近 10 笔订单、收藏分类分布、浏览 Top5 分类、最近浏览 10 本
// 无数据字段返回空数组而非报错
func (s *BrowseLogService) GetAgentProfile(userID int) (*AgentProfile, error) {
	profile := &AgentProfile{
		RecentOrders:        []*repository.ProfileOrder{},
		FavoriteCategories:  []repository.CategoryCount{},
		BrowseTopCategories: []repository.CategoryCount{},
		RecentBrowsed:       []*repository.ProfileBrowsedBook{},
	}

	orders, err := s.ProfileDAO.GetRecentOrders(userID, 10)
	if err != nil {
		return nil, err
	}
	if len(orders) > 0 {
		profile.RecentOrders = orders
	}

	favoriteCategories, err := s.ProfileDAO.GetFavoriteCategories(userID, 5)
	if err != nil {
		return nil, err
	}
	if len(favoriteCategories) > 0 {
		profile.FavoriteCategories = favoriteCategories
	}

	browseCategories, err := s.ProfileDAO.GetBrowseTopCategories(userID, 5)
	if err != nil {
		return nil, err
	}
	if len(browseCategories) > 0 {
		profile.BrowseTopCategories = browseCategories
	}

	recentBrowsed, err := s.ProfileDAO.GetRecentBrowsed(userID, 10)
	if err != nil {
		return nil, err
	}
	if len(recentBrowsed) > 0 {
		profile.RecentBrowsed = recentBrowsed
	}

	return profile, nil
}