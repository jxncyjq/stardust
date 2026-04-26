package limit

import (
	"context"
	"sync"
	"time"

	"github.com/jxncyjq/stardust/redis"
)

const periodLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCR", key)
if current == 1 then
    redis.call("expire", key, window)
end
if current > limit then
    return 0
end
return 1
`

// LimitStatus 限流状态
type LimitStatus int

const (
	AllowedStatus   LimitStatus = 1  // 允许
	OverQuotaStatus LimitStatus = 0  // 超出配额
	UnknownStatus   LimitStatus = -1 // 未知（Redis 错误）
)

// PeriodLimiter 滑动窗口限流器
type PeriodLimiter struct {
	period int    // 时间窗口（秒）
	quota  int    // 窗口内最大请求数
	prefix string // Redis key 前缀

	// 本地 fallback
	localMu    sync.Mutex
	localCount map[string]*localCounter
}

type localCounter struct {
	count    int
	expireAt time.Time
}

// NewPeriodLimiter 创建滑动窗口限流器
func NewPeriodLimiter(period, quota int, prefix string) *PeriodLimiter {
	pl := &PeriodLimiter{
		period:     period,
		quota:      quota,
		prefix:     prefix,
		localCount: make(map[string]*localCounter),
	}
	go pl.cleanupLoop()
	return pl
}

// cleanupLoop 定期清理已过期的本地计数条目，防止 map 无限增长。
func (pl *PeriodLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Duration(pl.period*2) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		pl.localMu.Lock()
		for k, c := range pl.localCount {
			if now.After(c.expireAt) {
				delete(pl.localCount, k)
			}
		}
		pl.localMu.Unlock()
	}
}

// Take 消耗一次配额，返回限流状态
func (pl *PeriodLimiter) Take(key string) (LimitStatus, error) {
	rdb := redis.GetRedisDb()
	if rdb == nil {
		return pl.localTake(key), nil
	}

	ctx := context.Background()
	fullKey := pl.prefix + ":" + key
	result, err := rdb.Eval(ctx, periodLimitScript,
		[]string{fullKey},
		pl.quota, pl.period,
	).Int64()

	if err != nil {
		return pl.localTake(key), nil
	}

	if result == 1 {
		return AllowedStatus, nil
	}
	return OverQuotaStatus, nil
}

func (pl *PeriodLimiter) localTake(key string) LimitStatus {
	pl.localMu.Lock()
	defer pl.localMu.Unlock()

	now := time.Now()
	counter, exists := pl.localCount[key]
	if !exists || now.After(counter.expireAt) {
		pl.localCount[key] = &localCounter{
			count:    1,
			expireAt: now.Add(time.Duration(pl.period) * time.Second),
		}
		return AllowedStatus
	}

	counter.count++
	if counter.count > pl.quota {
		return OverQuotaStatus
	}
	return AllowedStatus
}
