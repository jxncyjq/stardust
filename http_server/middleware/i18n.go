package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	stardi18n "github.com/jxncyjq/stardust/i18n"
)

// I18n 解析请求语言并将 Localizer 注入请求上下文。
func I18n() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := stardi18n.GetConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errCode": 2,
				"errMsg":  "i18n not initialized",
			})
			c.Abort()
			return
		}

		langs := make([]string, 0, 2)
		if lang := c.Query(cfg.QueryKey); lang != "" {
			langs = append(langs, lang)
		}
		if accept := c.GetHeader(cfg.HeaderKey); accept != "" {
			langs = append(langs, accept)
		}
		if len(langs) == 0 {
			langs = append(langs, cfg.DefaultLanguage)
		}

		localizer, err := stardi18n.GetLocalizer(langs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errCode": 2,
				"errMsg":  "i18n not initialized",
			})
			c.Abort()
			return
		}

		c.Set("localizer", localizer)
		c.Request = c.Request.WithContext(stardi18n.WithLocalizer(c.Request.Context(), localizer))
		c.Next()
	}
}
