package tracing

import (
	"context"
	"errors"
	"testing"
)

// MockSpanContext 模拟span上下文
type MockSpanContext struct {
	traceID string
	spanID  string
}

func (c *MockSpanContext) TraceID() string { return c.traceID }
func (c *MockSpanContext) SpanID() string  { return c.spanID }

// MockSpan 模拟span
type MockSpan struct {
	name   string
	tags   map[string]interface{}
	logs   []map[string]interface{}
	err    error
	ctx    *MockSpanContext
	closed bool
}

func NewMockSpan(name string) *MockSpan {
	return &MockSpan{
		name: name,
		tags: make(map[string]interface{}),
		ctx:  &MockSpanContext{traceID: "trace-123", spanID: "span-456"},
	}
}

func (s *MockSpan) Context() SpanContext { return s.ctx }

func (s *MockSpan) SetTag(key string, value interface{}) Span {
	s.tags[key] = value
	return s
}

func (s *MockSpan) LogFields(fields map[string]interface{}) {
	s.logs = append(s.logs, fields)
}

func (s *MockSpan) SetError(err error) {
	s.err = err
}

func (s *MockSpan) Finish() {
	s.closed = true
}

// MockTracer 模拟追踪器
type MockTracer struct {
	spans []*MockSpan
}

func NewMockTracer() *MockTracer {
	return &MockTracer{}
}

func (t *MockTracer) StartSpan(ctx context.Context, operationName string) (context.Context, Span) {
	span := NewMockSpan(operationName)
	t.spans = append(t.spans, span)
	return ctx, span
}

func (t *MockTracer) StartSpanFromParent(ctx context.Context, operationName string) (context.Context, Span) {
	return t.StartSpan(ctx, operationName)
}

func (t *MockTracer) Extract(ctx context.Context, carrier interface{}) (context.Context, error) {
	return ctx, nil
}

func (t *MockTracer) Inject(ctx context.Context, carrier interface{}) error {
	return nil
}

func (t *MockTracer) Close() error {
	return nil
}

func TestSpanContext(t *testing.T) {
	ctx := &MockSpanContext{traceID: "trace-abc", spanID: "span-def"}

	if ctx.TraceID() != "trace-abc" {
		t.Errorf("Expected traceID 'trace-abc', got '%s'", ctx.TraceID())
	}
	if ctx.SpanID() != "span-def" {
		t.Errorf("Expected spanID 'span-def', got '%s'", ctx.SpanID())
	}
}

func TestMockSpan(t *testing.T) {
	span := NewMockSpan("test-operation")

	// 测试设置标签
	span.SetTag("key1", "value1").SetTag("key2", 123)
	if span.tags["key1"] != "value1" {
		t.Errorf("Expected tag 'key1' = 'value1'")
	}

	// 测试日志
	span.LogFields(map[string]interface{}{"event": "test"})
	if len(span.logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(span.logs))
	}

	// 测试错误
	testErr := errors.New("test error")
	span.SetError(testErr)
	if span.err != testErr {
		t.Errorf("Expected error to be set")
	}

	// 测试结束
	span.Finish()
	if !span.closed {
		t.Errorf("Expected span to be closed")
	}

	// 测试上下文
	ctx := span.Context()
	if ctx.TraceID() != "trace-123" {
		t.Errorf("Expected traceID 'trace-123'")
	}
}

func TestMockTracer(t *testing.T) {
	tracer := NewMockTracer()
	ctx := context.Background()

	// 测试创建span
	ctx1, span1 := tracer.StartSpan(ctx, "operation1")
	if ctx1 == nil {
		t.Error("Expected context not nil")
	}
	if span1 == nil {
		t.Error("Expected span not nil")
	}

	// 测试从父span创建
	_, span2 := tracer.StartSpanFromParent(ctx1, "operation2")
	if span2 == nil {
		t.Error("Expected child span not nil")
	}

	if len(tracer.spans) != 2 {
		t.Errorf("Expected 2 spans, got %d", len(tracer.spans))
	}

	// 测试Extract和Inject
	ctx2, err := tracer.Extract(ctx, nil)
	if err != nil {
		t.Errorf("Extract failed: %v", err)
	}
	if ctx2 == nil {
		t.Error("Expected context from Extract")
	}

	err = tracer.Inject(ctx, nil)
	if err != nil {
		t.Errorf("Inject failed: %v", err)
	}

	// 测试关闭
	err = tracer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestCallItracksAbsInterface(t *testing.T) {
	// 验证MockTracer实现了接口
	var _ CallItracksAbsInterface = (*MockTracer)(nil)
	t.Log("MockTracer implements CallItracksAbsInterface")
}
