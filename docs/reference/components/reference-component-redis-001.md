---
id: "reference-component-redis-001"
title: "RedisComponent 使用说明"
aliases: ["RedisComponent", "Redis组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "redis", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-redis-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-redis-module-001"
    relation: "parent_of"
    path: "../reference-redis-module-001.md"
---

# RedisComponent 使用说明

<!-- @section: overview -->
## 概述

`components.RedisComponent()` 负责初始化 Redis 连接管理器。它读取配置 key `redis`，调用 `redis.Init(...)`，随后通过 `redis.GetRedisManager()` 触发懒初始化，提前暴露连接错误。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.RedisComponent()` |
| 组件名 | `redis` |
| 依赖 | `logs` |
| 配置 key | `redis` |
| Init | 初始化 Redis manager 并建立连接 |
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
        components.RedisComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务中通过 `redis.GetRedisManager()` 获取 manager，再按实例名获取客户端或 view。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[[redis]]
name       = "default"
addrs      = ["127.0.0.1:6379"]
password   = ""
key_prefix = "my-service"
db_index   = 0
pool_size  = 10
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
