---
id: "reference-component-databases-001"
title: "DatabasesComponent 使用说明"
aliases: ["DatabasesComponent", "数据库组件", "GORM组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "databases", "gorm", "lifecycle"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-databases-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-databases-module-001"
    relation: "parent_of"
    path: "../reference-databases-module-001.md"
---

# DatabasesComponent 使用说明

<!-- @section: overview -->
## 概述

`components.DatabasesComponent()` 负责初始化关系型数据库连接管理器。它读取配置 key `databases`，调用 `databases.Init(...)`，随后通过 `databases.GetDatabaseManager()` 提前建立连接并暴露配置错误。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.DatabasesComponent()` |
| 组件名 | `databases` |
| 依赖 | `logs` |
| 配置 key | `databases` |
| Init | 初始化数据库 manager 并建立连接 |
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
        components.DatabasesComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

业务中通过 `databases.GetDatabaseManager()` 获取 manager，再通过 `GetDBDao(name)` 获取对应 DAO。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[[databases]]
name             = "main"
db_type          = "mysql"
master           = "root:password@tcp(127.0.0.1:3306)/main_db?charset=utf8mb4&parseTime=True&loc=Local"
slaves           = ["root:password@tcp(127.0.0.1:3307)/main_db?charset=utf8mb4&parseTime=True&loc=Local"]
use_master_slave = true
max_conn         = 20
max_idle         = 5
show_sql         = false
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-component-logs-001]]
