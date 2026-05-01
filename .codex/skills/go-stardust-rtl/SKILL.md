---
name: go-stardust-rtl
description: 开发 stardust 微服务业务功能时使用：写 HTTP/gRPC 接口、注册路由、数据库 DAO 事务、Redis 缓存、NATS 消息、权限鉴权、国际化，或查询 IHandler/NewHandler/BaseDao/RedisCli 等 stardust API 用法。
---

# Go Stardust RTL

stardust Go 微服务库的可复用参考技能包。

## 快速路由

清楚知道需求时直接按下表读对应文件；不确定时先读 `references/reference-map.md` 获取全局视图。

| 场景 / 关键词 | 读取 |
| --- | --- |
| 写接口、写 API、注册路由、CRUD、IHandler、NewHandler、Response、统一返回 | `references/http-server.md` |
| HTTP 接口目录结构、IHandler 规范、禁止 Gin 原生写法、检查清单 | `references/http-server.md` + `docs/reference/reference-http-server-api-usage-001.md` |
| 新增业务模块、app 生命周期装配、Business/SetupHTTP/SetupGRPC | `references/components.md` |
| 注册 gRPC 服务、GRPCServerComponent、proto 服务注册 | `references/grpc-server.md` |
| 接口鉴权、后台权限、RBAC、Casbin 策略、AuthzComponent | `references/authz.md` |
| 多语言、错误消息本地化、语言中间件、I18nComponent | `references/i18n.md` |
| 保存数据、查询数据、事务、DAO、GORM、数据库迁移 | `references/databases.md` |
| 缓存、排行榜、验证码、分布式锁、Redis Stream/PubSub | `references/redis.md` |
| 文档数据、MongoDB CRUD、ObjectID | `references/mongodb.md` |
| 分析查询、批量写入、ClickHouse | `references/clickhouse.md` |
| 发消息、消费消息、事件驱动、工作队列、NATS/JetStream | `references/nats.md` |
| 日志初始化、logger、zap 字段助手 | `references/logs.md` |
| 请求量、耗时指标、Prometheus、HTTP metrics | `references/metric.md` |
| 链路追踪、trace/span、Jaeger/OpenTelemetry | `references/tracing.md` |
| 服务注册发现、网关同步、APISIX/etcd | `references/register.md` |
| 熔断、限流、降载、请求合并 | `references/microservice.md` |
| session ID、worker ID、Snowflake ID | `references/uuid.md` |
| databases/redis/mongodb/clickhouse/nats 数据层总览 | `references/data-infra.md` |
| 维护 reference 文档、WikiLink、front matter 同步 | `references/doc-maintenance.md` |

## 触发机制

以下关键词或意图出现时必须优先使用本 skill：

**业务意图（中文）：**
写接口 / 写 API / 注册路由 / CRUD 实现、保存数据 / 查询数据库 / 事务操作、
做缓存 / 读缓存、发消息 / 消费消息 / 异步处理、
权限控制 / 鉴权 / 登录验证、多语言 / 国际化错误消息、新增业务模块 / 接入服务

**API 关键词（English）：**
`IHandler` / `NewHandler` / `NewHandlerRaw` / `Response` / `BaseResponse`、
`HTTPServerFromApp` / `HTTPServerComponent` / `GRPCServerComponent` / `Business`、
`BaseDao` / `SessionDao` / `SessionWrapper` / `FindById` / `InsertOne`、
`RedisCli` / `GetRedisView` / `PubSub` / `Stream`、
`AuthzComponent` / `I18nComponent` / `TracingComponent`、
`app.New` / `WithHTTPGroup` / `conf.Get`

## 原则

- 优先沿用 stardust 现有 API、组件和包结构
- example 保持声明式，生命周期逻辑放进组件或 service manager
- 基础设施优先通过 `app/components` 初始化
- 文档变更要同步 front matter 关系和 `docs/docs-index.md`
- Go 代码变更后运行匹配范围的 `go test`

## 常用验证

| 改动 | 命令 |
| --- | --- |
| 组件或 app | `go test ./app/components ./example` |
| service manager | `go test ./service ./app/components` |
| 单个基础设施模块 | `go test ./<module> ./app/components` |
| 微服务支撑模块 | `go test ./breaker ./limit ./load ./metric ./register ./syncx ./tracing` |
| 文档 | 检查 front matter、WikiLink、`docs/docs-index.md` |
