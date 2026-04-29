---
id: "reference-microservice-module-001"
title: "microService 模块使用参考"
aliases: ["microService模块", "微服务支撑模块", "breaker limit load metric register syncx tracing"]
type: "reference"
category: "backend/library"
tags: ["microservice", "breaker", "limit", "load", "metric", "register", "syncx", "tracing"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: null
children:
  - "reference-metric-module-001"
  - "reference-register-module-001"
related_docs:
  - id: "reference-metric-module-001"
    relation: "related_to"
    path: "./reference-metric-module-001.md"
  - id: "reference-register-module-001"
    relation: "related_to"
    path: "./reference-register-module-001.md"
  - id: "reference-component-tracing-001"
    relation: "related_to"
    path: "./components/reference-component-tracing-001.md"
  - id: "reference-component-redis-001"
    relation: "related_to"
    path: "./components/reference-component-redis-001.md"
  - id: "reference-redis-module-001"
    relation: "related_to"
    path: "./reference-redis-module-001.md"
---

# microService 模块使用参考

<!-- @section: overview -->
## 概述

本文档把 `breaker`、`limit`、`load`、`metric`、`register`、`syncx`、`tracing` 归纳为 microService 支撑能力。它们不属于单个 Go package，而是一组面向微服务运行时的基础模块：

| 模块 | 能力 | 典型位置 |
| --- | --- | --- |
| `breaker` | 下游调用熔断 | RPC/HTTP client 外层 |
| `limit` | 分布式或本地限流 | API、用户、租户、任务入口 |
| `load` | 本机自适应降载 | 请求入口、任务入口 |
| `metric` | Prometheus 自定义指标 | 包级变量或应用初始化 |
| `register` | 服务注册发现、网关注册 | 服务启动、客户端发现 |
| `syncx` | 合并并发重复请求 | 缓存回源、配置加载、热点查询 |
| `tracing` | OpenTelemetry/Jaeger 链路追踪 | 请求链路、跨服务调用 |

推荐组合顺序：先做降载和限流，再做熔断保护，下游调用前后记录 tracing 和 metrics。
<!-- @end-section -->

<!-- @section: breaker -->
## breaker 熔断

`breaker` 模块提供 `Breaker` 接口和 Google SRE 风格实现：

<!-- @code: breaker-api -->
```go
type Breaker interface {
    Allow() (Promise, error)
    Do(req func() error) error
    DoWithAcceptable(req func() error, acceptable func(err error) bool) error
}
```
<!-- @end-code -->

创建熔断器：

<!-- @code: breaker-create -->
```go
brk := breaker.NewGoogleBreaker()
```
<!-- @end-code -->

推荐使用 `Do` 包装下游调用：

<!-- @code: breaker-do -->
```go
err := brk.Do(func() error {
    return callPaymentService(ctx, req)
})
if errors.Is(err, breaker.ErrServiceUnavailable) {
    return err
}
```
<!-- @end-code -->

如果某些业务错误不应计入失败，使用 `DoWithAcceptable`：

<!-- @code: breaker-acceptable -->
```go
err := brk.DoWithAcceptable(
    func() error {
        return callInventory(ctx, sku)
    },
    func(err error) bool {
        return errors.Is(err, ErrSKUNotFound)
    },
)
```
<!-- @end-code -->

`GoogleBreaker` 使用滑动窗口统计成功数和请求总数，请求总数小于 `100` 时不会触发熔断；触发后返回 `breaker.ErrServiceUnavailable`。
<!-- @end-section -->

<!-- @section: limit -->
## limit 限流

`limit` 模块提供两类限流器：

| 类型 | 构造函数 | 说明 |
| --- | --- | --- |
| 固定窗口 | `NewPeriodLimiter(period, quota, prefix)` | 一个时间窗口内最多允许 `quota` 次 |
| 令牌桶 | `NewTokenLimiter(rate, burst, key)` | 每秒生成 `rate` 个令牌，桶容量 `burst` |

两个限流器都优先使用 `redis.GetRedisDb()` 执行 Lua 脚本；Redis 未初始化或执行失败时，会退回本地内存限流。

固定窗口示例：

<!-- @code: period-limit -->
```go
limiter := limit.NewPeriodLimiter(60, 100, "api:user")

status, err := limiter.Take(userID)
if err != nil {
    return err
}
if status == limit.OverQuotaStatus {
    return ErrTooManyRequests
}
```
<!-- @end-code -->

令牌桶示例：

<!-- @code: token-limit -->
```go
limiter := limit.NewTokenLimiter(10, 20, "api:order:create")

if !limiter.Allow() {
    return ErrTooManyRequests
}
```
<!-- @end-code -->

注意：本地 fallback 只在当前进程内生效，不具备跨实例一致性。生产环境的分布式限流应先初始化 Redis。
<!-- @end-section -->

<!-- @section: load -->
## load 自适应降载

`load` 模块提供 `AdaptiveShedder`，用于在本机负载过高时主动拒绝部分请求：

<!-- @code: load-create -->
```go
shedder := load.NewAdaptiveShedder(
    load.WithWindow(5*time.Second),
    load.WithBuckets(50),
    load.WithCpuThreshold(900),
)
```
<!-- @end-code -->

使用方式：

<!-- @code: load-allow -->
```go
promise, err := shedder.Allow()
if errors.Is(err, load.ErrServiceOverloaded) {
    return err
}
if err != nil {
    return err
}

defer promise.Fail()

if err := handle(ctx, req); err != nil {
    return err
}
promise.Pass()
```
<!-- @end-code -->

`WithCpuThreshold(900)` 表示 CPU 使用率阈值为 90%。Linux 平台会采样 `/proc/stat`；非 Linux 平台 `getCpuUsage()` 返回 `0`，CPU 触发条件等于禁用。

`Pass()` 会记录成功和耗时，`Fail()` 只减少当前并发数。成功请求必须调用 `Pass()`，失败或提前返回应调用 `Fail()`，否则 `flying` 并发数会不准确。
<!-- @end-section -->

<!-- @section: syncx -->
## syncx 合并并发请求

`syncx.SharedCalls` 用于把同一个 key 的并发请求合并为一次实际执行，类似 singleflight：

<!-- @code: shared-calls -->
```go
shared := syncx.NewSharedCalls()

val, err := shared.Do("config:global", func() (interface{}, error) {
    return loadConfigFromDB(ctx)
})
if err != nil {
    return err
}
cfg := val.(*Config)
```
<!-- @end-code -->

如果需要知道当前调用是否是真正执行者，使用 `DoEx`：

<!-- @code: shared-calls-doex -->
```go
val, fresh, err := shared.DoEx("user:1001", func() (interface{}, error) {
    return queryUser(ctx, 1001)
})
_ = fresh
_ = val
_ = err
```
<!-- @end-code -->

`fresh=true` 表示当前 goroutine 执行了 `fn`；`fresh=false` 表示复用了正在执行的结果。
<!-- @end-section -->

<!-- @section: metric -->
## metric 指标

`metric` 模块封装 Prometheus `CounterVec`、`GaugeVec`、`HistogramVec`。详见 [[reference-metric-module-001]]。

微服务中常见指标：

<!-- @code: metric-example -->
```go
var requestsTotal = metric.NewCounter(metric.CounterOpts{
    Namespace: "user_service",
    Name:      "requests_total",
    Help:      "Total number of handled requests.",
    Labels:    []string{"method", "result"},
})

requestsTotal.Inc("CreateUser", "success")
```
<!-- @end-code -->

指标应在包级变量或初始化阶段创建一次。构造函数使用 `prometheus.MustRegister`，重复注册相同指标名会 panic。
<!-- @end-section -->

<!-- @section: register -->
## register 注册发现

`register` 模块提供服务注册发现和 APISIX 网关注册，详见 [[reference-register-module-001]]。

服务注册示例：

<!-- @code: register-example -->
```go
reg, err := register.NewEtcdRegister(etcdConfigBytes)
if err != nil {
    return err
}
defer reg.Close()

registry := register.NewServiceRegistry(reg)
if err := registry.Register("user-service", "127.0.0.1", 9103, []string{"grpc"}, nil); err != nil {
    return err
}
defer registry.Deregister()
```
<!-- @end-code -->

gRPC server/client 已经在 `http_server` 包内接入 etcd：服务端配置 `Etcd` 后启动时注册，客户端配置 `Etcd` 后通过 resolver watch 服务实例。
<!-- @end-section -->

<!-- @section: tracing -->
## tracing 链路追踪

`tracing` 模块定义抽象接口 `CallItracksAbsInterface`，并提供基于 OpenTelemetry OTLP HTTP exporter 的 `JaegerTracer`。

配置结构：

<!-- @code: tracing-config -->
```go
type JaegerConfig struct {
    ServiceName string  `json:"service_name" yaml:"service_name"`
    Endpoint    string  `json:"endpoint" yaml:"endpoint"`
    SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}
```
<!-- @end-code -->

组件初始化：

<!-- @code: tracing-component -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.TracingComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动使用：

<!-- @code: tracing-manual -->
```go
tracer, err := tracing.NewJaegerTracer(conf.Get("tracing"))
if err != nil {
    return err
}
defer tracer.Close()

ctx, span := tracer.StartSpan(ctx, "CreateUser")
defer span.Finish()

span.SetTag("user.id", userID)
if err := doWork(ctx); err != nil {
    span.SetError(err)
    return err
}
```
<!-- @end-code -->

跨服务传播使用 `Inject` 和 `Extract`，carrier 需要实现 `propagation.TextMapCarrier`。`TracingComponent.Stop` 会调用 `Shutdown(ctx)`，刷新未发送 span。
<!-- @end-section -->

<!-- @section: composition -->
## 组合示例

以下示例展示入口保护、熔断、追踪和指标的组合方式：

<!-- @code: composition-example -->
```go
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    shedderPromise, err := s.shedder.Allow()
    if errors.Is(err, load.ErrServiceOverloaded) {
        s.requestsTotal.Inc("CreateUser", "overloaded")
        return err
    }
    if err != nil {
        return err
    }
    defer shedderPromise.Fail()

    if !s.tokenLimiter.Allow() {
        s.requestsTotal.Inc("CreateUser", "limited")
        return ErrTooManyRequests
    }

    ctx, span := s.tracer.StartSpan(ctx, "CreateUser")
    defer span.Finish()

    err = s.breaker.Do(func() error {
        return s.repo.Create(ctx, req)
    })
    if err != nil {
        span.SetError(err)
        s.requestsTotal.Inc("CreateUser", "error")
        return err
    }

    shedderPromise.Pass()
    s.requestsTotal.Inc("CreateUser", "success")
    return nil
}
```
<!-- @end-code -->

实际项目中可把这些能力封装进业务 service、middleware 或 client wrapper，避免散落在每个 handler 中。
<!-- @end-section -->

<!-- @section: recommendations -->
## 使用建议

- 入口层优先使用 `load` 和 `limit`，避免过载请求继续进入核心业务。
- 调用不稳定下游时使用 `breaker`，业务可接受错误应通过 `DoWithAcceptable` 排除。
- 热点回源和重复并发查询使用 `syncx.SharedCalls`，减少缓存击穿。
- 自定义业务指标使用 `metric`，HTTP 请求指标使用 `http_server/middleware.Metrics`。
- 服务注册发现和网关注册优先复用 `register`，gRPC 服务可直接使用已有 etcd 集成。
- 链路追踪优先通过 `TracingComponent` 初始化，Stop 阶段确保 `Shutdown` 被调用。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `limit` 的 Redis 路径依赖 `redis.Init` / `RedisComponent`；Redis 不可用时只做本地兜底。
- `load` 的 CPU 降载仅 Linux 有效；非 Linux 主要依赖最近丢弃状态。
- `breaker.GoogleBreaker` 请求量不足 `100` 时不会熔断。
- `metric` 构造函数会注册到默认 Prometheus registry，重复创建同名指标会 panic。
- `register.Init`、`tracing.NewJaegerTracer`、Prometheus 默认 registry 都具有进程级全局影响，测试中要避免重复初始化冲突。
- `syncx.SharedCalls` 只合并同时进行中的请求，不缓存已完成结果。
<!-- @end-section -->

## 相关文档

- [[reference-metric-module-001]]
- [[reference-register-module-001]]
- [[reference-component-tracing-001]]
- [[reference-component-redis-001]]
- [[guide-app-components-module-001]]
