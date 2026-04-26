package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/logs"
)

type logsComponent struct{}

// LogsComponent 返回日志组件，无依赖，必须最先初始化。
func LogsComponent() app.Component { return &logsComponent{} }

func (c *logsComponent) Name() string          { return "logs" }
func (c *logsComponent) Dependencies() []string { return nil }

func (c *logsComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "logs")
	return logs.Init(requireConfig(configFn, "logs"))
}

func (c *logsComponent) Start(_ context.Context) error { return nil }
func (c *logsComponent) Stop(_ context.Context) error  { return nil }
