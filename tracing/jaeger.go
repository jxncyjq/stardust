package tracing

import (
	"context"
	"fmt"

	"github.com/jxncyjq/stardust/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	ServiceName string  `json:"service_name" yaml:"service_name"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"` // OTLP HTTP endpoint
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}

// JaegerTracer Jaeger链路追踪实现
type JaegerTracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

// NewJaegerTracer 创建Jaeger追踪器
func NewJaegerTracer(jaegerBytes []byte) (*JaegerTracer, error) {
	config, err := utils.Bytes2Struct[JaegerConfig](jaegerBytes)
	if err != nil {
		return nil, fmt.Errorf("jaeger config error: %w", err)
	}

	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	sampleRate := 1.0
	if config.SampleRate > 0 {
		sampleRate = config.SampleRate
	}

	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &JaegerTracer{
		tracer:   provider.Tracer(config.ServiceName),
		provider: provider,
	}, nil
}

// jaegerSpanContext span上下文实现
type jaegerSpanContext struct {
	span trace.Span
}

func (c *jaegerSpanContext) TraceID() string {
	return c.span.SpanContext().TraceID().String()
}

func (c *jaegerSpanContext) SpanID() string {
	return c.span.SpanContext().SpanID().String()
}

// jaegerSpan span实现
type jaegerSpan struct {
	span trace.Span
}

func (s *jaegerSpan) Context() SpanContext {
	return &jaegerSpanContext{span: s.span}
}

func (s *jaegerSpan) SetTag(key string, value interface{}) Span {
	s.span.SetAttributes(attribute.String(key, toString(value)))
	return s
}

func (s *jaegerSpan) LogFields(fields map[string]interface{}) {
	attrs := make([]attribute.KeyValue, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, attribute.String(k, toString(v)))
	}
	s.span.AddEvent("log", trace.WithAttributes(attrs...))
}

func (s *jaegerSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

func (s *jaegerSpan) Finish() {
	s.span.End()
}

// StartSpan 开始新span
func (t *JaegerTracer) StartSpan(ctx context.Context, operationName string) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, operationName)
	return ctx, &jaegerSpan{span: span}
}

// StartSpanFromParent 从父span创建子span
func (t *JaegerTracer) StartSpanFromParent(ctx context.Context, operationName string) (context.Context, Span) {
	return t.StartSpan(ctx, operationName)
}

// Extract 从载体提取上下文
func (t *JaegerTracer) Extract(ctx context.Context, carrier interface{}) (context.Context, error) {
	if c, ok := carrier.(propagation.TextMapCarrier); ok {
		return otel.GetTextMapPropagator().Extract(ctx, c), nil
	}
	return ctx, nil
}

// Inject 注入上下文到载体
func (t *JaegerTracer) Inject(ctx context.Context, carrier interface{}) error {
	if c, ok := carrier.(propagation.TextMapCarrier); ok {
		otel.GetTextMapPropagator().Inject(ctx, c)
	}
	return nil
}

// Close 关闭
func (t *JaegerTracer) Close() error {
	return t.Shutdown(context.Background())
}

// Shutdown 刷新未发送的 span 并关闭 provider。
func (t *JaegerTracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
