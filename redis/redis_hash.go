package redis

import "context"

func (r *redisView) HSetNX(ctx context.Context, key, field string, value []byte) error {
	return wrapResult(func() (interface{}, error) {
		return r.cmd.HSetNX(ctx, r.expandKey(key), field, value).Result()
	})
}

func (r *redisView) HSet(ctx context.Context, key, field string, value []byte) error {
	return wrapResult(func() (interface{}, error) {
		return r.cmd.HSet(ctx, r.expandKey(key), field, value).Result()
	})
}

func (r *redisView) HMSet(ctx context.Context, key string, Values map[string][]byte) error {
	if nil == Values {
		return ErrorInputValuesIsNil
	}
	in := make(map[string]string)
	for s, bytes := range Values {
		in[s] = string(bytes)
	}
	return wrapResult(func() (interface{}, error) {
		return r.cmd.HMSet(ctx, r.expandKey(key), in).Result()
	})
}

func (r *redisView) HGet(ctx context.Context, key, field string) ([]byte, error) {
	result, err := r.cmd.HGet(ctx, r.expandKey(key), field).Result()
	return []byte(result), err
}

func (r *redisView) HMGet(ctx context.Context, key string, fields ...string) ([][]byte, error) {
	result, err := r.cmd.HMGet(ctx, r.expandKey(key), fields...).Result()
	if nil != err {
		return nil, err
	}
	var out [][]byte
	for _, i2 := range result {
		out = append(out, i2.([]byte))
	}
	return out, nil
}

func (r *redisView) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	result, err := r.cmd.HGetAll(ctx, r.expandKey(key)).Result()
	if nil != err {
		return nil, err
	}
	out := make(map[string][]byte)
	for s, s2 := range result {
		out[s] = []byte(s2)
	}
	return out, nil
}

func (r *redisView) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return r.cmd.HDel(ctx, r.expandKey(key), fields...).Result()
}

func (r *redisView) HLen(ctx context.Context, key string) (int64, error) {
	return r.cmd.HLen(ctx, r.expandKey(key)).Result()
}

func (r *redisView) HKeys(ctx context.Context, key string) ([]string, error) {
	return r.cmd.HKeys(ctx, r.expandKey(key)).Result()
}

func (r *redisView) HValues(ctx context.Context, key string) ([][]byte, error) {
	return wrapSliceStringToSliceBytes(func() ([]string, error) {
		return r.cmd.HVals(ctx, r.expandKey(key)).Result()
	})
}

func (r *redisView) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.cmd.HExists(ctx, r.expandKey(key), field).Result()
}
