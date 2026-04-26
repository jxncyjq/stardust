package httpServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/utils"
)

type Context struct {
	*gin.Context
	RemoteAddr string
	ClientId   string
	Header     http.Header
}

func NewContext(c *gin.Context) *Context {
	ctx := &Context{
		Context:    c,
		RemoteAddr: utils.GetRemoteAddr(c.Request),
		ClientId:   c.GetHeader(ClientIDKey),
		Header:     c.Request.Header,
	}

	return ctx
}
