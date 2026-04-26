---
id: "design-app-component-001"
title: "app 包 — 组件化统一接口设计"
aliases: ["app component", "组件化接口", "app编排层", "Component接口"]
type: "design"
category: "backend/library"
tags: ["app", "component", "lifecycle", "topo-sort", "dependency-injection"]
version: "1.0.0"
created: "2026-04-26"
updated: "2026-04-26"
author: "jxncyjq"
status: "published"
parent: null
children: []
related_docs:
  - id: "reference-daily-20260426"
    relation: "related_to"
    path: "../memory/2026-04-26-remaining-work-items.md"
---

# app 包 — 组件化统一接口设计

<!-- @section: background -->
## 背景

stardust 库有 20+ 个包，每个包独立 `Init`，调用方痛点：

- 手动按隐式顺序调用各包 `Init`，写错就 panic
- 各包错误行为不一致（有的 panic，有的 return error）
- 手动拼装 `ServiceGroup` 管理生命周期

目标：引入统一的 `Component` 接口 + `Application` 编排层，让用户通过声明式 API 组合所需模块，框架自动处理依赖排序、Init/Start/Stop 生命周期。
<!-- @end-section -->

<!-- @section: principles -->
## 设计原则

- **零破坏**：纯叠加方案，现有所有包 API 不变，现有代码继续有效
- **声明式**：用户只声明需要哪些组件，框架自动排序
- **失败快**：Init 阶段提前触发连接，暴露配置错误而非运行时崩溃
- **panic 安全**：适配器统一将现有包的 panic 转为 error
<!-- @end-section -->

<!-- @section: architecture -->
## 架构概览

```
app/
├── component.go          ← Component 接口 + ConfigFunc 类型
├── container.go          ← 拓扑排序 + 统一 Init/Start/Stop
├── app.go                ← Application 流式 API
└── components/           ← 各包适配器
    ├── helpers.go        ← recoverToError / requireConfig
    ├── logs.go
    ├── redis.go
    ├── databases.go
    ├── mongodb.go
    ├── clickhouse.go
    ├── nats.go
    ├── tracing.go
    └── server.go
```
<!-- @end-section -->

<!-- @section: interfaces -->
## 核心接口

<!-- @code: component-interface -->
```go
// Component 是所有基础设施组件的统一接口。
type Component interface {
    Name() string                                        // 唯一标识，用于依赖声明
    Dependencies() []string                              // 依赖的其他组件名列表
    Init(ctx context.Context, configFn ConfigFunc) error // 初始化，不启动 goroutine
    Start(ctx context.Context) error                     // 启动（非阻塞）
    Stop(ctx context.Context) error                      // 优雅关闭
}

// ConfigFunc 对接 conf.Get(key string) []byte。
type ConfigFunc func(key string) []byte
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: topo-sort -->
## 依赖排序（Kahn 算法）

`Container.Init` 在执行前先对注册的组件做拓扑排序：

1. 检测重复组件名 → error
2. 检测未注册的依赖声明 → error
3. Kahn BFS 排序 → 得到合法初始化顺序
4. 检测循环依赖（排序结果数量不等于总数）→ error

`Stop` 按逆拓扑顺序执行，收集全部 error 后一并返回。

### 各组件依赖关系

| 组件 | Name | Dependencies | Start | Stop |
|------|------|-------------|-------|------|
| LogsComponent | `logs` | — | — | — |
| RedisComponent | `redis` | `["logs"]` | — | — |
| DatabasesComponent | `databases` | `["logs"]` | — | — |
| MongoDBComponent | `mongodb` | `["logs"]` | — | — |
| ClickhouseComponent | `clickhouse` | `["logs"]` | — | — |
| NatsComponent | `nats` | `["logs"]` | `go StartAll()` | `CloseAll()` |
| TracingComponent | `tracing` | `["logs"]` | — | `Shutdown(ctx)` |

Init 顺序示例（全量注册）：
```
logs → redis, databases, mongodb, clickhouse, nats, tracing（并列，按注册顺序）
Stop 顺序：逆序
```
<!-- @end-section -->

<!-- @section: adapter-pattern -->
## 适配器模式

每个适配器的标准结构：

<!-- @code: adapter-pattern -->
```go
type redisComponent struct{}

func RedisComponent() app.Component { return &redisComponent{} }

func (c *redisComponent) Name() string          { return "redis" }
func (c *redisComponent) Dependencies() []string { return []string{"logs"} }

func (c *redisComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
    defer recoverToError(&retErr, "redis")       // panic → error
    redis.Init(requireConfig(configFn, "redis")) // 缺配置 → panic → error
    redis.GetRedisManager()                      // 提前建连，暴露连接错误
    return nil
}

func (c *redisComponent) Start(_ context.Context) error { return nil }
func (c *redisComponent) Stop(_ context.Context) error  { return nil }
```
<!-- @end-code -->

**特殊处理：**

- `nats`：`NatsConnection.Start()` 内含阻塞 for 循环，组件 `Start()` 用 `go c.manager.StartAll()` 包裹
- `tracing`：`JaegerTracer` 新增 `Shutdown(ctx context.Context) error` 方法（唯一的现有文件改动），Stop 调用它刷新未发送的 span
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

<!-- @code: full-example -->
```go
func main() {
    conf.Init()

    httpSrv, _ := components.NewHTTPServerFromConfig(conf.Get)
    httpSrv.RegisterHealthCheck()

    grpcSrv, _ := components.NewGRPCServerFromConfig(conf.Get)

    app.New(conf.Get).
        Use(
            components.LogsComponent(),
            components.RedisComponent(),
            components.DatabasesComponent(),
            components.NatsComponent(),
            components.TracingComponent(),
        ).
        WithHTTP(httpSrv).
        WithGRPC(grpcSrv).
        Run(context.Background())
}
```
<!-- @end-code -->

按需组合（只用部分组件）：

<!-- @code: partial-example -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.DatabasesComponent(),
    ).
    WithGRPC(grpcSrv).
    Run(context.Background())
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: existing-changes -->
## 现有文件改动（最小化）

| 文件 | 改动 |
|------|------|
| `tracing/jaeger.go` | 新增 `Shutdown(ctx context.Context) error`（3 行），`Close()` 改为复用它 |

其余所有包：**零改动**。
<!-- @end-section -->

<!-- @section: config -->
## 配置文件结构（TOML）

```toml
[global]
app_name         = "my-service"
app_version      = "1.0.0"
redis_key_prefix = "ms"

[logs]
filename = "./logs/app.log"
level    = 0

[redis]
name = "default"
addr = "127.0.0.1:6379"

[databases]
name   = "default"
driver = "mysql"
dsn    = "root:pwd@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True"

[tracing]
service_name = "my-service"
endpoint     = "jaeger:4318"
sample_rate  = 0.1

[[nats]]
name = "default"
url  = "nats://127.0.0.1:4222"

[http_server]
port    = 8080
address = "0.0.0.0"

[grpc_server]
listen_on = "0.0.0.0:9090"
timeout   = 5000
```
<!-- @end-section -->

## 相关文档

- [[reference-daily-20260426|2026-04-26 工作总结]]
