package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpServer "github.com/jxncyjq/stardust/http_server"
)

// HTTPGroupDef 记录一个预声明的 HTTP 中间件组，供 components.HTTPServerFromApp 读取。
type HTTPGroupDef struct {
	Name       string
	Middleware []httpServer.HandlerFunc
}

// Application 是服务编排入口，提供声明式 API 组合所需组件和服务器。
type Application struct {
	container  *Container
	configFn   ConfigFunc
	httpServer *httpServer.HttpServer
	grpcServer *httpServer.GrpcServer
	httpGroups []HTTPGroupDef // WithHTTPGroup 预声明的中间件组
}

// New 创建 Application，configFn 通常为 conf.Get。
func New(configFn ConfigFunc) *Application {
	return &Application{
		container: &Container{},
		configFn:  configFn,
	}
}

// Use 注册一组组件，返回自身以支持链式调用。
func (a *Application) Use(components ...Component) *Application {
	a.container.Register(components...)
	return a
}

// WithHTTP 设置 HTTP 服务器。
func (a *Application) WithHTTP(srv *httpServer.HttpServer) *Application {
	a.httpServer = srv
	return a
}

// WithGRPC 设置 gRPC 服务器。
func (a *Application) WithGRPC(srv *httpServer.GrpcServer) *Application {
	a.grpcServer = srv
	return a
}

// WithHTTPGroup 预声明一个 HTTP 中间件组。
// 声明的组在 components.HTTPServerFromApp 组件的 Init 阶段自动应用，
// setupFn 无需再调用 srv.AddGroup，只需专注于服务构造和路由注册。
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
func (a *Application) WithHTTPGroup(group string, mw ...httpServer.HandlerFunc) *Application {
	a.httpGroups = append(a.httpGroups, HTTPGroupDef{Name: group, Middleware: mw})
	return a
}

// HTTPGroups 返回通过 WithHTTPGroup 预声明的中间件组列表，供 components.HTTPServerFromApp 读取。
func (a *Application) HTTPGroups() []HTTPGroupDef {
	return a.httpGroups
}

// Run 按 Init → Start → 等信号 → Stop 顺序编排整个服务生命周期。
// 阻塞直到收到 SIGINT/SIGTERM 或 ctx 取消。
func (a *Application) Run(ctx context.Context) error {
	if err := a.container.Init(ctx, a.configFn); err != nil {
		return fmt.Errorf("component init: %w", err)
	}

	if err := a.container.Start(ctx); err != nil {
		return fmt.Errorf("component start: %w", err)
	}

	if a.httpServer != nil {
		if err := a.httpServer.Startup(); err != nil {
			return fmt.Errorf("http server startup: %w", err)
		}
	}
	if a.grpcServer != nil {
		if err := a.grpcServer.Startup(); err != nil {
			return fmt.Errorf("grpc server startup: %w", err)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-ctx.Done():
	}
	signal.Stop(quit)

	if a.grpcServer != nil {
		a.grpcServer.Stop()
	}
	if a.httpServer != nil {
		a.httpServer.Stop()
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return a.container.Stop(stopCtx)
}
