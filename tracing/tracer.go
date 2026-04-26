package tracing

import (
	"context"
)

// SpanContext span上下文
type SpanContext interface {
	TraceID() string
	SpanID() string
}

// Span 链路追踪span
type Span interface {
	// Context 获取span上下文
	Context() SpanContext
	// SetTag 设置标签
	SetTag(key string, value interface{}) Span
	// LogFields 记录日志
	LogFields(fields map[string]interface{})
	// SetError 设置错误
	SetError(err error)
	// Finish 结束span
	Finish()
}

// CallItracksAbsInterface 链路追踪接口
type CallItracksAbsInterface interface {
	// StartSpan 开始一个新的span
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)
	// StartSpanFromParent 从父span创建子span
	StartSpanFromParent(ctx context.Context, operationName string) (context.Context, Span)
	// Extract 从载体中提取span上下文
	Extract(ctx context.Context, carrier interface{}) (context.Context, error)
	// Inject 将span上下文注入载体
	Inject(ctx context.Context, carrier interface{}) error
	// Close 关闭tracer
	Close() error
}
