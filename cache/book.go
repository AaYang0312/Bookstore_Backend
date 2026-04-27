package cache

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

const (
	bookDetailKey = "book:detail:%d"
	bookHotKey    = "book:hot:%d"
	bookNewKey    = "book:new:%d"
	bookListKey   = "book:list:%d:%d"

	bookDetailTTL = 10 * time.Minute
	bookListTTL   = 5 * time.Minute
	nullValueTTL  = 1 * time.Minute
	nullValue     = "null"

	// TTL 随机抖动范围，防止缓存雪崩
	jitterRange = 60 * time.Second
)

type BookCache struct {
	rdb *redis.Client
	ctx context.Context
	sfg singleflight.Group
}

func NewBookCache() *BookCache {
	return &BookCache{
		rdb: global.RedisClient,
		ctx: context.Background(),
	}
}

// jitter 在 TTL 基础上加随机抖动，防止大量 key 同时过期
func jitter(base time.Duration) time.Duration {
	return base + time.Duration(rand.Int63n(int64(jitterRange)))
}

// GetBookDetail 返回 (book, found)
// found=true 且 book=nil 表示缓存了"不存在"（防穿透）
// found=false 表示缓存未命中
func (c *BookCache) GetBookDetail(id int) (*model.Book, bool) {
	val, err := c.rdb.Get(c.ctx, fmt.Sprintf(bookDetailKey, id)).Result()
	if err != nil {
		return nil, false
	}
	if val == nullValue {
		return nil, true
	}
	var book model.Book
	if err := json.Unmarshal([]byte(val), &book); err != nil {
		return nil, false
	}
	return &book, true
}

func (c *BookCache) SetBookDetail(id int, book *model.Book) {
	key := fmt.Sprintf(bookDetailKey, id)
	if book == nil {
		c.rdb.Set(c.ctx, key, nullValue, nullValueTTL)
		return
	}
	data, _ := json.Marshal(book)
	c.rdb.Set(c.ctx, key, data, jitter(bookDetailTTL))
}

// DoWithSingleFlight 用 singleflight 合并对同一 key 的并发 DB 查询
// 防止缓存击穿：热点 key 过期时，只有一个请求穿透到 DB
func (c *BookCache) DoWithSingleFlight(key string, fn func() (any, error)) (any, error) {
	val, err, _ := c.sfg.Do(key, fn)
	return val, err
}

func (c *BookCache) GetHotBooks(limit int) ([]*model.Book, bool) {
	return c.getBookSlice(fmt.Sprintf(bookHotKey, limit))
}

func (c *BookCache) SetHotBooks(limit int, books []*model.Book) {
	c.setBookSlice(fmt.Sprintf(bookHotKey, limit), books, bookListTTL)
}

func (c *BookCache) GetNewBooks(limit int) ([]*model.Book, bool) {
	return c.getBookSlice(fmt.Sprintf(bookNewKey, limit))
}

func (c *BookCache) SetNewBooks(limit int, books []*model.Book) {
	c.setBookSlice(fmt.Sprintf(bookNewKey, limit), books, bookListTTL)
}

type bookListResult struct {
	Books []*model.Book `json:"books"`
	Total int64         `json:"total"`
}

func (c *BookCache) GetBookList(page, pageSize int) ([]*model.Book, int64, bool) {
	val, err := c.rdb.Get(c.ctx, fmt.Sprintf(bookListKey, page, pageSize)).Result()
	if err != nil {
		return nil, 0, false
	}
	var result bookListResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, 0, false
	}
	return result.Books, result.Total, true
}

func (c *BookCache) SetBookList(page, pageSize int, books []*model.Book, total int64) {
	data, _ := json.Marshal(bookListResult{Books: books, Total: total})
	c.rdb.Set(c.ctx, fmt.Sprintf(bookListKey, page, pageSize), data, jitter(bookListTTL))
}

func (c *BookCache) getBookSlice(key string) ([]*model.Book, bool) {
	val, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		return nil, false
	}
	var books []*model.Book
	if err := json.Unmarshal([]byte(val), &books); err != nil {
		return nil, false
	}
	return books, true
}

func (c *BookCache) setBookSlice(key string, books []*model.Book, ttl time.Duration) {
	data, _ := json.Marshal(books)
	c.rdb.Set(c.ctx, key, data, jitter(ttl))
}
