package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/redis"
)

type redisComponent struct{}

// RedisComponent 返回 Redis 组件，依赖 logs。
func RedisComponent() app.Component { return &redisComponent{} }

func (c *redisComponent) Name() string          { return "redis" }
func (c *redisComponent) Dependencies() []string { return []string{"logs"} }

func (c *redisComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "redis")
	if err := redis.Init(requireConfig(configFn, "redis")); err != nil {
		return err
	}
	_, err := redis.GetRedisManager() // 触发懒初始化，提前暴露连接错误
	return err
}

func (c *redisComponent) Start(_ context.Context) error { return nil }
func (c *redisComponent) Stop(_ context.Context) error  { return nil }
