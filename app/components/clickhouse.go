package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/clickhouse"
)

type clickhouseComponent struct{}

// ClickhouseComponent 返回 ClickHouse 组件，依赖 logs。
func ClickhouseComponent() app.Component { return &clickhouseComponent{} }

func (c *clickhouseComponent) Name() string          { return "clickhouse" }
func (c *clickhouseComponent) Dependencies() []string { return []string{"logs"} }

func (c *clickhouseComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "clickhouse")
	if err := clickhouse.Init(requireConfig(configFn, "clickhouse")); err != nil {
		return err
	}
	_, err := clickhouse.GetClickHouseManager() // 触发懒初始化，提前暴露连接错误
	return err
}

func (c *clickhouseComponent) Start(_ context.Context) error { return nil }
func (c *clickhouseComponent) Stop(_ context.Context) error  { return nil }
