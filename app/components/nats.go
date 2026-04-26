package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/nats"
)

type natsComponent struct {
	manager *nats.NatsConnManager
}

// NatsComponent 返回 NATS 组件，依赖 logs。
// Init 建立连接，Start 启动连接监控 goroutine，Stop 关闭所有连接。
func NatsComponent() app.Component { return &natsComponent{} }

func (c *natsComponent) Name() string          { return "nats" }
func (c *natsComponent) Dependencies() []string { return []string{"logs"} }

func (c *natsComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "nats")
	if err := nats.Init(requireConfig(configFn, "nats")); err != nil {
		return err
	}
	m, err := nats.GetNatsManager() // 建立所有连接
	if err != nil {
		return err
	}
	c.manager = m
	return nil
}

// Start 在后台启动所有连接的监控循环（非阻塞）。
func (c *natsComponent) Start(_ context.Context) error {
	go c.manager.StartAll()
	return nil
}

func (c *natsComponent) Stop(_ context.Context) error {
	return c.manager.CloseAll()
}
