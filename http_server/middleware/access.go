package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/conf"
	"github.com/jxncyjq/stardust/jwt"
	"github.com/jxncyjq/stardust/redis"
	"github.com/jxncyjq/stardust/utils"
)

// Access 用于判断是否具有访问权限
func Access() gin.HandlerFunc {
	return func(c *gin.Context) {
		//在这里处理拦截请求的逻辑
		initRedisClient()
		jwtStr := c.GetHeader("jwt")

		if jwtStr != "" {
			secret := fmt.Sprintf("%s-%s", conf.GetAppName(), conf.GetAppVersion())
			jwtObj, err := jwt.JWTDecrypt(jwtStr, secret)
			if err != nil || jwtObj == nil || jwtObj["token"] == nil || jwtObj["id"] == nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errCode": 2,
					"errMsg":  "jwt解析错误",
				})
				c.Abort()
				return
			}

			id, ok := jwtObj["id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errCode": 2,
					"errMsg":  "用户信息获取失败",
				})
				c.Abort()
				return
			}
			redisCmd := redis.GetRedisDb()
			key := fmt.Sprintf("%s:user:token:%s", conf.GetAppName(), id)
			oldToken, err1 := redisCmd.Get(context.Background(), key).Result()
			if err1 != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errCode": 2,
					"errMsg":  "获取token失败",
				})
				c.Abort()
				return
			}
			if jwtObj["token"] != oldToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errCode": 2,
					"errMsg":  "账户已经在其他终端登录",
				})
				c.Abort()
				return
			}
			c.Request.URL.Query().Add("id", id)
		} else {
			ip := c.ClientIP()
			// 判断是否内网IP
			ok := utils.IsInnerIp(ip)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errCode": 2,
					"errMsg":  "服务器繁忙,请稍后再试!",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func SetHeaders(headers map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for key, value := range headers {
			c.Header(key, value)
		}
		c.Next()
	}
}
