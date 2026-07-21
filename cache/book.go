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
	bookCacheVersionKey = "book:cache:version"
	bookDetailKey       = "book:detail:%d"
	bookHotKey          = "book:hot:v%d:%d"
	bookNewKey          = "book:new:v%d:%d"
	bookListKey         = "book:list:v%d:%d:%d"

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

func (c *BookCache) GetHotBooks(limit int) ([]*model.Book, int64, bool) {
	version := c.collectionVersion()
	books, found := c.getBookSlice(hotBooksCacheKey(version, limit))
	return books, version, found
}

func (c *BookCache) SetHotBooks(limit int, version int64, books []*model.Book) {
	if c.collectionVersion() != version {
		return
	}
	c.setBookSlice(hotBooksCacheKey(version, limit), books, bookListTTL)
}

func (c *BookCache) GetNewBooks(limit int) ([]*model.Book, int64, bool) {
	version := c.collectionVersion()
	books, found := c.getBookSlice(newBooksCacheKey(version, limit))
	return books, version, found
}

func (c *BookCache) SetNewBooks(limit int, version int64, books []*model.Book) {
	if c.collectionVersion() != version {
		return
	}
	c.setBookSlice(newBooksCacheKey(version, limit), books, bookListTTL)
}

type bookListResult struct {
	Books []*model.Book `json:"books"`
	Total int64         `json:"total"`
}

func (c *BookCache) GetBookList(page, pageSize int) ([]*model.Book, int64, int64, bool) {
	version := c.collectionVersion()
	val, err := c.rdb.Get(c.ctx, bookListCacheKey(version, page, pageSize)).Result()
	if err != nil {
		return nil, 0, version, false
	}
	var result bookListResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, 0, version, false
	}
	return result.Books, result.Total, version, true
}

func (c *BookCache) SetBookList(page, pageSize int, version int64, books []*model.Book, total int64) {
	if c.collectionVersion() != version {
		return
	}
	data, _ := json.Marshal(bookListResult{Books: books, Total: total})
	c.rdb.Set(c.ctx, bookListCacheKey(version, page, pageSize), data, jitter(bookListTTL))
}

func hotBooksCacheKey(version int64, limit int) string {
	return fmt.Sprintf(bookHotKey, version, limit)
}

func newBooksCacheKey(version int64, limit int) string {
	return fmt.Sprintf(bookNewKey, version, limit)
}

func bookListCacheKey(version int64, page, pageSize int) string {
	return fmt.Sprintf(bookListKey, version, page, pageSize)
}

// collectionVersion 返回当前图书集合缓存版本。版本键不存在或 Redis
// 暂时不可用时使用 0；首次失效时 INCR 会将版本推进到 1。
func (c *BookCache) collectionVersion() int64 {
	version, err := c.rdb.Get(c.ctx, bookCacheVersionKey).Int64()
	if err != nil {
		return 0
	}
	return version
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

// InvalidateBook 删除单本详情缓存并推进集合缓存版本。旧版本的列表、
// 热销和新书缓存不再被读取，等待 TTL 自动过期，无需扫描 Redis Key。
func (c *BookCache) InvalidateBook(id int) {
	_, _ = c.rdb.TxPipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(c.ctx, bookCacheVersionKey)
		pipe.Del(c.ctx, fmt.Sprintf(bookDetailKey, id))
		return nil
	})
}
