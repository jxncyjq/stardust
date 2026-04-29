---
id: "reference-i18n-module-001"
title: "i18n 模块使用参考"
aliases: ["i18n模块", "go-i18n", "本地化"]
type: "reference"
category: "backend/library"
tags: ["i18n", "translation", "localization", "middleware"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-i18n-001"
children: []
related_docs:
  - id: "reference-component-i18n-001"
    relation: "depends_on"
    path: "./components/reference-component-i18n-001.md"
---

# i18n 模块使用参考

<!-- @section: overview -->
## 概述

`i18n` 模块基于 `github.com/nicksnyder/go-i18n/v2/i18n` 提供本地化能力。它只负责 bundle、localizer 和请求上下文的管理，不自己实现翻译引擎。
<!-- @end-section -->

<!-- @section: config -->
## 配置

核心配置类型是 `i18n.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    DefaultLanguage    string
    SupportedLanguages []string
    QueryKey           string
    HeaderKey          string
}
```
<!-- @end-code -->

默认值：

| 字段 | 默认值 |
| --- | --- |
| `default_language` | `zh` |
| `supported_languages` | `["zh", "en"]` |
| `query_key` | `lang` |
| `header_key` | `Accept-Language` |

校验规则：

| 项 | 行为 |
| --- | --- |
| `default_language` 为空 | `ErrInvalidConfig` |
| `supported_languages` 为空 | `ErrInvalidConfig` |
| `default_language` 不在 `supported_languages` 中 | `ErrInvalidConfig` |
| 语言标签非法 | `ErrUnknownLanguage` |
<!-- @end-section -->

<!-- @section: init -->
## 初始化

推荐通过组件初始化：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.I18nComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

也可以直接初始化：

<!-- @code: manual-init -->
```go
if err := i18n.Init(conf.Get("i18n")); err != nil {
    return err
}

bundle, err := i18n.LoadEmbeddedBundle()
if err != nil {
    return err
}
i18n.SetBundle(bundle)
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

语言上下文：

<!-- @code: context -->
```go
localizer, err := i18n.GetLocalizer("en")
msg, err := localizer.Localize(&i18n.LocalizeConfig{
    MessageID: "Hello",
    TemplateData: map[string]any{"Name": "Alice"},
})
```
<!-- @end-code -->

请求级本地化：

<!-- @code: request-localize -->
```go
msg, err := i18n.Localize(c.Request.Context(), "Hello", map[string]any{
    "Name": c.Param("name"),
})
```
<!-- @end-code -->

`middleware.I18n()` 会优先读取 query 参数 `query_key`，再读取 `header_key`，最后回退到默认语言。
<!-- @end-section -->

## 相关文档

- [[reference-component-i18n-001]]
- [[reference-docs-index]]
