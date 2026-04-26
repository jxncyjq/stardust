package service

import (
	"io"

	"go.uber.org/zap"
)

// CleanupService 清理服务，在 ServiceGroup 关闭时执行清理逻辑
// 实现 Service 接口，Start 为空操作，Stop 按逆序执行所有清理函数并关闭 Closer
type CleanupService struct {
	closers  []io.Closer
	cleanups []func()
	logger   *zap.Logger
}

// NewCleanupService 创建清理服务
func NewCleanupService() *CleanupService {
	return &CleanupService{
		logger: getLoggerSafe("cleanup"),
	}
}

// AddCloser 添加需要关闭的资源 (如 ServiceRegistry, ApisixGateway)
func (c *CleanupService) AddCloser(closer io.Closer) {
	c.closers = append(c.closers, closer)
}

// AddCleanup 添加自定义清理函数
func (c *CleanupService) AddCleanup(fn func()) {
	c.cleanups = append(c.cleanups, fn)
}

// Start 实现 Service 接口 (空操作)
func (c *CleanupService) Start() {}

// Stop 实现 Service 接口，逆序执行清理
func (c *CleanupService) Stop() {
	// 先执行自定义清理函数（逆序）
	for i := len(c.cleanups) - 1; i >= 0; i-- {
		c.cleanups[i]()
	}

	// 再关闭 Closer（逆序）
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i].Close(); err != nil {
			if c.logger != nil {
				c.logger.Error("cleanup close error", zap.Error(err))
			}
		}
	}

	if c.logger != nil {
		c.logger.Info("cleanup completed")
	}
}
