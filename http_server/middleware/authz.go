package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/authz"
)

// Authz 返回授权中间件。
//
// 可选传入一个 authorizer。若不传，则在每个请求中从 authz 包全局实例读取。
func Authz(authorizers ...authz.Authorizer) gin.HandlerFunc {
	var fixed authz.Authorizer
	if len(authorizers) > 0 {
		fixed = authorizers[0]
	}

	return func(c *gin.Context) {
		authorizer := fixed
		if authorizer == nil {
			var err error
			authorizer, err = authz.Get()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"errCode": 2,
					"errMsg":  "授权器未初始化",
				})
				c.Abort()
				return
			}
		}
		if authorizer == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errCode": 2,
				"errMsg":  "授权器未初始化",
			})
			c.Abort()
			return
		}

		subValue, ok := c.Get("id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errCode": 2,
				"errMsg":  "用户未认证",
			})
			c.Abort()
			return
		}
		sub, ok := subValue.(string)
		if !ok || sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errCode": 2,
				"errMsg":  "用户信息无效",
			})
			c.Abort()
			return
		}

		obj := c.FullPath()
		if obj == "" {
			obj = c.Request.URL.Path
		}

		allowed, err := authorizer.Enforce(c.Request.Context(), sub, obj, c.Request.Method)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errCode": 2,
				"errMsg":  "授权检查失败",
			})
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"errCode": 2,
				"errMsg":  "无访问权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
