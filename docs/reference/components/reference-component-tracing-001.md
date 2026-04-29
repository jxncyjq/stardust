---
id: "reference-component-tracing-001"
title: "TracingComponent 使用说明"
aliases: ["TracingComponent", "链路追踪组件", "Jaeger组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "tracing", "jaeger", "opentelemetry"]
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
  - id: "reference-microservice-module-001"
    relation: "related_to"
    path: "../reference-microservice-module-001.md"
---

# TracingComponent 使用说明

<!-- @section: overview -->
## 概述

`components.TracingComponent()` 负责初始化 OpenTelemetry/Jaeger 链路追踪。它读取配置 key `tracing`，调用 `tracing.NewJaegerTracer(...)`。`Stop` 阶段会调用 `Shutdown(ctx)` 刷新未发送的 span 并关闭 provider。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.TracingComponent()` |
| 组件名 | `tracing` |
| 依赖 | `logs` |
| 配置 key | `tracing` |
| Init | 初始化 Jaeger tracer |
| Start | 无操作 |
| Stop | `tracer.Shutdown(ctx)` |
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: usage -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.TracingComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

HTTP/gRPC server 组件依赖 `tracing`，如果使用它们，通常同时注册 `TracingComponent`。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[tracing]
service_name = "my-service"
endpoint     = "127.0.0.1:4318"
sample_rate  = 1.0
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
