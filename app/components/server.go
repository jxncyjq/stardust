package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jxncyjq/stardust/app"
	httpServer "github.com/jxncyjq/stardust/http_server"
	"github.com/jxncyjq/stardust/utils"
	"google.golang.org/grpc"
)

// ════════════════════════════════════════════════════════════════
// HTTP 服务器组件
// ════════════════════════════════════════════════════════════════

type httpServerComponent struct {
	setup func(*httpServer.HttpServer)
	srv   *httpServer.HttpServer
}

// HTTPServerComponent 返回 HTTP 服务器组件。
//
// setupFn 在 Init 阶段（所有基础设施组件就绪后）被调用，用于注册中间件和路由：
//
//	components.HTTPServerComponent(func(srv *httpServer.HttpServer) {
//	    srv.AddGroup("v1", middleware.Metrics("svc"), middleware.Access())
//	    srv.Post("hello", "v1", httpServer.NewHandler(...))
//	})
//
// 依赖：logs。tracing 为可选项；使用 middleware.Tracing 时请显式注册 TracingComponent。
func HTTPServerComponent(setupFn func(*httpServer.HttpServer)) app.Component {
	return &httpServerComponent{setup: setupFn}
}

func (c *httpServerComponent) Name() string           { return "http_server" }
func (c *httpServerComponent) Dependencies() []string { return []string{"logs"} }

func (c *httpServerComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "http_server")
	cfg := requireConfig(configFn, "http_server")
	srv, err := httpServer.NewHttpServer(cfg)
	if err != nil {
		return fmt.Errorf("create http server: %w", err)
	}
	if c.setup != nil {
		c.setup(srv) // 注册中间件和路由
	}
	c.srv = srv
	return nil
}

// Start 启动 HTTP 监听（非阻塞）
func (c *httpServerComponent) Start(_ context.Context) error {
	return c.srv.Startup()
}

// Stop 优雅关闭，等待进行中请求处理完毕（内部超时 10s）
func (c *httpServerComponent) Stop(_ context.Context) error {
	c.srv.Stop()
	return nil
}

// ════════════════════════════════════════════════════════════════
// gRPC 服务器组件
// ════════════════════════════════════════════════════════════════

type grpcServerComponent struct {
	setup func(*grpc.Server)
	srv   *httpServer.GrpcServer
}

// GRPCServerComponent 返回 gRPC 服务器组件。
//
// setupFn 在 Init 阶段被调用，用于注册 proto 服务：
//
//	components.GRPCServerComponent(func(s *grpc.Server) {
//	    pb.RegisterUserServiceServer(s, &UserServiceImpl{})
//	})
//
// 依赖：logs、tracing。
// 内置拦截器链由 NewGrpcServer 自动注入（metric → tracing → breaker → timeout）。
func GRPCServerComponent(setupFn func(*grpc.Server)) app.Component {
	return &grpcServerComponent{setup: setupFn}
}

func (c *grpcServerComponent) Name() string           { return "grpc_server" }
func (c *grpcServerComponent) Dependencies() []string { return []string{"logs", "tracing"} }

func (c *grpcServerComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "grpc_server")
	cfg := requireConfig(configFn, "grpc_server")
	conf, err := utils.Bytes2Struct[httpServer.GrpcServerConfig](cfg)
	if err != nil {
		return fmt.Errorf("parse grpc config: %w", err)
	}
	srv, err := httpServer.NewGrpcServer(conf)
	if err != nil {
		return fmt.Errorf("create grpc server: %w", err)
	}
	if c.setup != nil {
		c.setup(srv.Server()) // 注册 proto 服务
	}
	c.srv = srv
	return nil
}

// Start 启动 gRPC 监听（非阻塞）
func (c *grpcServerComponent) Start(_ context.Context) error {
	return c.srv.Startup()
}

// Stop 优雅关闭，等待进行中 RPC 调用完成
func (c *grpcServerComponent) Stop(_ context.Context) error {
	c.srv.Stop()
	return nil
}

