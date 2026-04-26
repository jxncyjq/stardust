---
id: "reference-docs-index"
title: "文档索引"
aliases: ["docs index", "文档目录", "索引"]
type: "reference"
category: "meta"
tags: ["index", "meta"]
version: "1.0.0"
created: "2026-04-26"
updated: "2026-04-26"
author: "jxncyjq"
status: "published"
---

# 文档索引

> 所有文档的统一索引。新增文档后必须在此登记。AI 通过此文件快速定位任意文档。

<!-- @section: index -->
## 索引表

| ID                                  | 标题              | 类型        | 路径                                               | 标签                                   |
| ----------------------------------- | --------------- | --------- | ------------------------------------------------ | ------------------------------------ |
| [[design-app-component-001]]        | app 包组件化统一接口设计  | design    | `docs/design/design-app-component-001.md`        | app, component, lifecycle, topo-sort |
| [[progress]]                        | 实现进度记录          | guide     | `docs/progress.md`                               | progress, app, component             |
| [[2026-04-26-remaining-work-items]] | 2026-04-26 工作总结 | reference | `docs/memory/2026-04-26-remaining-work-items.md` | daily, app, component                |
| [[design-stardust-bugfix-plan-001]] | stardust 库问题修补计划 | design | `docs/design/design-stardust-bugfix-plan-001.md` | bugfix, concurrency, reliability, panic |
<!-- @end-section -->

<!-- @section: keywords -->
## 关键词索引

| 关键词 | 文档 |
|--------|------|
| Component / 组件接口 | [[design-app-component-001]] |
| Container / 拓扑排序 | [[design-app-component-001]] |
| Application / 编排层 | [[design-app-component-001]] |
| Init / Start / Stop | [[design-app-component-001]] |
| nats / tracing 适配器 | [[design-app-component-001]] |
| 进度记录 | [[guide-progress-001]] |
| 2026-04-26 日志 | [[reference-daily-20260426]] |
| WebSocket 死锁 / panic | [[design-stardust-bugfix-plan-001]] |
| NATS StartAll 锁死 | [[design-stardust-bugfix-plan-001]] |
| panic → error 迁移 | [[design-stardust-bugfix-plan-001]] |
| goroutine 泄漏 | [[design-stardust-bugfix-plan-001]] |
<!-- @end-section -->
