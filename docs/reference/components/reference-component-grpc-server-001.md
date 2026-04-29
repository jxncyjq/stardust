---
id: "reference-component-grpc-server-001"
title: "GRPCServerComponent 使用说明"
aliases: ["GRPCServerComponent", "gRPC服务器组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "grpc", "server"]
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
  - id: "reference-register-module-001"
    relation: "related_to"
    path: "../reference-register-module-001.md"
---

# GRPCServerComponent 使用说明

<!-- @section: overview -->
## 概述

`components.GRPCServerComponent(setupFn)` 负责创建并启动 `http_server.GrpcServer`。它读取配置 key `grpc_server`，在 `Init` 阶段创建 server 并调用 `setupFn` 注册 protobuf 服务。底层 `NewGrpcServer` 会注入 metric、tracing、breaker、timeout 拦截器链。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.GRPCServerComponent(setupFn)` |
| 组件名 | `grpc_server` |
| 依赖 | `logs`, `tracing` |
| 配置 key | `grpc_server` |
| Init | 创建 gRPC server，执行 protobuf 注册函数 |
| Start | `srv.Startup()` |
| Stop | `srv.Stop()` |
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: usage -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.TracingComponent(),
        components.GRPCServerComponent(func(s *grpc.Server) {
            pb.RegisterUserServiceServer(s, &UserServiceImpl{})
        }),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务模块也可以实现 `SetupGRPC(*grpc.Server)`，再通过 [[reference-component-business-001]] 统一注册。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[grpc_server]
listen_on = "0.0.0.0:9090"
timeout   = 5000
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[reference-component-business-001]]
- [[reference-component-tracing-001]]
