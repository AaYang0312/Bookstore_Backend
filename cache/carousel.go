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
	carouselActiveKey    = "carousel:active"
	carouselActiveMinTTL = 5 * time.Minute
	carouselActiveMaxTTL = 10 * time.Minute
)

type CarouselCache struct {
	rdb *redis.Client
	ctx context.Context
	sfg singleflight.Group
}

func NewCarouselCache() *CarouselCache {
	return &CarouselCache{
		rdb: global.RedisClient,
		ctx: context.Background(),
	}
}

func (c *CarouselCache) GetActiveCarousels() ([]*model.Carousel, bool) {
	val, err := c.rdb.Get(c.ctx, carouselActiveKey).Result()
	if err != nil {
		return nil, false
	}
	var carousels []*model.Carousel
	if err := json.Unmarshal([]byte(val), &carousels); err != nil {
		return nil, false
	}
	return carousels, true
}

func (c *CarouselCache) SetActiveCarousels(carousels []*model.Carousel) {
	data, _ := json.Marshal(carousels)
	ttl := carouselActiveMinTTL + time.Duration(rand.Int63n(int64(carouselActiveMaxTTL-carouselActiveMinTTL)))
	c.rdb.Set(c.ctx, carouselActiveKey, data, ttl)
}

func (c *CarouselCache) InvalidateActiveCarousels() {
	c.rdb.Del(c.ctx, carouselActiveKey)
}

func (c *CarouselCache) DoWithSingleFlight(key string, fn func() (any, error)) (any, error) {
	val, err, _ := c.sfg.Do(key, fn)
	return val, err
}
