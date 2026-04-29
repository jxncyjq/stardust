---
id: "reference-component-nats-001"
title: "NatsComponent 使用说明"
aliases: ["NatsComponent", "NATS组件", "JetStream组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "nats", "jetstream", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-nats-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-nats-module-001"
    relation: "parent_of"
    path: "../reference-nats-module-001.md"
---

# NatsComponent 使用说明

<!-- @section: overview -->
## 概述

`components.NatsComponent()` 负责初始化 NATS/JetStream 连接管理器。它读取配置 key `nats`，调用 `nats.Init(...)` 和 `nats.GetNatsManager()` 建立连接。`Start` 阶段会后台运行 `StartAll()` 连接监控，`Stop` 阶段调用 `CloseAll()` 关闭所有连接。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.NatsComponent()` |
| 组件名 | `nats` |
| 依赖 | `logs` |
| 配置 key | `nats` |
| Init | 初始化 NATS manager 并建立连接 |
| Start | `go manager.StartAll()` |
| Stop | `manager.CloseAll()` |
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: usage -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.NatsComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务中通过 `nats.GetNatsManager()` 获取 manager，再按实例名获取连接。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[[nats]]
name        = "default"
url         = "nats://127.0.0.1:4222"
use_stream  = true
stream_name = "my-stream"
subjects    = ["my-service.>"]
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
