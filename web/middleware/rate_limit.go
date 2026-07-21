package middleware

import (
	"bookstore-manager/global"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimitKeyFunc func(ctx *gin.Context) (string, bool)

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])

if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
`)

func RedisRateLimit(
	prefix string,
	limit int64,
	window time.Duration,
	keyFunc RateLimitKeyFunc,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := keyFunc(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "无法识别限流对象",
			})
			c.Abort()
			return
		}

		key := fmt.Sprintf("rate:%s:%s", prefix, subject)
		windowSeconds := int64(window.Seconds())

		count, err := rateLimitScript.Run(
			c.Request.Context(),
			global.RedisClient,
			[]string{key},
			windowSeconds,
		).Int64()

		if err != nil {
			// Redis 故障时暂时放行，避免整个服务不可用。
			// 登录接口也可以根据安全要求改成拒绝请求。
			c.Next()
			return
		}

		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if count > limit {
			c.Header("Retry-After", strconv.FormatInt(windowSeconds, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    -1,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func IPKey(c *gin.Context) (string, bool) {
	ip := c.ClientIP()
	return ip, ip != ""
}

func UserIDKey(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return fmt.Sprint(userID), true
}
