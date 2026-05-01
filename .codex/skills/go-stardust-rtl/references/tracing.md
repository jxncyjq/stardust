# tracing 参考

`tracing` 是 stardust 的链路追踪参考，覆盖 Jaeger / OpenTelemetry 初始化、
`TracingComponent` 接入，以及手动 span 使用方式。

## 什么时候读这个文件

- 需要初始化链路追踪组件
- 需要在业务代码里手动创建 span
- 需要理解 HTTP / gRPC server 与 tracing 的关系

## 推荐接入方式

```go
components.LogsComponent()
components.TracingComponent()
```

`components.TracingComponent()` 会读取 `tracing` 配置并创建全局 tracer provider。

## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.TracingComponent()` |
| 组件名 | `tracing` |
| 依赖 | `logs` |
| 配置 key | `tracing` |
| Init | 初始化 Jaeger tracer |
| Start | 无操作 |
| Stop | `Shutdown(ctx)` |

## 配置

```toml
[tracing]
service_name = "my-service"
endpoint     = "127.0.0.1:4318"
sample_rate  = 1.0
```

## 手动 span

```go
tracer, err := tracing.NewJaegerTracer(conf.Get("tracing"))
if err != nil {
    return err
}

ctx, span := tracer.StartSpan(ctx, "CreateUser")
defer span.Finish()
```

## 使用建议

- HTTP server 组件不强依赖 `tracing`；只有挂载 `middleware.Tracing(...)`
  或需要全局 tracer provider 时才注册 `TracingComponent()`。gRPC server 组件默认注入 tracing interceptor，仍依赖 `tracing`
- `Stop` 阶段要调用 `Shutdown(ctx)`，不要直接丢弃 provider
- span 名称建议使用业务动作名，不要直接用函数内部实现细节

## 注意事项

- `sample_rate` 影响采样比例
- 需要先初始化 `logs`，否则 tracing 组件无法正常输出日志
- HTTP 请求 tracing 通常和 `http_server/middleware.Tracing` 一起使用

## 验证

- `go test ./tracing`
- `go test ./app/components`
