---
id: "reference-component-business-001"
title: "BusinessComponent 使用说明"
aliases: ["BusinessComponent", "Business组件", "业务聚合组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "business", "module", "service"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children: []
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-component-http-server-from-app-001"
    relation: "related_to"
    path: "./reference-component-http-server-from-app-001.md"
  - id: "reference-component-grpc-server-001"
    relation: "related_to"
    path: "./reference-component-grpc-server-001.md"
---

# BusinessComponent 使用说明

<!-- @section: overview -->
## 概述

`components.Business(manager, services...)` 创建 `BusinessComponent`，用于把多个业务模块接入 `app.Component` 生命周期。业务模块至少实现 `service.Service`；如果实现 `SetupHTTP` 或 `SetupGRPC`，组件会自动分发 HTTP/gRPC 注册。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.Business(manager, services...)` |
| 组件名 | `business` |
| 依赖 | 通过 `WithDependencies(...)` 设置 |
| 配置 key | 无 |
| Init | `manager.AddMany(services...)` |
| Start | `manager.StartAsync()` |
| Stop | `manager.Stop()` |
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: usage -->
```go
userModule := &UserModule{}
gameModule := &GameModule{}

businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
).WithDependencies("logs", "redis", "databases")
```
<!-- @end-code -->

配合 HTTP/gRPC server：

<!-- @code: with-server -->
```go
myApp.Use(
    businessComponent,
    components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
    components.GRPCServerComponent(businessComponent.SetupGRPC),
)
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: binders -->
## 绑定接口

HTTP 绑定：

<!-- @code: http-binder -->
```go
func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
    srv.Get("user/:id", "v1", httpServer.NewHandler("user/:id", []string{"user"}, m.GetUser))
}
```
<!-- @end-code -->

gRPC 绑定：

<!-- @code: grpc-binder -->
```go
func (m *UserModule) SetupGRPC(s *grpc.Server) {
    // pb.RegisterUserServiceServer(s, &UserServiceImpl{service: m.Service()})
}
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-http-server-from-app-001]]
- [[reference-component-grpc-server-001]]
