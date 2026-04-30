package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/errors"
)

// 定义handler参数的结构
type HandlerParams struct {
	Name    string
	Tags    []string
	Handler interface{}
}

type Handler[Req any, Resp any] struct {
	Path string // 路径
	Name string
	Tags []string
	Func func(*gin.Context, Req) (Resp, error)
}

// 抽象接口
type IHandler interface {
	GetName() string
	GetTags() []string
	GetFunc() gin.HandlerFunc
}

func NewHandler[Req any, Resp any](
	name string,
	tags []string,
	f func(*gin.Context, Req) (Resp, error),
) *Handler[Req, Resp] {
	return &Handler[Req, Resp]{
		Name: name,
		Tags: tags,
		Func: f,
	}
}

func (h *Handler[Req, Resp]) GetName() string {
	return h.Name
}

func (h *Handler[Req, Resp]) GetTags() []string {
	return h.Tags
}

func (h *Handler[Req, Resp]) GetFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBind(&req); err != nil {
			Response(c, errors.New(err.Error(), 40000), nil)
			return
		}
		resp, err := h.Func(c, req)
		if err != nil {
			if se, ok := err.(*errors.StackError); ok {
				Response(c, se, nil)
			} else {
				Response(c, errors.New(err.Error(), 50000), nil)
			}
			return
		}
		Response(c, nil, resp)
	}
}

// 句柄管理器抽象接口
type IHandlers interface {
	GetHandlers() []IHandler
	AddHandlers(handler IHandler)
	GetHandlersLen() int
}

// 句柄管理器
type Handlers struct {
	handlers []IHandler
}

func NewHandlers() IHandlers {
	return &Handlers{
		handlers: make([]IHandler, 0),
	}
}

func (h *Handlers) GetHandlers() []IHandler {
	return h.handlers
}

func (h *Handlers) AddHandlers(handler IHandler) {
	// This method can be overridden to add handlers
	h.handlers = append(h.handlers, handler)
}

func (h *Handlers) GetHandlersLen() int {
	return len(h.handlers)
}

type StarDustGroup struct {
	Prefix string
	Group  *gin.RouterGroup
}

func NewStarDustGroup(prefix string, group *gin.RouterGroup) *StarDustGroup {
	return &StarDustGroup{
		Prefix: prefix,
		Group:  group,
	}
}
