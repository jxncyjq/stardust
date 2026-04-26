package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Request 记录请求与响应日志。
func Request() gin.HandlerFunc {
	initLogger()
	return func(c *gin.Context) {
		requestBody := ""
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			requestBody = "failed to read request body"
		} else {
			var jsonData interface{}
			if err := json.Unmarshal(b, &jsonData); err != nil {
				requestBody = string(b)
			} else {
				formatted, _ := json.MarshalIndent(jsonData, "", "  ")
				requestBody = string(formatted)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		}

		host := c.Request.Host
		uri := c.Request.RequestURI
		method := c.Request.Method
		agent := c.Request.UserAgent()
		ip := c.ClientIP()

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		spanCtx := trace.SpanFromContext(c.Request.Context()).SpanContext()
		traceID := spanCtx.TraceID().String()
		spanID := spanCtx.SpanID().String()

		logger.Info("request",
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("method", method),
			zap.String("uri", uri),
			zap.Int("status", blw.Status()),
			zap.String("ip", ip),
			zap.String("agent", agent),
			zap.String("host", host),
			zap.String("body", requestBody),
		)
		logger.Info("response",
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("response", blw.body.String()),
		)
	}
}
