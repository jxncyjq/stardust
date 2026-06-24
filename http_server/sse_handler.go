package httpServer

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/errors"
)

// SSESendFunc sends one SSE event to the client.
// event may be empty (omits the "event:" line; browser treats it as "message").
// data is JSON-encoded unless already a string or []byte.
// Returns the context error when the client has disconnected.
type SSESendFunc func(event string, data any) error

// SSEOptions configures the SSE IHandler adapter.
type SSEOptions struct {
	// Handler is called once per client connection. It must block—looping and
	// calling send—and return only when done or c.Request.Context() is cancelled.
	Handler func(c *gin.Context, send SSESendFunc)

	// RetryMs instructs the browser to reconnect after this many milliseconds on
	// disconnect. 0 omits the retry directive entirely.
	RetryMs int
}

// SSEHandler streams Server-Sent Events and implements IHandler.
type SSEHandler struct {
	name    string
	tags    []string
	options SSEOptions
}

// NewSSEHandler creates an IHandler for SSE routes.
func NewSSEHandler(name string, options SSEOptions, tags ...string) IHandler {
	return &SSEHandler{name: name, tags: tags, options: options}
}

func (h *SSEHandler) GetName() string  { return h.name }
func (h *SSEHandler) GetTags() []string { return h.tags }

func (h *SSEHandler) GetFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.options.Handler == nil {
			Response(c, errors.New("sse handler is required", 50000), nil)
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // disable nginx proxy buffering

		w := c.Writer
		if h.options.RetryMs > 0 {
			fmt.Fprintf(w, "retry: %d\n\n", h.options.RetryMs)
			w.Flush()
		}

		send := func(event string, data any) error {
			if err := c.Request.Context().Err(); err != nil {
				return err
			}
			if event != "" {
				fmt.Fprintf(w, "event: %s\n", event)
			}
			var payload string
			switch v := data.(type) {
			case string:
				payload = v
			case []byte:
				payload = string(v)
			default:
				b, err := json.Marshal(data)
				if err != nil {
					return err
				}
				payload = string(b)
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)
			w.Flush()
			return nil
		}

		h.options.Handler(c, send)
	}
}
