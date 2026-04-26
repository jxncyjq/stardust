package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ErrorInputValuesIsNil = errors.New("input values is nil")
var ErrorResultNotOK = errors.New("result is not OK")
var ErrorResultNotTrue = errors.New("result is not true")

const RedisKeySep = ":"

var _ RedisCli = (*redisView)(nil)

type redisView struct {
	prefix string
	cmd    RedisCmd
	pubSub redis.PubSub // nolint:unused
	logger *zap.Logger
}

func (r *redisView) XAdd(ctx context.Context, a redis.XAddArgs) *redis.StringCmd {
	return r.cmd.XAdd(ctx, &a)
}

func NewRedisView(cmd RedisCmd, prefix string, logger *zap.Logger) RedisCli {
	view := &redisView{cmd: cmd, prefix: prefix, logger: logger}
	return view
}

func (r *redisView) KeyPrefix() string {
	return r.prefix
}
func (r *redisView) NativeCmd() RedisCmd {
	return r.cmd
}

func (r *redisView) SetNX(ctx context.Context, key string, value []byte, duration string) (bool, error) {
	if duration != "" {
		timeout, err := time.ParseDuration(duration)
		if nil != err {
			return false, err
		}
		return r.cmd.SetNX(ctx, r.expandKey(key), value, timeout).Result()
	}
	return r.cmd.SetNX(ctx, r.expandKey(key), value, 0).Result()
}

func (r *redisView) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, error) {
	result, _, err := r.cmd.Scan(ctx, cursor, r.expandKey(match), count).Result()
	if nil != err {
		return nil, err
	}
	return result, err
}

func (r *redisView) Get(ctx context.Context, key string) ([]byte, error) {
	if r.logger != nil {
		r.logger.Debug("Redis GET operation",
			zap.String("key", r.expandKey(key)))
	}
	result, err := r.cmd.Get(ctx, r.expandKey(key)).Result()
	if nil != err {
		return nil, err
	}
	return []byte(result), nil
}

func (r *redisView) Set(ctx context.Context, key string, value []byte, duration string) error {
	if r.logger != nil {
		r.logger.Debug("Redis SET operation",
			zap.String("key", r.expandKey(key)),
			zap.String("value", string(value)),
			zap.String("duration", duration))
	}

	if duration != "" {
		timeout, err := time.ParseDuration(duration)
		if nil != err {
			return err
		}
		return r.cmd.Set(ctx, r.expandKey(key), value, timeout).Err()
	}
	return r.cmd.Set(ctx, r.expandKey(key), value, 0).Err()
}

func (r *redisView) Del(ctx context.Context, keys ...string) (int64, error) {
	var all []string
	for _, key := range keys {
		all = append(all, r.expandKey(key))
	}
	return r.cmd.Del(ctx, all...).Result()
}

func (r *redisView) Expire(ctx context.Context, key string, duration string) error {
	timeout, err := time.ParseDuration(duration)
	if nil != err {
		return err
	}
	return wrapResult(func() (interface{}, error) {
		return r.cmd.Expire(ctx, r.expandKey(key), timeout).Result()
	})
}

func (r *redisView) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.cmd.Exists(ctx, r.expandKey(key)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
