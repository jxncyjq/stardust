# metric 参考

`metric` 是 stardust 的 Prometheus 指标封装层，提供 `Counter`、`Gauge`、
`Histogram` 三类常用指标构造器。

## 什么时候读这个文件

- 需要为业务创建自定义 Prometheus 指标
- 需要把 HTTP 请求指标接入 Gin 中间件
- 需要知道 label 约束、注册时机、默认 registry 行为

## 推荐接入方式

```go
total := metric.NewCounter(metric.CounterOpts{
    Namespace: "user_service",
    Name:      "requests_total",
    Help:      "Total number of requests.",
    Labels:    []string{"method", "result"},
})
```

构造函数会立即调用 `prometheus.MustRegister`，因此同名指标只能创建一次。

## 指标类型

### Counter

只增不减，适合请求数、任务完成数、错误数。

常用方法：

- `Inc(labels...)`
- `Add(val, labels...)`

### Gauge

适合当前连接数、队列长度、在线人数等瞬时值。

常用方法：

- `Set(val, labels...)`
- `Inc(labels...)`
- `Add(val, labels...)`

### Histogram

适合延迟、耗时、分布统计。

常用方法：

- `Observe(val, labels...)`

## 标签约定

写入时传入的 label value 数量必须与配置中的 `Labels` 数量一致。
高基数字段不要放进 label，例如用户 ID、订单 ID、请求 ID。

## HTTP 指标

`metric` 模块本身只负责创建和写入指标，不负责暴露 `/metrics`。
Gin 服务通常配合 `http_server/middleware` 使用：

```go
r := gin.New()
r.GET("/metrics", middleware.MetricsHandler())
r.Use(middleware.Metrics("user_service"))
```

`middleware.Metrics(...)` 会注册并写入 HTTP 请求总量、耗时和 in-flight 指标。

## 生命周期建议

指标应在包级变量或初始化阶段创建一次，不要在请求处理函数里反复创建。

## 注意事项

- 重复注册同名指标会 panic
- 当前封装使用默认 Prometheus registry
- 没有 `Unregister`，测试中要避免反复构造同名指标
- Histogram 的 bucket 应按业务单位提前设计

## 验证

- `go test ./metric`
- `go test ./http_server`
