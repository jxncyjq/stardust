package redis

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Location struct {
	Longitude float64
	Latitude  float64
}

func (r *redisView) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) (int64, error) {
	result, err := r.cmd.GeoAdd(ctx, r.expandKey(key), geoLocation...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *redisView) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	result, err := r.cmd.GeoRadius(ctx, r.expandKey(key), longitude, latitude, query).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *redisView) GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	result, err := r.cmd.GeoRadiusByMember(ctx, r.expandKey(key), member, query).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *redisView) GeoDist(ctx context.Context, key string, member1, member2, unit string) (float64, error) {
	result, err := r.cmd.GeoDist(ctx, r.expandKey(key), member1, member2, unit).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *redisView) GeoHash(ctx context.Context, key string, members ...string) ([]string, error) {
	result, err := r.cmd.GeoHash(ctx, r.expandKey(key), members...).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *redisView) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	result, err := r.cmd.GeoPos(ctx, r.expandKey(key), members...).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

/*
	计算两点之间的距离：存入缓存进行计算，计算后的结果返回
	注意：
		1.计算过程的数据，包括key，在计算完成后会清除
		2.key值临时使用，在存入缓存时会加上随机数，保证key不会影响其它缓存数据
*/

func (r *redisView) GeoCalculateDistance(ctx context.Context, key string, location1 Location, location2 Location) (float64, error) {
	//需要生成一个随机数
	key = key + ":" + uuid.New().String()

	defer func() {
		//清除key
		r.Del(context.Background(), key)
	}()

	//当redis中存在相同的key+name的时候会覆盖，存在并发问题
	if _, err := r.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      "location1",
		Longitude: location1.Longitude,
		Latitude:  location1.Latitude,
	}); err != nil {
		return 0, err
	}

	if _, err := r.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      "location2",
		Longitude: location2.Longitude,
		Latitude:  location2.Latitude,
	}); err != nil {
		return 0, err
	}
	dist, err := r.GeoDist(ctx, key, "location1", "location2", "km")
	if err != nil {
		return 0, err
	}

	return dist, nil
}
