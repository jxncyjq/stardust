# logs 参考

`logs` 是 stardust 的日志基础层。它封装了 `zap` 和文件滚动，
所有组件和基础设施模块都建议先初始化它。

## 什么时候读这个文件

- 需要初始化全局 logger
- 需要在业务代码里统一使用 `logs.GetLogger(...)`
- 需要构造带 `module`、`duration`、`error` 的结构化日志字段

## 推荐接入方式

```go
components.LogsComponent()
```

`LogsComponent()` 必须最先初始化，因为后续 `redis`、`nats`、`register`、
`http_server` 等模块都会取 logger。

## 配置

`logs.Init(...)` 读取 JSON 配置，常用字段：

- `filename`
- `maxsize`
- `maxage`
- `maxbackups`
- `localtime`
- `compress`
- `level`

## 使用方式

```go
if err := logs.Init(conf.Get("logs")); err != nil {
    return err
}

logger := logs.GetLogger("user_service")
logger.Info("service started", logs.String("name", "user"))
```

`GetLogger(module)` 会自动附加 `module` 字段；未初始化时返回 `zap.NewNop()`。

## 字段助手

常用 helper：

- `logs.String(k, v)`
- `logs.Int(k, v)`
- `logs.Int64(k, v)`
- `logs.Int32(k, v)`
- `logs.Duration(k, v)`
- `logs.ErrorInfo(err)`

## 注意事项

- `Init` 只会执行一次
- `LogsComponent()` 依赖为空，但必须放在最前面
- `ErrorInfo(nil)` 会返回 `zap.Skip()`

## 验证

- `go test ./logs`
- `go test ./app/components`
