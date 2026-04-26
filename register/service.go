package register

import (
	"context"
	"fmt"

	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/uuid"
	"go.uber.org/zap"
)

// ServiceRegistry 服务注册管理器
type ServiceRegistry struct {
	register Register
	info     *ServiceInfo
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewServiceRegistry 创建服务注册管理器
func NewServiceRegistry(register Register) *ServiceRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceRegistry{
		register: register,
		logger:   logs.GetLogger("service_registry"),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Register 注册服务
func (s *ServiceRegistry) Register(name, address string, port int, tags []string, meta map[string]string) error {
	s.info = &ServiceInfo{
		Name:    name,
		ID:      fmt.Sprintf("%s-%s", name, uuid.GenSessionId()),
		Address: address,
		Port:    port,
		Tags:    tags,
		Meta:    meta,
	}

	if err := s.register.Register(s.ctx, s.info); err != nil {
		s.logger.Error("Failed to register service", zap.Error(err))
		return err
	}

	s.logger.Info("Service registered",
		zap.String("name", s.info.Name),
		zap.String("id", s.info.ID),
		zap.String("address", fmt.Sprintf("%s:%d", address, port)))
	return nil
}

// Deregister 注销服务
func (s *ServiceRegistry) Deregister() error {
	if s.info == nil {
		return nil
	}

	s.cancel()
	if err := s.register.Deregister(context.Background(), s.info.ID); err != nil {
		s.logger.Error("Failed to deregister service", zap.Error(err))
		return err
	}

	s.logger.Info("Service deregistered", zap.String("id", s.info.ID))
	return nil
}

// Discover 服务发现
func (s *ServiceRegistry) Discover(serviceName string) ([]*ServiceInfo, error) {
	return s.register.GetService(s.ctx, serviceName)
}

// Watch 监听服务变化
func (s *ServiceRegistry) Watch(serviceName string) (<-chan []*ServiceInfo, error) {
	return s.register.Watch(s.ctx, serviceName)
}

// Close 关闭
func (s *ServiceRegistry) Close() error {
	s.Deregister()
	return s.register.Close()
}
