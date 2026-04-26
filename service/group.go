package service

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jxncyjq/stardust/logs"
	"go.uber.org/zap"
)

// Service 服务接口（参照 go-zero Service）
type Service interface {
	Start()
	Stop()
}

// ServiceGroup 服务组，统一管理多个服务的生命周期
type ServiceGroup struct {
	services []Service
	stopOnce sync.Once
	stopCh   chan struct{}
	logger   *zap.Logger
}

// NewServiceGroup 创建服务组
func NewServiceGroup() *ServiceGroup {
	return &ServiceGroup{
		services: make([]Service, 0),
		stopCh:   make(chan struct{}),
		logger:   getLoggerSafe("service_group"),
	}
}

// Add 添加服务到组
func (sg *ServiceGroup) Add(svc Service) {
	sg.services = append(sg.services, svc)
}

// Start 启动所有服务并监听退出信号
func (sg *ServiceGroup) Start() {
	sg.startAll()
	sg.waitForSignal()
	sg.Stop()
}

func (sg *ServiceGroup) startAll() {
	for _, svc := range sg.services {
		go svc.Start()
	}
}

// Stop 停止所有服务（逆序关闭，后启动的先关闭）
func (sg *ServiceGroup) Stop() {
	sg.stopOnce.Do(func() {
		close(sg.stopCh)
		for i := len(sg.services) - 1; i >= 0; i-- {
			sg.services[i].Stop()
		}
		if sg.logger != nil {
			sg.logger.Info("all services stopped")
		}
	})
}

func (sg *ServiceGroup) waitForSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		if sg.logger != nil {
			sg.logger.Info("received signal", zap.String("signal", sig.String()))
		}
	case <-sg.stopCh:
	}
}

// getLoggerSafe 安全获取 logger，未初始化时返回 nil
func getLoggerSafe(module string) *zap.Logger {
	defer func() {
		recover()
	}()
	return logs.GetLogger(module)
}
