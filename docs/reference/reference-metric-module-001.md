---
id: "reference-metric-module-001"
title: "metric 模块使用参考"
aliases: ["metric模块", "Prometheus指标", "CounterVec", "GaugeVec", "HistogramVec"]
type: "reference"
category: "backend/library"
tags: ["metric", "prometheus", "counter", "gauge", "histogram"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-microservice-module-001"
children: []
related_docs:
  - id: "reference-component-http-server-001"
    relation: "related_to"
    path: "./components/reference-component-http-server-001.md"
  - id: "reference-microservice-module-001"
    relation: "depends_on"
    path: "./reference-microservice-module-001.md"
---

# metric 模块使用参考

<!-- @section: overview -->
## 概述

`metric` 模块基于 `github.com/prometheus/client_golang/prometheus` 封装常用 Prometheus 指标类型，提供更简洁的 `CounterVec`、`GaugeVec`、`HistogramVec` 构造和写入方法。

该模块适合业务自定义指标。HTTP 请求指标由 `http_server/middleware` 中的 `Metrics` 中间件单独实现。
<!-- @end-section -->

<!-- @section: metric-types -->
## 指标类型

| 类型 | 构造函数 | 常用场景 |
| --- | --- | --- |
| Counter | `NewCounter` | 只增不减的累计值，例如请求数、任务完成数 |
| Gauge | `NewGauge` | 可增可减的瞬时值，例如连接数、队列长度 |
| Histogram | `NewHistogram` | 分布统计，例如请求耗时、任务处理耗时 |

三个构造函数都会创建对应的 Prometheus `*Vec`，并立即调用 `prometheus.MustRegister` 注册到默认 registry。
<!-- @end-section -->

<!-- @section: counter -->
## Counter

配置结构：

<!-- @code: counter-opts -->
```go
type CounterOpts struct {
    Namespace string
    Subsystem string
    Name      string
    Help      string
    Labels    []string
}
```
<!-- @end-code -->

使用示例：

<!-- @code: counter-example -->
```go
requests := metric.NewCounter(metric.CounterOpts{
    Namespace: "stardust",
    Subsystem: "http",
    Name:      "requests_total",
    Help:      "Total number of HTTP requests.",
    Labels:    []string{"method", "path", "code"},
})

requests.Inc("GET", "/api/users", "200")
requests.Add(5, "POST", "/api/users", "201")
```
<!-- @end-code -->

方法：

| 方法 | 说明 |
| --- | --- |
| `Inc(labels...)` | 指标值加 1 |
| `Add(val, labels...)` | 指标值增加 `val` |
<!-- @end-section -->

<!-- @section: gauge -->
## Gauge

配置结构：

<!-- @code: gauge-opts -->
```go
type GaugeOpts struct {
    Namespace string
    Subsystem string
    Name      string
    Help      string
    Labels    []string
}
```
<!-- @end-code -->

使用示例：

<!-- @code: gauge-example -->
```go
connections := metric.NewGauge(metric.GaugeOpts{
    Namespace: "stardust",
    Subsystem: "server",
    Name:      "connections",
    Help:      "Current active connections.",
    Labels:    []string{"type"},
})

connections.Set(10, "websocket")
connections.Inc("websocket")
connections.Add(-1, "websocket")
```
<!-- @end-code -->

方法：

| 方法 | 说明 |
| --- | --- |
| `Set(val, labels...)` | 设置当前值 |
| `Inc(labels...)` | 当前值加 1 |
| `Add(val, labels...)` | 当前值增加 `val`，可传负数 |
<!-- @end-section -->

<!-- @section: histogram -->
## Histogram

配置结构：

<!-- @code: histogram-opts -->
```go
type HistogramOpts struct {
    Namespace string
    Subsystem string
    Name      string
    Help      string
    Labels    []string
    Buckets   []float64
}
```
<!-- @end-code -->

使用示例：

<!-- @code: histogram-example -->
```go
duration := metric.NewHistogram(metric.HistogramOpts{
    Namespace: "stardust",
    Subsystem: "jobs",
    Name:      "duration_ms",
    Help:      "Job processing duration in milliseconds.",
    Labels:    []string{"name", "status"},
    Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
})

duration.Observe(42.5, "sync-user", "success")
```
<!-- @end-code -->

如果 `Buckets` 为空，构造函数会使用 `prometheus.DefBuckets`。
<!-- @end-section -->

<!-- @section: labels -->
## 标签约定

写入指标时传入的 label value 数量必须与配置中的 `Labels` 数量一致：

<!-- @code: labels -->
```go
counter := metric.NewCounter(metric.CounterOpts{
    Namespace: "stardust",
    Name:      "events_total",
    Help:      "Total number of events.",
    Labels:    []string{"type", "result"},
})

counter.Inc("order.created", "success")
```
<!-- @end-code -->

如果 label 数量不匹配，底层 Prometheus 客户端会 panic。高基数字段不要放进 label，例如用户 ID、订单 ID、请求 ID。
<!-- @end-section -->

<!-- @section: expose -->
## 暴露指标

`metric` 模块只负责创建和写入指标，不负责暴露 `/metrics` HTTP 端点。Gin 服务可使用 `http_server/middleware` 中的处理器：

<!-- @code: expose -->
```go
r := gin.New()
r.GET("/metrics", middleware.MetricsHandler())
```
<!-- @end-code -->

如果需要采集 HTTP 请求指标，可以使用中间件：

<!-- @code: http-metrics -->
```go
r.Use(middleware.Metrics("user_service"))
```
<!-- @end-code -->

`middleware.Metrics` 会注册并写入 `http_requests_total`、`http_request_duration_ms`、`http_requests_in_flight` 三类 HTTP 指标。
<!-- @end-section -->

<!-- @section: lifecycle -->
## 生命周期建议

指标应在包级变量或应用初始化阶段创建一次：

<!-- @code: lifecycle -->
```go
var taskTotal = metric.NewCounter(metric.CounterOpts{
    Namespace: "stardust",
    Subsystem: "worker",
    Name:      "tasks_total",
    Help:      "Total number of processed tasks.",
    Labels:    []string{"type", "result"},
})

func handleTask(taskType string, err error) {
    if err != nil {
        taskTotal.Inc(taskType, "error")
        return
    }
    taskTotal.Inc(taskType, "success")
}
```
<!-- @end-code -->

不要在请求处理函数、循环或可重复执行的初始化路径里反复调用 `NewCounter`、`NewGauge`、`NewHistogram`。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- 构造函数使用 `prometheus.MustRegister`，相同指标名重复注册会 panic。
- 当前封装只使用默认 Prometheus registry，不支持自定义 registry。
- 当前封装没有 `Unregister` 方法，测试中需要避免重复创建同名指标。
- `Namespace`、`Subsystem`、`Name` 会共同组成最终指标名，命名应保持稳定。
- `Help` 应清晰描述指标含义和单位，例如 `_ms` 指标应说明单位为毫秒。
- Histogram 的观测值单位由调用方约定，模块不会自动转换。
<!-- @end-section -->

## 相关文档

- [[reference-component-http-server-001]]
- [[guide-app-components-module-001]]
