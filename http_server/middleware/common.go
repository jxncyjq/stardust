package middleware

import (
	"sync"

	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/redis"
	"go.uber.org/zap"
)

var (
	logger  *zap.Logger
	logOnce sync.Once

	redisOnce sync.Once
	redisView redis.RedisCli
)

func initLogger() {
	logOnce.Do(func() {
		logger = logs.GetLogger("access middleware")
	})
}

func initRedisClient() {
	initLogger()
	redisOnce.Do(func() {
		m, err := redis.GetRedisManager()
		if err != nil {
			logger.Warn("failed to initialize Redis client for middleware, Redis operations will be unavailable", zap.Error(err))
			return
		}
		redisView = m.GetRedisView("default", "middleware", logger)
		if redisView == nil {
			logger.Warn("failed to initialize Redis client for middleware, Redis operations will be unavailable")
		}
	})
}
