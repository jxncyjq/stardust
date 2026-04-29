---
id: "guide-progress-001"
title: "实现进度记录"
aliases: ["进度", "progress"]
type: "guide"
category: "meta"
tags: ["progress", "app", "component"]
version: "1.0.0"
created: "2026-04-26"
updated: "2026-04-26"
author: "jxncyjq"
status: "published"
parent: null
children: []
related_docs:
  - id: "design-app-component-001"
    relation: "references"
    path: "./design/design-app-component-001.md"
---


# app 包组件化统一接口 — 实现进度

<!-- @section: overview -->
## 概述

为 stardust 库引入统一的 `Component` 接口 + `Application` 编排层。
用户通过声明式 API 组合所需模块，框架自动处理依赖排序、Init/Start/Stop 生命周期。
纯叠加方案，零破坏现有 API。
<!-- @end-section -->

<!-- @section: files -->
## 新增文件清单

| 文件                             | 说明                                                |
| ------------------------------ | ------------------------------------------------- |
| `app/component.go`             | Component 接口 + ConfigFunc 类型                      |
| `app/container.go`             | Kahn 拓扑排序，统一 Init/Start/Stop，循环依赖检测               |
| `app/app.go`                   | Application 流式 API，信号监听，优雅关闭                      |
| `app/container_test.go`        | 7 个单元测试（正常/循环/缺失/重复依赖 + 顺序验证）                     |
| `app/components/helpers.go`    | recoverToError / requireConfig 共享工具               |
| `app/components/logs.go`       | logs 适配器                                          |
| `app/components/redis.go`      | redis 适配器                                         |
| `app/components/databases.go`  | databases 适配器                                     |
| `app/components/mongodb.go`    | mongodb 适配器                                       |
| `app/components/clickhouse.go` | clickhouse 适配器                                    |
| `app/components/nats.go`       | nats 适配器，Start 非阻塞（go StartAll），Stop 调 CloseAll   |
| `app/components/tracing.go`    | tracing 适配器，Stop 调 Shutdown(ctx) 刷 span           |
| `app/components/server.go`     | NewHTTPServerFromConfig / NewGRPCServerFromConfig |
| `example/main.go`              | 演示声明式启动（最新 canonical 示例）                       |
<!-- @end-section -->

<!-- @section: changes -->
## 现有文件改动

| 文件 | 改动 |
|------|------|
| `tracing/jaeger.go` | 新增 `Shutdown(ctx context.Context) error`，`Close()` 改为复用它 |
<!-- @end-section -->

<!-- @section: decisions -->
## 关键设计决策

- **零破坏**：现有所有包 API 不变，新包纯叠加
- **nats 阻塞问题**：`NatsConnection.Start()` 内含阻塞 for 循环，组件 `Start()` 改用 `go StartAll()` 包裹
- **panic 转 error**：各适配器 Init 通过 `defer recoverToError` 将现有包的 panic 转为 error
- **依赖缺失报错**：topoSort 检测未注册的依赖，返回明确错误而非静默跳过
<!-- @end-section -->

<!-- @section: verification -->
## 验证结果

```
go build ./app/...   ✅
go build ./...       ✅ 全量，零破坏现有包
go test ./app/... -v ✅ 7/7 PASS
```
<!-- @end-section -->

## 相关文档

- [[design-app-component-001|app 组件化统一接口设计文档]]
- [[reference-daily-20260426|2026-04-26 工作总结]]