// ════════════════════════════════════════════════════════════════
// 与 Application 绑定的 HTTP 服务器组件（自动应用预声明中间件组）
// ════════════════════════════════════════════════════════════════

// boundHTTPServerComponent 在 Init 阶段先将 Application.WithHTTPGroup
// 预声明的中间件组应用到服务器，再调用 setupFn，
// 使 setupFn 只需关注服务构造和路由注册。
type boundHTTPServerComponent struct {
	application *app.Application
	setup       func(*httpServer.HttpServer)
	srv         *httpServer.HttpServer
}

// HTTPServerFromApp 返回与 Application 绑定的 HTTP 服务器组件。
//
// 与 HTTPServerComponent 的区别：
//   - Init 阶段自动将 a.WithHTTPGroup(...) 预声明的中间件组应用到服务器
//   - setupFn 只需构造 Service 并注册路由，无需手动调用 srv.AddGroup
//
// ⚠️ 与 HTTPServerComponent 互斥，同一 Application 只能注册其中一个。
//
// 示例：
//
//	myApp := app.New(conf.Get).
//	    WithHTTPGroup("v1",
//	        middleware.Metrics(appName),
//	        middleware.Tracing(appName),
//	        middleware.CircuitBreaker(),
//	        middleware.Timeout(0),
//	    )
//	myApp.Use(components.HTTPServerFromApp(myApp, setupHTTP)).Run(ctx)
//
//	func setupHTTP(srv *httpServer.HttpServer) {
//	    // "v1" 组已就绪，只需注册路由
//	    srv.Get("user/:id", "v1", httpServer.NewHandler(...))
//	}
func HTTPServerFromApp(a *app.Application, setupFn func(*httpServer.HttpServer)) app.Component {
	return &boundHTTPServerComponent{application: a, setup: setupFn}
}

func (c *boundHTTPServerComponent) Name() string           { return "http_server" }
func (c *boundHTTPServerComponent) Dependencies() []string { return []string{"logs"} }

func (c *boundHTTPServerComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "http_server")
	cfg := requireConfig(configFn, "http_server")
	srv, err := httpServer.NewHttpServer(cfg)
	if err != nil {
		return fmt.Errorf("create http server: %w", err)
	}

	// 自动应用 WithHTTPGroup 预声明的中间件组，setupFn 无需再 AddGroup
	for _, g := range c.application.HTTPGroups() {
		srv.AddGroup(g.Name, g.Middleware...)
	}

	if c.setup != nil {
		c.setup(srv)
	}
	c.srv = srv
	return nil
}

// Start 启动 HTTP 监听（非阻塞）
func (c *boundHTTPServerComponent) Start(_ context.Context) error {
	return c.srv.Startup()
}

// Stop 优雅关闭，等待进行中请求处理完毕
func (c *boundHTTPServerComponent) Stop(_ context.Context) error {
	c.srv.Stop()
	return nil
}

// ════════════════════════════════════════════════════════════════
// 保留：工厂函数（供不使用 app.Component 的场景使用）
// ════════════════════════════════════════════════════════════════

// NewHTTPServerFromConfig 从 configFn("http_server") 读取配置创建 HttpServer。
func NewHTTPServerFromConfig(configFn func(string) []byte) (*httpServer.HttpServer, error) {
	cfg := configFn("http_server")
	if cfg == nil {
		return nil, fmt.Errorf("config key %q not found", "http_server")
	}
	return httpServer.NewHttpServer(cfg)
}

// NewGRPCServerFromConfig 从 configFn("grpc_server") 读取配置创建 GrpcServer。
func NewGRPCServerFromConfig(configFn func(string) []byte) (*httpServer.GrpcServer, error) {
	cfg := configFn("grpc_server")
	if cfg == nil {
		return nil, fmt.Errorf("config key %q not found", "grpc_server")
	}
	var conf httpServer.GrpcServerConfig
	if err := json.Unmarshal(cfg, &conf); err != nil {
		return nil, err
	}
	return httpServer.NewGrpcServer(conf)
}
