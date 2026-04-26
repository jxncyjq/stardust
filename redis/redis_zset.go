package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *redisView) ZLen(ctx context.Context, key string) (int64, error) {
	return r.cmd.ZCard(ctx, r.expandKey(key)).Result()
}

func (r *redisView) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	return r.cmd.ZCount(ctx, r.expandKey(key),
		fmt.Sprintf("%f", min), fmt.Sprintf("%f", max)).Result()
}

func (r *redisView) ZLexCount(ctx context.Context, key, min, max string) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *redisView) ZAdd(ctx context.Context, key string, members ...*ZMember) (int64, error) {
	var zS []redis.Z
	for _, member := range members {
		z := redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
		zS = append(zS, z)
	}
	return r.cmd.ZAdd(ctx, r.expandKey(key), zS...).Result()
}

func (r *redisView) ZRem(ctx context.Context, key string, members ...*ZMember) (int64, error) {
	var zS []interface{}
	for _, member := range members {
		zS = append(zS, member.Member)
	}
	return r.cmd.ZRem(ctx, r.expandKey(key), zS...).Result()
}

func (r *redisView) ZRemRangeByLex(ctx context.Context, key, min, max string) (int64, error) {
	return r.cmd.ZRemRangeByLex(ctx, r.expandKey(key), min, max).Result()
}

func (r *redisView) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error) {
	return r.cmd.ZRemRangeByScore(ctx, r.expandKey(key),
		fmt.Sprintf("%f", min), fmt.Sprintf("%f", max)).Result()
}

func (r *redisView) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return r.cmd.ZRemRangeByRank(ctx, r.expandKey(key), start, stop).Result()
}

func (r *redisView) ZRange(ctx context.Context, key string, start, stop int64, reverse, withScores bool) ([]*ZMember, error) {
	var err error
	var members []string
	var zSlice []redis.Z
	if !reverse && !withScores {
		members, err = r.cmd.ZRange(ctx, r.expandKey(key), start, stop).Result()
	}
	if reverse && withScores {
		zSlice, err = r.cmd.ZRevRangeWithScores(ctx, r.expandKey(key), start, stop).Result()
	}
	if reverse {
		members, err = r.cmd.ZRevRange(ctx, r.expandKey(key), start, stop).Result()
	}
	if withScores {
		zSlice, err = r.cmd.ZRangeWithScores(ctx, r.expandKey(key), start, stop).Result()
	}
	return toRangeZMembers(err, members, zSlice)
}

func (r *redisView) ZRangeByScore(ctx context.Context, key string, rangeBy ZRangeBy, reverse, withScores bool) ([]*ZMember, error) {
	var err error
	var members []string
	var zSlice []redis.Z
	rb := rangeBy.ToRedisRangeBy()
	if !reverse && !withScores {
		members, err = r.cmd.ZRangeByScore(ctx, r.expandKey(key), &rb).Result()
	}
	if reverse && withScores {
		zSlice, err = r.cmd.ZRevRangeByScoreWithScores(ctx, r.expandKey(key), &rb).Result()
	}
	if reverse {
		members, err = r.cmd.ZRevRangeByScore(ctx, r.expandKey(key), &rb).Result()
	}
	if withScores {
		zSlice, err = r.cmd.ZRangeByScoreWithScores(ctx, r.expandKey(key), &rb).Result()
	}
	return toRangeZMembers(err, members, zSlice)
}

func (r *redisView) ZRangeByLex(ctx context.Context, key string, rangeBy ZRangeBy, reverse bool) ([]*ZMember, error) {
	var err error
	var members []string
	rb := rangeBy.ToRedisRangeBy()
	if reverse {
		members, err = r.cmd.ZRevRangeByLex(ctx, key, &rb).Result()
	}
	members, err = r.cmd.ZRangeByLex(ctx, key, &rb).Result()
	return toRangeZMembers(err, members, []redis.Z{})
}

func (r *redisView) ZRank(ctx context.Context, key string, member []byte, reverse bool) (int64, error) {
	if reverse {
		return r.cmd.ZRevRank(ctx, r.expandKey(key), string(member)).Result()
	}
	return r.cmd.ZRank(ctx, r.expandKey(key), string(member)).Result()
}

func (r *redisView) ZIncr(ctx context.Context, key string, member *ZMember) (float64, error) {
	//return r.cmd.ZIncr(ctx, r.expandKey(key), &redis.Z{Score: member.Score, Member: member.Member}).Result()
	return r.cmd.ZIncrBy(ctx, r.expandKey(key), member.Score, string(member.Member)).Result()
}

func (r *redisView) ZIncrNX(ctx context.Context, key string, member *ZMember) (float64, error) {
	zs := []redis.Z{
		{
			Score:  member.Score,
			Member: member.Member,
		},
	}
	return r.cmd.ZAddArgsIncr(ctx, key, redis.ZAddArgs{
		NX:      true,
		Members: zs,
	}).Result()
}

func (r *redisView) ZInterMerge(ctx context.Context, destination string, keys []string, wights []float64, aggregate string) (int64, error) {

	return r.cmd.ZInterStore(ctx, destination, &redis.ZStore{Keys: keys, Weights: wights, Aggregate: aggregate}).Result()
}

func (r *redisView) ZUnionMerge(ctx context.Context, destination string, keys []string, wights []float64, aggregate string) (int64, error) {
	return r.cmd.ZUnionStore(ctx, destination, &redis.ZStore{Keys: keys, Weights: wights, Aggregate: aggregate}).Result()
}
