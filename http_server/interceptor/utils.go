package interceptor

import (
	"github.com/jxncyjq/stardust/logs"
	"go.uber.org/zap"
)

// getLoggerSafe 安全获取 logger，未初始化时返回 nil
func getLoggerSafe(module string) *zap.Logger {
	defer func() {
		recover()
	}()
	return logs.GetLogger(module)
}
