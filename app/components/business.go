package components

import (
	"context"
	"sync"

	"github.com/jxncyjq/stardust/app"
	httpServer "github.com/jxncyjq/stardust/http_server"
	"github.com/jxncyjq/stardust/service"
	"google.golang.org/grpc"
)

// HTTPBinder 表示可绑定 HTTP 路由的业务服务。
type HTTPBinder interface {
	SetupHTTP(srv *httpServer.HttpServer)
}

// GRPCBinder 表示可绑定 gRPC 服务的业务服务。
type GRPCBinder interface {
	SetupGRPC(s *grpc.Server)
}

// BusinessComponent 将多个业务服务接入 app.Component 生命周期。
type BusinessComponent struct {
	manager      service.Manager
	services     []service.Service
	dependencies []string
	registerOnce sync.Once
	startOnce    sync.Once
	started      bool
}

// Business 创建业务服务聚合组件。
//
// 业务服务只需实现 service.Service；如果还实现 HTTPBinder / GRPCBinder，
// 组件会在 SetupHTTP / SetupGRPC 阶段自动注册协议入口。
func Business(manager service.Manager, services ...service.Service) *BusinessComponent {
	return (&BusinessComponent{
		manager: manager,
	}).Add(services...)
}

// BusinessServices 创建业务服务聚合组件。
//
// Deprecated: use Business instead.
func BusinessServices(manager service.Manager, services ...service.Service) *BusinessComponent {
	return Business(manager, services...)
}

// Add 添加一个或多个业务服务。
func (c *BusinessComponent) Add(services ...service.Service) *BusinessComponent {
	c.services = append(c.services, services...)
	return c
}

// WithDependencies 设置业务服务组件依赖。
func (c *BusinessComponent) WithDependencies(dependencies ...string) *BusinessComponent {
	c.dependencies = append(c.dependencies[:0], dependencies...)
	return c
}

// Name 返回组件名。
func (c *BusinessComponent) Name() string { return "business" }

// Dependencies 返回业务服务组件依赖。
func (c *BusinessComponent) Dependencies() []string {
	return c.dependencies
}

// Init 将多个业务服务注册到 service.Manager。
func (c *BusinessComponent) Init(_ context.Context, _ app.ConfigFunc) error {
	c.registerServices()
	return nil
}

// registerServices 幂等地把 services 注册到 manager。
func (c *BusinessComponent) registerServices() {
	c.registerOnce.Do(func() {
		c.manager.AddMany(c.services...)
	})
}

// Start 非阻塞启动所有业务服务。
// 若已通过 StartServices 同步启动过，本方法成为 no-op，避免重复 Start。
func (c *BusinessComponent) Start(_ context.Context) error {
	c.startOnce.Do(func() {
		c.manager.StartAsync()
		c.started = true
	})
	return nil
}

// StartServices 同步启动所有业务服务，供 app.WithPreHTTPStart 在 HTTP 路由注册前调用。
//
// 与 Start 的区别：
//   - Start 异步（go svc.Start()），用于已托管在 app.Component 生命周期内的常规启动
//   - StartServices 同步串行调用每个 svc.Start()，确保返回前所有 service 已完成初始化
//     （前提：业务 service 的 Start 方法本身不阻塞）
//
// 重复调用幂等；与 Start 互斥（先调用谁后者即 no-op）。
func (c *BusinessComponent) StartServices() error {
	c.startOnce.Do(func() {
		c.registerServices()
		for _, svc := range c.services {
			svc.Start()
		}
		c.started = true
	})
	return nil
}

// Stop 逆序停止所有业务服务。
func (c *BusinessComponent) Stop(_ context.Context) error {
	c.manager.Stop()
	return nil
}

// SetupHTTP 统一注册所有业务服务的 HTTP 路由。
func (c *BusinessComponent) SetupHTTP(srv *httpServer.HttpServer) {
	for _, svc := range c.services {
		binder, ok := svc.(HTTPBinder)
		if !ok {
			continue
		}
		binder.SetupHTTP(srv)
	}
}

// SetupGRPC 统一注册所有业务服务的 gRPC 服务。
func (c *BusinessComponent) SetupGRPC(s *grpc.Server) {
	for _, svc := range c.services {
		binder, ok := svc.(GRPCBinder)
		if !ok {
			continue
		}
		binder.SetupGRPC(s)
	}
}
