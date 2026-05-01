---
id: "reference-docs-index"
title: "文档索引说明"
aliases: ["docs index", "文档目录", "索引", "文档索引"]
type: "reference"
category: "meta"
tags: ["index", "meta", "docs"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: null
children:
  - "reference-docs-index-table"
related_docs:
  - id: "reference-docs-index-table"
    relation: "related_to"
    path: "./docs-index.md"
---

# 文档索引说明

<!-- @section: overview -->
## 概述

本文档说明 stardust 文档索引的使用和维护方式。

当前项目存在两个相关文件：

| 文件 | 作用 |
| --- | --- |
| `docs/reference-docs-index.md` | 文档索引说明页，供 `[[reference-docs-index]]` WikiLink 跳转 |
| `docs/docs-index.md` | 实际索引表和关键词索引，新增文档后必须同步维护 |

新增文档时，优先更新 `docs/docs-index.md` 的索引表和关键词索引；本文档只在索引维护规则变化时更新。
<!-- @end-section -->

<!-- @section: index-entry -->
## 索引登记要求

每新增一篇文档，必须在 `docs/docs-index.md` 增加一行索引：

<!-- @code: index-row-template -->
```markdown
| [[文档ID]] | 文档标题 | type | `docs/相对路径.md` | tag1, tag2 |
```
<!-- @end-code -->

同时在关键词索引中增加可检索入口：

<!-- @code: keyword-row-template -->
```markdown
| 关键词 / 别名 | [[文档ID]] |
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: current-categories -->
## 当前文档分类

| 分类 | 目录 | 说明 |
| --- | --- | --- |
| 设计文档 | `docs/design/` | 架构设计、改造方案、技术决策 |
| 开发指南 | `docs/guides/` | 面向开发者的实践说明 |
| 模块参考 | `docs/reference/` | 核心模块、底层库和业务 API 使用说明 |
| 组件参考 | `docs/reference/components/` | `app/components` 各组件使用说明 |
| 记忆记录 | `docs/memory/` | 工作总结、待办、过程记忆 |
| 进度记录 | `docs/progress.md` | 当前实现进度 |
| 总索引 | `docs/docs-index.md` | 文档索引表和关键词索引 |
<!-- @end-section -->

<!-- @section: current-documents -->
## 当前核心文档

- [[design-app-component-001]] — app 包组件化统一接口设计
- [[guide-app-components-module-001]] — app/components/module 最小实践
- [[design-stardust-bugfix-plan-001]] — stardust 库问题修补计划
- [[reference-databases-module-001]] — databases 模块使用参考
- [[reference-redis-module-001]] — redis 模块使用参考
- [[reference-mongodb-module-001]] — mongodb 模块使用参考
- [[reference-nats-module-001]] — nats 模块使用参考
- [[reference-register-module-001]] — register 模块使用参考
- [[reference-metric-module-001]] — metric 模块使用参考
- [[reference-clickhouse-module-001]] — clickhouse 模块使用参考
- [[reference-microservice-module-001]] — microService 模块使用参考
- [[reference-http-server-api-usage-001]] — http_server 接口使用参考
- [[reference-uuid-module-001]] — uuid 模块使用参考
- [[reference-component-logs-001]] — LogsComponent 使用说明
- [[reference-component-redis-001]] — RedisComponent 使用说明
- [[reference-component-databases-001]] — DatabasesComponent 使用说明
- [[reference-component-mongodb-001]] — MongoDBComponent 使用说明
- [[reference-component-clickhouse-001]] — ClickhouseComponent 使用说明
- [[reference-component-nats-001]] — NatsComponent 使用说明
- [[reference-component-tracing-001]] — TracingComponent 使用说明
- [[reference-component-http-server-001]] — HTTPServerComponent 使用说明
- [[reference-component-http-server-from-app-001]] — HTTPServerFromApp 使用说明
- [[reference-component-grpc-server-001]] — GRPCServerComponent 使用说明
- [[reference-component-business-001]] — BusinessComponent 使用说明
- [[reference-component-authz-001]] — AuthzComponent 使用说明
- [[reference-component-i18n-001]] — I18nComponent 使用说明
- [[reference-authz-module-001]] — authz 模块使用参考
- [[reference-i18n-module-001]] — i18n 模块使用参考
<!-- @end-section -->

<!-- @section: maintenance -->
## 维护规则

- 文档必须包含 Front Matter。
- 内部链接使用 WikiLink：`[[文档ID]]`。
- 新增文档后同步更新 `docs/docs-index.md`。
- 修改文档内容时更新对应文档的 `updated` 字段。
- 如果新增文档类型或目录，需要同步更新本文档的“当前文档分类”。
<!-- @end-section -->

## 相关文档

- [[design-app-component-001]]
- [[guide-app-components-module-001]]
