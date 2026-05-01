# nats 参考

`nats` 是 stardust 的消息中间件层，封装了连接管理、发布、
普通订阅和 JetStream 工作队列消费。

## 什么时候读这个文件

- 需要初始化 `components.NatsComponent()`
- 需要按 name 管理多个 NATS 连接
- 需要发布消息、监听主题、消费 JetStream

## 推荐接入方式

```go
components.NatsComponent()
```

组件会读取 `nats` 配置，调用 `nats.Init(...)`，再通过
`nats.GetNatsManager()` 建立所有连接。

## 配置

`NatsConfig` 常用字段：

- `name`
- `url`
- `use_stream`
- `stream_name`
- `subjects`
- `username`
- `password`

`Init` 读取的是 `[]*NatsConfig`，也就是 JSON 数组。

## 使用流程

```go
nats.Init(conf.Get("nats"))
mgr, err := nats.GetNatsManager()
conn, ok := mgr.GetClient("default")
if !ok {
    return
}
go mgr.StartAll()
```

## 发布消息

```go
if err := conn.Publish("user.created", payload); err != nil {
    return err
}
```

启用 `use_stream` 时会走 JetStream；否则走普通 NATS publish。
需要异步发布时用 `PublishAsync(...)`。

## 订阅消息

```go
err := conn.StartSubscription("user.created", "user-created", func(msg *nats.Msg) {
    // 处理消息
})
```

`StartSubscription` 会自动维护订阅状态，并在 JetStream 场景下创建 consumer。

## JetStream

- `EnsureStream()` 确保 stream 存在
- `AddStream(...)` 手动创建 stream
- `AddConsumer(...)` 手动创建 consumer
- `GetJetStream()` 获取底层 `nats.JetStreamContext`
- `GetNativeConn()` 获取底层 `*nats.Conn`

## 生命周期

- `StartAll()` 只负责启动连接监控 goroutine
- `CloseAll()` 会关闭所有连接并清空管理器
- `Start()` 是单连接的监控循环
- `Stop()` 关闭当前连接

## 注意事项

- `use_stream=true` 时必须提供 `stream_name`
- `subjects` 为空时 stream 不知道要绑定哪些 subject
- 业务订阅应显式调用 `StartSubscription`
- `PublishAsync` 仅在 JetStream 下才真正异步

## 验证

- `go test ./nats`
- `go test ./app/components`
