---
id: "reference-component-clickhouse-001"
title: "ClickhouseComponent 使用说明"
aliases: ["ClickhouseComponent", "ClickHouse组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "clickhouse", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-clickhouse-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-clickhouse-module-001"
    relation: "parent_of"
    path: "../reference-clickhouse-module-001.md"
---

# ClickhouseComponent 使用说明

<!-- @section: overview -->
## 概述

`components.ClickhouseComponent()` 负责初始化 ClickHouse 连接管理器。它读取配置 key `clickhouse`，调用 `clickhouse.Init(...)`，随后通过 `clickhouse.GetClickHouseManager()` 触发连接初始化。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.ClickhouseComponent()` |
| 组件名 | `clickhouse` |
| 依赖 | `logs` |
| 配置 key | `clickhouse` |
| Init | 初始化 ClickHouse manager 并建立连接 |
| Start | 无操作 |
| Stop | 无操作 |
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: usage -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.ClickhouseComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务中通过 `clickhouse.GetClickHouseManager()` 获取 manager，再通过 `GetClient(name)` 获取客户端。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[clickhouse]
name           = "default"
addr           = "127.0.0.1:9000"
database       = "mydb"
username       = "default"
password       = ""
max_conn       = 10
max_idle       = 5
dial_timeout_s = 5
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
