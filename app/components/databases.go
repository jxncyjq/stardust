package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/databases"
)

type databasesComponent struct{}

// DatabasesComponent 返回关系型数据库组件，依赖 logs。
func DatabasesComponent() app.Component { return &databasesComponent{} }

func (c *databasesComponent) Name() string          { return "databases" }
func (c *databasesComponent) Dependencies() []string { return []string{"logs"} }

func (c *databasesComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "databases")
	if err := databases.Init(requireConfig(configFn, "databases")); err != nil {
		return err
	}
	_, err := databases.GetDatabaseManager() // 触发懒初始化，提前暴露连接错误
	return err
}

func (c *databasesComponent) Start(_ context.Context) error { return nil }
func (c *databasesComponent) Stop(_ context.Context) error  { return nil }
