---
id: "reference-index-001"
title: "reference 文档目录"
aliases: ["reference index", "参考文档目录", "reference目录"]
type: "reference"
category: "backend/library"
tags: ["reference", "index", "docs"]
version: "1.0.0"
created: "2026-05-01"
updated: "2026-05-01"
author: "jxncyjq"
status: "published"
parent: "reference-docs-index"
children: []
related_docs:
  - id: "reference-docs-index"
    relation: "depends_on"
    path: "../reference-docs-index.md"
  - id: "reference-docs-index-table"
    relation: "related_to"
    path: "../docs-index.md"
---

# reference 文档目录

<!-- @section: overview -->
## 概述

本文档按功能分类整理 `docs/reference/` 下的参考文档。新增、删除或移动 reference 文档时，需要同步维护 [[reference-docs-index-table]]。
<!-- @end-section -->

<!-- @section: core-modules -->
## 核心基础模块

| 文档 | 说明 |
| --- | --- |
| [[reference-uuid-module-001]] | Snowflake ID、随机字符串/字节、worker 管理 |
| [[reference-i18n-module-001]] | 本地化、Bundle 管理、请求上下文语言切换 |
| [[reference-authz-module-001]] | Casbin 授权、RBAC 模型、策略文件配置、GORM 适配器 |
<!-- @end-section -->

<!-- @section: production-modules -->
## 生产支撑模块

| 文档 | 说明 |
| --- | --- |
| [[reference-microservice-module-001]] | breaker、limit、load、syncx 等微服务支撑能力 |
| [[reference-metric-module-001]] | Prometheus Counter、Gauge、Histogram 和 HTTP 指标 |
| [[reference-register-module-001]] | 服务注册/发现、etcd、APISIX 网关、gRPC 集成 |
| [[reference-http-server-api-usage-001]] | IHandler 接口注册、Response 统一返回、HTTP 模块目录结构 |
<!-- @end-section -->

<!-- @section: data-storage -->
## 数据存储模块

| 文档 | 说明 |
| --- | --- |
| [[reference-redis-module-001]] | Redis 连接管理、RedisCli 视图、Pub/Sub、Stream、Notification |
| [[reference-mongodb-module-001]] | MongoDB 多实例管理、CRUD 封装、ObjectID、错误约定 |
| [[reference-databases-module-001]] | GORM 封装、DAO、分页、事务、主从读写分离 |
| [[reference-clickhouse-module-001]] | ClickHouse 多实例管理、SQL 查询/执行、异步写入、批量写入 |
| [[reference-nats-module-001]] | NATS 连接管理、普通发布订阅、JetStream 工作队列、Stream/Consumer |
<!-- @end-section -->

<!-- @section: components -->
## 组件使用说明

| 文档 | 说明 |
| --- | --- |
| [[reference-component-logs-001]] | LogsComponent 日志初始化与配置 |
| [[reference-component-tracing-001]] | TracingComponent OpenTelemetry/Jaeger 链路追踪初始化 |
| [[reference-component-redis-001]] | RedisComponent 连接管理器初始化 |
| [[reference-component-nats-001]] | NatsComponent NATS/JetStream 连接管理器初始化 |
| [[reference-component-mongodb-001]] | MongoDBComponent 连接管理器初始化 |
| [[reference-component-databases-001]] | DatabasesComponent GORM 数据库管理器初始化 |
| [[reference-component-clickhouse-001]] | ClickhouseComponent ClickHouse 连接管理器初始化 |
| [[reference-component-i18n-001]] | I18nComponent go-i18n 国际化支撑层初始化 |
| [[reference-component-authz-001]] | AuthzComponent Casbin 授权器初始化 |
| [[reference-component-http-server-001]] | HTTPServerComponent Gin HTTP 服务器创建与启动 |
| [[reference-component-http-server-from-app-001]] | HTTPServerFromApp 绑定 Application 的 HTTP 组件写法 |
| [[reference-component-grpc-server-001]] | GRPCServerComponent gRPC 服务器创建与 Protobuf 注册 |
| [[reference-component-business-001]] | BusinessComponent 业务模块聚合、HTTP/gRPC 路由自动分发 |
<!-- @end-section -->

<!-- @section: dependency-notes -->
## 依赖关系速查

- [[reference-uuid-module-001]] 被 register 服务实例 ID 和 http_server session/worker 场景使用。
- [[reference-register-module-001]] 关联 gRPC 服务注册、发现和 APISIX 网关注册。
- [[reference-microservice-module-001]] 聚合 breaker、limit、load、syncx，并关联 metric、register、tracing。
- [[reference-metric-module-001]] 被 HTTP 中间件和微服务支撑模块引用。
- [[reference-authz-module-001]] 可使用 file 适配器，也可通过 databases/GORM 持久化策略。
- app/components 组件应优先通过 [[guide-app-components-module-001]] 接入生命周期。
<!-- @end-section -->

## 相关文档

- [[reference-docs-index]]
- [[reference-docs-index-table]]
