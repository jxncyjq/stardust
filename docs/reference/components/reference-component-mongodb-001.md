---
id: "reference-component-mongodb-001"
title: "MongoDBComponent 使用说明"
aliases: ["MongoDBComponent", "MongoDB组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "mongodb", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-mongodb-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-mongodb-module-001"
    relation: "parent_of"
    path: "../reference-mongodb-module-001.md"
---

# MongoDBComponent 使用说明

<!-- @section: overview -->
## 概述

`components.MongoDBComponent()` 负责初始化 MongoDB 连接管理器。它读取配置 key `mongodb`，调用 `mongodb.Init(...)`，随后通过 `mongodb.GetMongoManager()` 触发连接初始化。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.MongoDBComponent()` |
| 组件名 | `mongodb` |
| 依赖 | `logs` |
| 配置 key | `mongodb` |
| Init | 初始化 MongoDB manager 并建立连接 |
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
        components.MongoDBComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务中通过 `mongodb.GetMongoManager()` 或 `mongodb.GetClient(name)` 获取客户端。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[mongodb]
name      = "default"
uri       = "mongodb://root:password@127.0.0.1:27017/mydb?authSource=admin"
database  = "mydb"
max_pool  = 10
min_pool  = 2
timeout_s = 5
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
