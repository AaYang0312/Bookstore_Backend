package cache

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

const (
	categoryActiveKey    = "category:active"
	categoryActiveMinTTL = 10 * time.Minute
	categoryActiveMaxTTL = 30 * time.Minute
)

type CategoryCache struct {
	rdb *redis.Client
	ctx context.Context
	sfg singleflight.Group
}

func NewCategoryCache() *CategoryCache {
	return &CategoryCache{
		rdb: global.RedisClient,
		ctx: context.Background(),
	}
}

func (c *CategoryCache) GetActiveCategories() ([]*model.Category, bool) {
	val, err := c.rdb.Get(c.ctx, categoryActiveKey).Result()
	if err != nil {
		return nil, false
	}
	var categories []*model.Category
	if err := json.Unmarshal([]byte(val), &categories); err != nil {
		return nil, false
	}
	return categories, true
}

func (c *CategoryCache) SetActiveCategories(categories []*model.Category) {
	data, _ := json.Marshal(categories)
	ttl := categoryActiveMinTTL + time.Duration(rand.Int63n(int64(categoryActiveMaxTTL-categoryActiveMinTTL)))
	c.rdb.Set(c.ctx, categoryActiveKey, data, ttl)
}

func (c *CategoryCache) InvalidateActiveCategories() {
	c.rdb.Del(c.ctx, categoryActiveKey)
}

func (c *CategoryCache) DoWithSingleFlight(key string, fn func() (any, error)) (any, error) {
	val, err, _ := c.sfg.Do(key, fn)
	return val, err
}
