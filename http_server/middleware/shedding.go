package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/load"
)

// AdaptiveShedding 自适应降载中间件
func AdaptiveShedding(opts ...load.Option) gin.HandlerFunc {
	shedder := load.NewAdaptiveShedder(opts...)
	return func(c *gin.Context) {
		promise, err := shedder.Allow()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code": http.StatusServiceUnavailable,
				"msg":  "service overloaded, please retry later",
			})
			return
		}

		c.Next()

		if c.Writer.Status() >= http.StatusInternalServerError {
			promise.Fail()
		} else {
			promise.Pass()
		}
	}
}
