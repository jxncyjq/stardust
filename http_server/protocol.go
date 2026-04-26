package httpServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/errors"
)

// 定义了 HTTP 请求的通用结构和一个辅助函数 BindAndValidate 来统一处理请求绑定和验证。
type PageReq struct {
	Page     int    `json:"page"`      //当前面
	PageSize int    `json:"page_size"` // 一页数量
	Sort     string `json:"sort"`      //排序
}

// 返回定义
type BaseResponse struct {
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type BasePageResponse struct {
	ErrCode int      `json:"errCode"`
	ErrMsg  string   `json:"errMsg,omitempty"`
	Data    any      `json:"data,omitempty"`
	Page    PageResp `json:"page"`
}

type PageResp struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Sort     string `json:"sort"`
	Total    int64  `json:"total"`
}

func Response(c *gin.Context, err *errors.StackError, data any) {
	if err == nil {
		if data != nil {
			c.JSON(http.StatusOK, BaseResponse{
				ErrCode: 0,
				ErrMsg:  "操作成功!",
				Data:    data,
			})
			return
		}
		c.JSON(http.StatusOK, BaseResponse{
			ErrCode: 0,
			ErrMsg:  "操作成功!",
		})
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		ErrCode: err.Code(),
		ErrMsg:  err.Msg(),
		Data:    nil,
	})
}
