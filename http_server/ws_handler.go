package httpServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust/codec"
	"github.com/jxncyjq/stardust/errors"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/uuid"
	"go.uber.org/zap"
)

// WebSocketOptions configures the WebSocket IHandler adapter.
type WebSocketOptions struct {
	Codec       codec.ICodec
	Logger      *zap.Logger
	Manager     IClientManager
	Handler     codec.IMessageProcessor
	UserID      func(*gin.Context) string
	SessionID   func(*gin.Context) string
	CheckOrigin func(*http.Request) bool
}

// WebSocketHandler upgrades HTTP requests and registers clients through the
// framework IHandler contract.
type WebSocketHandler struct {
	name    string
	tags    []string
	options WebSocketOptions
}

// NewWebSocketHandler creates an IHandler for WebSocket routes.
func NewWebSocketHandler(name string, options WebSocketOptions, tags ...string) IHandler {
	return &WebSocketHandler{
		name:    name,
		tags:    tags,
		options: options,
	}
}

func (h *WebSocketHandler) GetName() string {
	return h.name
}

func (h *WebSocketHandler) GetTags() []string {
	return h.tags
}

func (h *WebSocketHandler) GetFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		options := h.options.withDefaults()
		if msg := options.validate(); msg != "" {
			Response(c, errors.New(msg, 50000), nil)
			return
		}

		upgrader := websocket.Upgrader{
			CheckOrigin: options.CheckOrigin,
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			options.Logger.Warn("websocket upgrade failed", logs.ErrorInfo(err))
			return
		}

		client := NewClient(
			options.UserID(c),
			options.SessionID(c),
			conn,
			options.Codec,
			options.Logger,
			c.Request.Context(),
			options.Handler,
			options.Manager,
		)
		options.Manager.RegisterClient(client)
		client.Listen()
	}
}

func (o WebSocketOptions) withDefaults() WebSocketOptions {
	if o.Codec == nil {
		o.Codec = codec.NewJsonCodec()
	}
	if o.Logger == nil {
		o.Logger = logs.GetLogger("websocket")
	}
	if o.UserID == nil {
		o.UserID = func(c *gin.Context) string {
			return c.Query("userId")
		}
	}
	if o.SessionID == nil {
		o.SessionID = func(*gin.Context) string {
			return uuid.GenSessionId()
		}
	}
	if o.CheckOrigin == nil {
		o.CheckOrigin = func(*http.Request) bool {
			return true
		}
	}
	return o
}

func (o WebSocketOptions) validate() string {
	if o.Manager == nil {
		return "websocket manager is required"
	}
	if o.Handler == nil {
		return "websocket message handler is required"
	}
	return ""
}
