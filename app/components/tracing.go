package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/tracing"
)

type tracingComponent struct {
	tracer *tracing.JaegerTracer
}

// TracingComponent 返回链路追踪组件，依赖 logs。
// Stop 时刷新未发送的 span 并关闭 provider。
func TracingComponent() app.Component { return &tracingComponent{} }

func (c *tracingComponent) Name() string          { return "tracing" }
func (c *tracingComponent) Dependencies() []string { return []string{"logs"} }

func (c *tracingComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "tracing")
	t, err := tracing.NewJaegerTracer(requireConfig(configFn, "tracing"))
	if err != nil {
		return err
	}
	c.tracer = t
	return nil
}

func (c *tracingComponent) Start(_ context.Context) error { return nil }

func (c *tracingComponent) Stop(ctx context.Context) error {
	if c.tracer != nil {
		return c.tracer.Shutdown(ctx)
	}
	return nil
}
