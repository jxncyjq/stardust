package limit

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/jxncyjq/stardust/redis"
)

// Redis Lua 脚本: 令牌桶算法
const tokenLimitScript = `
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity / rate
local ttl = math.floor(fill_time * 2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end
local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end
local delta = math.max(0, now - last_refreshed)
local filled_tokens = math.min(capacity, last_tokens + (delta * rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end
redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)
if allowed then
    return 1
else
    return 0
end
`

// TokenLimiter 令牌桶限流器（参照 go-zero core/limit）
type TokenLimiter struct {
	rate  int    // 每秒生成的令牌数
	burst int    // 桶容量
	key   string // Redis key

	// 本地 fallback（Redis 不可用时使用）
	localMu     sync.Mutex
	localTokens float64
	localLast   time.Time
}

// NewTokenLimiter 创建令牌桶限流器
func NewTokenLimiter(rate, burst int, key string) *TokenLimiter {
	return &TokenLimiter{
		rate:        rate,
		burst:       burst,
		key:         key,
		localTokens: float64(burst),
		localLast:   time.Now(),
	}
}

// Allow 判断是否允许 1 个请求
func (tl *TokenLimiter) Allow() bool {
	return tl.AllowN(time.Now(), 1)
}

// AllowN 判断是否允许 n 个请求
func (tl *TokenLimiter) AllowN(now time.Time, n int) bool {
	return tl.reserveN(now, n)
}

func (tl *TokenLimiter) reserveN(now time.Time, n int) bool {
	rdb := redis.GetRedisDb()
	if rdb == nil {
		return tl.localAllow(now, n)
	}

	ctx := context.Background()
	tokenKey := tl.key + ":tokens"
	tsKey := tl.key + ":ts"
	nowSec := strconv.FormatFloat(float64(now.Unix()), 'f', -1, 64)

	result, err := rdb.Eval(ctx, tokenLimitScript,
		[]string{tokenKey, tsKey},
		tl.rate, tl.burst, nowSec, n,
	).Int64()

	if err != nil {
		return tl.localAllow(now, n)
	}

	return result == 1
}

// localAllow 本地令牌桶 fallback
func (tl *TokenLimiter) localAllow(now time.Time, n int) bool {
	tl.localMu.Lock()
	defer tl.localMu.Unlock()

	elapsed := now.Sub(tl.localLast).Seconds()
	tl.localTokens += elapsed * float64(tl.rate)
	if tl.localTokens > float64(tl.burst) {
		tl.localTokens = float64(tl.burst)
	}
	tl.localLast = now

	if tl.localTokens >= float64(n) {
		tl.localTokens -= float64(n)
		return true
	}
	return false
}
