package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/limit"
)

// RateLimit 令牌桶限流中间件
func RateLimit(rate, burst int, key string) gin.HandlerFunc {
	limiter := limit.NewTokenLimiter(rate, burst, key)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "too many requests",
			})
			return
		}
		c.Next()
	}
}

// PeriodRateLimit 滑动窗口限流中间件（按客户端 IP 限流）
func PeriodRateLimit(period, quota int, prefix string) gin.HandlerFunc {
	limiter := limit.NewPeriodLimiter(period, quota, prefix)
	return func(c *gin.Context) {
		key := c.ClientIP()
		status, _ := limiter.Take(key)
		if status == limit.OverQuotaStatus {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
