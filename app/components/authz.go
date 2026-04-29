package components

import (
	"context"
	"encoding/json"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/authz"
	"github.com/jxncyjq/stardust/databases"
)

type authzComponent struct{}

// AuthzComponent 返回授权组件，依赖 logs。
// 当 authz 配置使用 gorm 且指定 database_name 时，组件会自动读取 [databases] 配置并初始化数据库连接。
func AuthzComponent() app.Component { return &authzComponent{} }

func (c *authzComponent) Name() string { return "authz" }

func (c *authzComponent) Dependencies() []string { return []string{"logs"} }

func (c *authzComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "authz")
	configBytes := requireConfig(configFn, "authz")
	var cfg authz.Config
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return err
	}

	if cfg.Adapter == authz.GormAdapter && cfg.DatabaseName != "" {
		if err := databases.Init(requireConfig(configFn, "databases")); err != nil {
			return err
		}
		if _, err := databases.GetDatabaseManager(); err != nil {
			return err
		}
	}

	return authz.Init(configBytes)
}

func (c *authzComponent) Start(_ context.Context) error { return nil }

func (c *authzComponent) Stop(_ context.Context) error { return nil }
