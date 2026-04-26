package app

import "context"

// Component 是所有基础设施组件的统一接口。
type Component interface {
	Name() string                                        // 唯一标识，用于依赖声明和日志
	Dependencies() []string                              // 依赖的其他组件名列表
	Init(ctx context.Context, configFn ConfigFunc) error // 初始化，不启动 goroutine
	Start(ctx context.Context) error                     // 启动（非阻塞）
	Stop(ctx context.Context) error                      // 优雅关闭
}

// ConfigFunc 是配置获取函数，key 对应 TOML 顶级 section 名。
// 直接对接现有 conf.Get(key string) []byte。
type ConfigFunc func(key string) []byte
