package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"golang.org/x/time/rate"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/errors"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/lru"
)

const (
	capacity = 100
	quantum  = 20
	cap      = 200
)

// UserRateLimiter 针对每个用户的请求进行限速
// TODO 限速大小从配置中读取
func UserRateLimiter() gin.HandlerFunc {
	cache := lru.NewLRUCache(cap)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !cache.Contains(clientIP) {
			cache.Add(clientIP, ratelimit.NewBucketWithQuantum(time.Second, capacity, quantum))
			return
		}
		// 通过 ClientIP 取出 bucket
		val := cache.Get(clientIP)
		if val == nil {
			return
		}

		// 判断是否还有可用的 bucket
		bucket := val.(*ratelimit.Bucket)
		if bucket.TakeAvailable(1) == 0 {
			httputils.AbortFailedWithCode(c, http.StatusForbidden, errors.ErrBusySystem)
		}
	}
}

func Limiter() gin.HandlerFunc {
	// 初始化一个限速器，每秒产生 1000 个令牌，桶的大小为 1000 个
	// 初始化状态桶是满的
	// TODO: 限速的值从配置或者环境变量中获取
	limiter := rate.NewLimiter(1000, 1000)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			httputils.AbortFailedWithCode(c, http.StatusForbidden, errors.ErrBusySystem)
		}
	}
}
