---
id: "reference-component-http-server-001"
title: "HTTPServerComponent 使用说明"
aliases: ["HTTPServerComponent", "HTTP服务器组件", "Gin组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "http", "gin", "server"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-30"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-component-http-server-from-app-001"
  - "reference-http-server-api-usage-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-component-http-server-from-app-001"
    relation: "related_to"
    path: "./reference-component-http-server-from-app-001.md"
  - id: "reference-http-server-api-usage-001"
    relation: "related_to"
    path: "../reference-http-server-api-usage-001.md"
  - id: "reference-metric-module-001"
    relation: "related_to"
    path: "../reference-metric-module-001.md"
  - id: "reference-uuid-module-001"
    relation: "related_to"
    path: "../reference-uuid-module-001.md"
---

# HTTPServerComponent 使用说明

<!-- @section: overview -->
## 概述

`components.HTTPServerComponent(setupFn)` 负责创建并启动 `http_server.HttpServer`。它读取配置 key `http_server`，在 `Init` 阶段创建 server 并调用 `setupFn` 注册中间件组和路由。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.HTTPServerComponent(setupFn)` |
| 组件名 | `http_server` |
| 依赖 | `logs` |
| 配置 key | `http_server` |
| Init | 创建 HTTP server，执行路由注册函数 |
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
        components.HTTPServerComponent(func(srv *httpServer.HttpServer) {
            srv.AddGroup("v1", middleware.Metrics("my-service"))
            srv.Get("health", "", httpServer.NewHandler("health", nil, healthHandler))
        }),
    ).
    Run(context.Background())
```
<!-- @end-code -->

`TracingComponent` 对 HTTP 服务是可选项；只有挂载 `middleware.Tracing(...)` 或需要全局 tracer provider 时才显式注册。

HTTP 路由便捷方法覆盖 `Get`、`Post`、`Put`、`Patch`、`Delete`、`Head`、`Options`、`Connect`、`Trace`；其它自定义方法可继续使用 `Handle(method, path, handler)`。

WebSocket 路由使用 `httpServer.NewWebSocketHandler(...)` 创建 `IHandler`，再通过 `srv.Get(...)` 注册，避免业务代码直接依赖 Gin 原生路由。

如果希望中间件组集中声明在 `app.Application`，优先使用 [[reference-component-http-server-from-app-001]]。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[http_server]
port        = 8080
address     = "0.0.0.0"
cors        = true
request_log = true
access      = false
mode        = "gin"
worker_id   = 1
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[reference-component-http-server-from-app-001]]
- [[reference-http-server-api-usage-001]]
- [[reference-component-tracing-001]]
