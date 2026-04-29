---
id: "reference-component-logs-001"
title: "LogsComponent 使用说明"
aliases: ["LogsComponent", "日志组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "logs", "lifecycle"]
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
---

# LogsComponent 使用说明

<!-- @section: overview -->
## 概述

`components.LogsComponent()` 负责初始化 `logs` 包，是其他多数组件的基础依赖。它读取配置 key `logs`，调用 `logs.Init(...)`。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.LogsComponent()` |
| 组件名 | `logs` |
| 依赖 | 无 |
| 配置 key | `logs` |
| Init | 初始化日志 |
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
    ).
    Run(context.Background())
```
<!-- @end-code -->

`LogsComponent` 应放在需要日志的组件之前注册。即使注册顺序不严格，`app.Container` 也会根据依赖排序保证它先初始化。
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[logs]
filename   = "./logs/app.log"
maxsize    = 100
maxage     = 7
maxbackups = 5
localtime  = true
compress   = false
level      = 0
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[design-app-component-001]]
