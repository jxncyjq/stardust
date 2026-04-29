---
id: "reference-component-i18n-001"
title: "I18nComponent 使用说明"
aliases: ["I18nComponent", "国际化组件", "翻译组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "i18n", "translation", "localization"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-i18n-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-i18n-module-001"
    relation: "parent_of"
    path: "../reference-i18n-module-001.md"
---

# I18nComponent 使用说明

<!-- @section: overview -->
## 概述

`components.I18nComponent()` 负责初始化 go-i18n 支撑层。它读取 `i18n` 配置，加载嵌入式翻译文件，并把 bundle 注册到 `i18n` 包级管理器中，供 `middleware.I18n()` 和业务代码复用。
<!-- @end-section -->

<!-- @section: contract -->
## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.I18nComponent()` |
| 组件名 | `i18n` |
| 依赖 | `logs` |
| 配置 key | `i18n` |
| Init | 读取配置、加载嵌入式翻译文件、设置 bundle |
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
        components.I18nComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

路由层通常和 `middleware.I18n()` 配合：

<!-- @code: middleware -->
```go
myApp.WithHTTPGroup("v1",
    middleware.I18n(),
    middleware.Metrics(conf.GetAppName()),
)
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: config -->
## 配置示例

<!-- @code: config -->
```toml
[i18n]
default_language = "zh"
supported_languages = ["zh", "en"]
query_key = "lang"
header_key = "Accept-Language"
```
<!-- @end-code -->
<!-- @end-section -->

## 相关文档

- [[guide-app-components-module-001]]
- [[reference-i18n-module-001]]
