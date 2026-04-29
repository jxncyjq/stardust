---
id: "guide-app-components-module-001"
title: "app/components/module 最小实践"
aliases: ["app components module", "业务模块最小实践", "BusinessComponent"]
type: "guide"
category: "backend/library"
tags: ["app", "components", "module", "service", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "design-app-component-001"
children:
  - "reference-component-logs-001"
  - "reference-component-redis-001"
  - "reference-component-databases-001"
  - "reference-component-mongodb-001"
  - "reference-component-clickhouse-001"
  - "reference-component-nats-001"
  - "reference-component-tracing-001"
  - "reference-component-http-server-001"
  - "reference-component-grpc-server-001"
  - "reference-component-business-001"
  - "reference-component-authz-001"
  - "reference-component-i18n-001"
related_docs:
  - id: "design-app-component-001"
    relation: "depends_on"
    path: "../design/design-app-component-001.md"
---

# app/components/module 最小实践

<!-- @section: overview -->
## 概述

`example/main.go` 展示了推荐的服务组织方式：基础设施由 `app/components` 统一初始化，业务服务由 `service.Manager` 统一管理，HTTP 和 gRPC 入口由业务模块各自声明，再由 `BusinessComponent` 统一分发。

目标是让 `main` 保持声明式：

<!-- @code: minimal-main -->
```go
userModule := &UserModule{}
gameModule := &GameModule{}
businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
).WithDependencies("logs", "redis", "databases")

app.New(conf.Get).
    WithHTTPGroup("v1", middleware.I18n(), middleware.Metrics(appName), middleware.Tracing(appName), middleware.Authz()).
    Use(
        components.LogsComponent(),
        components.RedisComponent(),
        components.DatabasesComponent(),
        components.I18nComponent(),
        components.AuthzComponent(),
        businessComponent,
        components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
        components.GRPCServerComponent(businessComponent.SetupGRPC),
    ).
    Run(context.Background())
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: relationship -->
## 关系说明

四层关系如下：

```text
app.Application
  └── app.Container
      ├── infrastructure components
      │   ├── LogsComponent
      │   ├── RedisComponent
      │   ├── DatabasesComponent
      │   ├── I18nComponent
      │   ├── AuthzComponent
      │   ├── HTTPServerFromApp
      │   └── GRPCServerComponent
      └── BusinessComponent
          └── service.Manager
              └── service.Service
                  ├── UserModule
                  └── GameModule
```

| 层 | 职责 |
| --- | --- |
| `app.Application` | 应用编排入口，负责组合组件、运行生命周期 |
| `app.Component` | 基础设施或业务聚合组件的统一生命周期接口 |
| `components.Business` | 创建 `BusinessComponent`，把多个业务模块接入 `app.Component`，并统一分发 HTTP/gRPC 注册 |
| `components.I18nComponent` | 创建国际化组件，读取 `i18n` 配置并加载嵌入式翻译文件 |
| `components.AuthzComponent` | 创建授权组件，读取 `authz` 配置并初始化 Casbin 授权器 |
| `service.Manager` | 管理多个 `service.Service` 的启动和停止顺序 |
| `Module` | 业务模块，持有业务 service，声明 HTTP/gRPC 入口 |
| `Service` | 纯业务逻辑实现 |

`BusinessComponent` 不是 `ServiceGroup` 本身。它使用 `service.Manager`，通常传入 `service.NewServiceGroup()`，把业务模块放进统一生命周期中。
<!-- @end-section -->

<!-- @section: module-contract -->
## 业务模块约定

业务模块至少实现 `service.Service`：

<!-- @code: service-contract -->
```go
type Service interface {
    Start()
    Stop()
}
```
<!-- @end-code -->

如果模块需要绑定 HTTP 路由，则实现：

<!-- @code: http-binder -->
```go
func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
    srv.Get("user/:id", "v1", httpServer.NewHandler(
        "user/:id",
        []string{"user"},
        m.GetUser,
    ))
}
```
<!-- @end-code -->

如果模块需要绑定 gRPC 服务，则实现：

<!-- @code: grpc-binder -->
```go
func (m *UserModule) SetupGRPC(s *grpc.Server) {
    // pb.RegisterUserServiceServer(s, &UserServiceImpl{service: m.Service()})
}
```
<!-- @end-code -->

只实现 `service.Service` 的模块也可以被 `BusinessComponent` 管理生命周期；未实现 HTTP/gRPC 绑定接口时，会自动跳过协议注册。
<!-- @end-section -->

<!-- @section: lifecycle -->
## 生命周期

`BusinessComponent` 作为 `app.Component` 被 `app.Container` 调用：

1. `Init`：调用 `service.Manager.AddMany(...)` 注册所有业务模块
2. `Start`：调用 `service.Manager.StartAsync()` 非阻塞启动业务模块
3. `Stop`：调用 `service.Manager.Stop()` 逆序停止业务模块

为什么使用 `StartAsync()`：

- `app.Application.Run` 已经负责信号监听和整体退出流程
- `ServiceGroup.Start()` 会阻塞等待系统信号
- 因此在组件体系内应使用非阻塞启动，避免卡住后续 HTTP/gRPC server 启动
<!-- @end-section -->

<!-- @section: dependency -->
## 依赖声明

业务聚合组件通过 `WithDependencies` 声明基础设施依赖：

<!-- @code: dependencies -->
```go
businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
).WithDependencies("logs", "redis", "databases")
```
<!-- @end-code -->

依赖的意义是让 `app.Container` 在拓扑排序时保证基础设施先完成 `Init`。例如 `UserModule.Service()` 依赖 Redis 和数据库，则必须确保 `RedisComponent`、`DatabasesComponent` 先初始化。

如果某个业务模块不需要 Redis/DB，可以仍放在同一个 `BusinessComponent` 中；依赖声明以整组业务服务的最大依赖集为准。若依赖差异很大，可以拆成多个 `BusinessComponent` 组件。

如果路由需要多语言消息，再在 HTTP 组里增加 `middleware.I18n()`，并在 handler 中使用 `i18n.Localize(c.Request.Context(), ...)` 读取翻译结果。
<!-- @end-section -->

<!-- @section: adding-service -->
## 新增业务服务

新增一个 `OrderModule` 时，只需要三步：

1. 实现 `service.Service`
2. 按需实现 `SetupHTTP` / `SetupGRPC`
3. 加入 `BusinessComponent`

如果业务路由需要授权控制，再额外注册 `components.AuthzComponent()`，并在 `WithHTTPGroup(...)` 中加入 `middleware.Authz()`。认证与授权分离时，`middleware.Access()` 负责把主体信息写入上下文，`middleware.Authz()` 负责做 Casbin 校验。

<!-- @code: add-service -->
```go
orderModule := &OrderModule{}

businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
    orderModule,
).WithDependencies("logs", "redis", "databases")
```
<!-- @end-code -->

不需要修改 `BusinessComponent` 内部结构，也不需要修改 HTTP/gRPC server 组件。
<!-- @end-section -->

<!-- @section: boundaries -->
## 边界建议

- `main` 只负责声明组件和模块，不写业务初始化细节
- `Module` 负责协议绑定和业务 service 的依赖装配
- `Service` 只写业务逻辑，避免直接感知 app/container 生命周期
- `components.Business` 只做聚合和分发，不承载业务逻辑
- `service.ServiceGroup` 只做生命周期管理，不关心 HTTP/gRPC
<!-- @end-section -->

## 相关文档

- [[design-app-component-001]] — app 包组件化统一接口设计
- [[reference-docs-index]] — 文档索引
