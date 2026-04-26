package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Subscribe 订阅
// @return *redis.PubSub, error
func (r *redisView) Subscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	switch v := r.cmd.(type) {
	case *redis.Client:
		return v.Subscribe(ctx, channels...), nil
	case interface {
		Subscribe(context.Context, ...string) *redis.PubSub
	}:
		return v.Subscribe(ctx, channels...), nil
	default:
		return nil, errors.New("UnSupported")
	}
}

// PSubscribe 模式订阅
// @return *redis.PubSub, error
func (r *redisView) PSubscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	switch v := r.cmd.(type) {
	case *redis.Client:
		return v.PSubscribe(ctx, channels...), nil
	case interface {
		PSubscribe(context.Context, ...string) *redis.PubSub
	}:
		return v.PSubscribe(ctx, channels...), nil
	default:
		return nil, errors.New("UnSupported")
	}
}

// Publish 发布消息
// @return int64, error
func (r *redisView) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	switch v := r.cmd.(type) {
	case *redis.Client:
		return v.Publish(ctx, channel, message).Result()
	case interface {
		Publish(context.Context, string, interface{}) *redis.IntCmd
	}:
		return v.Publish(ctx, channel, message).Result()
	default:
		return 0, errors.New("UnSupported")
	}
}
