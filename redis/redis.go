package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const Nil = redis.Nil

type PubSub = redis.PubSub

type RedisCli interface {
	KeyPrefix() string

	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, error)

	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	SetNX(ctx context.Context, key string, value []byte, duration string) (bool, error)
	Get(ctx context.Context, key string) ([]byte, error)
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Set(ctx context.Context, key string, value []byte, duration string) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, duration string) error
	HSetNX(ctx context.Context, key, field string, value []byte) error
	HSet(ctx context.Context, key, field string, value []byte) error
	HMSet(ctx context.Context, key string, Values map[string][]byte) error
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HMGet(ctx context.Context, key string, fields ...string) ([][]byte, error)
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HLen(ctx context.Context, key string) (int64, error)
	HKeys(ctx context.Context, key string) ([]string, error)
	HValues(ctx context.Context, key string) ([][]byte, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	Exists(ctx context.Context, key string) (bool, error)
	LRem(ctx context.Context, key string, count int64, value []byte) (int64, error)
	LIndex(ctx context.Context, key string, index int64) ([]byte, error)
	LTrim(ctx context.Context, key string, start, stop int64) error
	LSet(ctx context.Context, key string, index int64, value []byte) error
	LPush(ctx context.Context, key string, values ...[]byte) (int64, error)
	LAppend(ctx context.Context, key string, values ...[]byte) (int64, error)
	LPop(ctx context.Context, key string) ([]byte, error)
	LRPop(ctx context.Context, key string) ([]byte, error)
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)
	LLen(ctx context.Context, key string) (int64, error)
	LInsert(ctx context.Context, key string, op InsertOP, pivot, value []byte) (int64, error)
	SLen(ctx context.Context, key string) (int64, error)
	SAdd(ctx context.Context, key string, values ...[]byte) (int64, error)
	SRem(ctx context.Context, key string, values ...[]byte) (int64, error)
	SPop(ctx context.Context, key string) ([]byte, error)
	SPopN(ctx context.Context, key string, count int64) ([][]byte, error)
	SDiff(ctx context.Context, keys ...string) ([][]byte, error)
	SDiffMerge(ctx context.Context, destination string, keys ...string) (int64, error)
	SInter(ctx context.Context, keys ...string) ([][]byte, error)
	SInterMerge(ctx context.Context, destination string, keys ...string) (int64, error)
	SUnion(ctx context.Context, keys ...string) ([][]byte, error)
	SUnionMerge(ctx context.Context, destination string, keys ...string) (int64, error)
	ZLen(ctx context.Context, key string) (int64, error)
	ZCount(ctx context.Context, key string, min, max float64) (int64, error)
	ZLexCount(ctx context.Context, key, min, max string) (int64, error)
	ZAdd(ctx context.Context, key string, members ...*ZMember) (int64, error)
	ZRem(ctx context.Context, key string, members ...*ZMember) (int64, error)
	ZRemRangeByLex(ctx context.Context, key, min, max string) (int64, error)
	ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)

	ZRange(ctx context.Context, key string, start, stop int64, reverse, withScores bool) ([]*ZMember, error)

	ZRangeByScore(ctx context.Context, key string, rangeBy ZRangeBy, reverse, withScores bool) ([]*ZMember, error)

	ZRangeByLex(ctx context.Context, key string, rangeBy ZRangeBy, reverse bool) ([]*ZMember, error)

	ZRank(ctx context.Context, key string, member []byte, reverse bool) (int64, error)

	ZIncr(ctx context.Context, key string, member *ZMember) (float64, error)
	ZIncrNX(ctx context.Context, key string, member *ZMember) (float64, error)

	ZInterMerge(ctx context.Context, destination string, keys []string, wights []float64, aggregate string) (int64, error)
	ZUnionMerge(ctx context.Context, destination string, keys []string, wights []float64, aggregate string) (int64, error)

	GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) (int64, error)
	GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoDist(ctx context.Context, key string, member1, member2, unit string) (float64, error)
	GeoHash(ctx context.Context, key string, members ...string) ([]string, error)
	GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error)
	GeoCalculateDistance(ctx context.Context, key string, location1 Location, location2 Location) (float64, error)

	Subscribe(ctx context.Context, channels ...string) (*redis.PubSub, error)
	PSubscribe(ctx context.Context, channels ...string) (*redis.PubSub, error)
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)

	NativeCmd() RedisCmd
	XAdd(ctx context.Context, a redis.XAddArgs) *redis.StringCmd
}

// RedisEntry represents a key-value pair.
type RedisEntry struct {
	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

// ZMember represents a member of a sorted set.
type ZMember struct {
	Score  float64 `protobuf:"fixed32,1,opt,name=score,proto3" json:"score,omitempty"`
	Member []byte  `protobuf:"bytes,2,opt,name=member,proto3" json:"member,omitempty"`
}

type ZRangeBy struct {
	Min    string `protobuf:"bytes,1,opt,name=min,proto3" json:"min,omitempty"`
	Max    string `protobuf:"bytes,2,opt,name=max,proto3" json:"max,omitempty"`
	Offset int64  `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Count  int64  `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
}

type ZS struct {
	Keys    []string
	Weights []float64
	// Can be SUM, MIN or MAX.
	Aggregate string
}

func (z ZRangeBy) ToRedisRangeBy() redis.ZRangeBy {
	return redis.ZRangeBy{
		Min:    z.Min,
		Max:    z.Max,
		Offset: z.Offset,
		Count:  z.Count,
	}
}

// ZMerge used as an arg to ZInterMerge and ZUnionMerge.
type ZMerge struct {
	Weights   []float64 `protobuf:"fixed64,1,rep,packed,name=Weights,proto3" json:"Weights,omitempty"`
	Aggregate string    `protobuf:"bytes,2,opt,name=Aggregate,proto3" json:"Aggregate,omitempty"` //Can be SUM, MIN or MAX.
}

func (z ZMerge) ToZStore() redis.ZStore {
	return redis.ZStore{
		Weights:   z.Weights,
		Aggregate: z.Aggregate,
	}
}
