package register

import "context"

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Tags    []string `json:"tags,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Register 服务注册接口
type Register interface {
	// Register 注册服务
	Register(ctx context.Context, info *ServiceInfo) error
	// Deregister 注销服务
	Deregister(ctx context.Context, serviceID string) error
	// GetService 获取服务实例列表
	GetService(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	// Watch 监听服务变化
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)
	// Close 关闭连接
	Close() error
}
