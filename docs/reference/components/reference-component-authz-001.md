---
id: "reference-component-authz-001"
title: "AuthzComponent 使用说明"
aliases: ["AuthzComponent", "授权组件", "Casbin组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "authz", "casbin", "authorization"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-authz-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-authz-module-001"
    relation: "parent_of"
    path: "../reference-authz-module-001.md"
---

# AuthzComponent 使用说明

<!-- @section: overview -->
## 概述

`components.AuthzComponent()` 负责初始化 `authz` 模块。它读取配置 key `authz`，调用 `authz.Init(...)`，把 Casbin 授权器注册到包级全局实例中，供 `middleware.Authz()` 和业务代码复用。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.AuthzComponent()` |
| 组件名 | `authz` |
| 依赖 | `logs` |
| 配置 key | `authz` |
| Init | 读取 `authz` 配置并初始化 Casbin |
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
        components.AuthzComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

路由层通常再配合：

<!-- @code: middleware -->
```go
myApp.WithHTTPGroup("v1", middleware.Access(), middleware.Authz())
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[authz]
model_path  = "./example/config/casbin/model.conf"
policy_path = "./example/config/casbin/policy.csv"
adapter     = "file"
subject_key = "id"
object_mode = "route"
action_mode = "method"
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-authz-module-001]]
