package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/breaker"
)

var (
	httpBreakers   = make(map[string]breaker.Breaker)
	httpBreakersMu sync.RWMutex
)

func getHttpBreaker(path string) breaker.Breaker {
	httpBreakersMu.RLock()
	b, ok := httpBreakers[path]
	httpBreakersMu.RUnlock()
	if ok {
		return b
	}

	httpBreakersMu.Lock()
	defer httpBreakersMu.Unlock()
	b, ok = httpBreakers[path]
	if ok {
		return b
	}
	b = breaker.NewGoogleBreaker()
	httpBreakers[path] = b
	return b
}

// CircuitBreaker 熔断器中间件（按路径隔离）
func CircuitBreaker() gin.HandlerFunc {
	return func(c *gin.Context) {
		b := getHttpBreaker(c.FullPath())
		promise, err := b.Allow()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code": http.StatusServiceUnavailable,
				"msg":  "service unavailable (circuit breaker open)",
			})
			return
		}

		c.Next()

		if c.Writer.Status() >= http.StatusInternalServerError {
			promise.Reject(nil)
		} else {
			promise.Accept()
		}
	}
}
