---
id: "reference-nats-module-001"
title: "nats 模块使用参考"
aliases: ["nats模块", "NatsConnManager", "NatsConnection", "JetStream"]
type: "reference"
category: "backend/library"
tags: ["nats", "jetstream", "pubsub", "queue", "stream"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-05-02"
author: "jxncyjq"
status: "published"
parent: "reference-component-nats-001"
children: []
related_docs:
  - id: "reference-component-nats-001"
    relation: "depends_on"
    path: "./components/reference-component-nats-001.md"
---

# nats 模块使用参考

<!-- @section: overview -->
## 概述

`nats` 模块基于 `github.com/nats-io/nats.go` 封装 NATS 连接管理、普通发布订阅和 JetStream 工作队列消费。模块提供 `NatsConnManager` 管理多个命名连接，每个 `NatsConnection` 负责单条 NATS 连接的发布、订阅、Stream/Consumer 管理和连接状态监控。

推荐在应用层通过 `components.NatsComponent()` 初始化并启动连接监控；业务代码通过 `nats.GetNatsManager()` 获取连接，再显式调用发布或订阅方法。
<!-- @end-section -->

<!-- @section: config -->
## 配置结构

核心配置类型是 `nats.NatsConfig`：

<!-- @code: config-struct -->
```go
type NatsConfig struct {
    Name       string   `json:"name"`
    Url        string   `json:"url"`
    UseStream  bool     `json:"use_stream"`
    StreamName string   `json:"stream_name"`
    Type       string   `json:"type"`
    Subject    []string `json:"subjects"`
    Username   string   `json:"username"`
    Password   string   `json:"password"`
}
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `name` | 连接名称，`NatsConnManager` 按此名称检索连接 |
| `url` | NATS 服务地址，例如 `nats://127.0.0.1:4222` |
| `use_stream` | 是否启用 JetStream |
| `stream_name` | JetStream Stream 名称 |
| `type` | 预留字段，当前实现未使用 |
| `subjects` | Stream 绑定的 subjects，例如 `["orders.>"]` |
| `username` / `password` | 用户名密码；两者都非空时才启用 `nats.UserInfo` |

`NatsConfig.Validate()` 定义了两条校验规则：`url` 必填；`use_stream=true` 时 `stream_name` 必填。当前 `GetNatsManager()` 初始化路径没有主动调用 `Validate()`，因此业务或配置层应保证配置完整。
<!-- @end-section -->

<!-- @section: toml -->
## TOML 示例

当前 `nats.Init` 直接把配置解析为 `[]*NatsConfig`，因此配置应使用数组格式：

<!-- @code: toml-config -->
```toml
[[nats]]
name        = "default"
url         = "nats://127.0.0.1:4222"
use_stream  = true
stream_name = "orders"
subjects    = ["orders.>"]

[[nats]]
name       = "events"
url        = "nats://127.0.0.1:4222"
use_stream = false
subjects   = ["events.>"]
username   = "app"
password   = "secret"
```
<!-- @end-code -->

普通 NATS 发布订阅不需要 `stream_name`；JetStream 模式需要 `stream_name` 和 `subjects`。
<!-- @end-section -->

<!-- @section: initialization -->
## 初始化方式

推荐使用组件：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.NatsComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动初始化：

<!-- @code: manual-init -->
```go
conf.Init()

if err := nats.Init(conf.Get("nats")); err != nil {
    return err
}

natsMgr, err := nats.GetNatsManager()
if err != nil {
    return err
}

conn, ok := natsMgr.GetClient("default")
if !ok {
    return errors.New("nats client not found")
}
```
<!-- @end-code -->

`NatsComponent.Init` 会调用 `nats.Init(...)` 和 `nats.GetNatsManager()` 建立所有连接；`NatsComponent.Start` 会后台调用 `manager.StartAll()`，用于连接状态监控和 JetStream Stream 检查。JetStream 模式下，已有 Stream 会合并配置中的新增 subjects，不需要重启 NATS Server。
<!-- @end-section -->

<!-- @section: manager -->
## NatsConnManager

`NatsConnManager` 以配置中的 `name` 管理多个连接：

| 方法 | 说明 |
| --- | --- |
| `GetNatsManager()` | 返回全局单例 manager，首次调用时建立所有连接 |
| `GetClient(key)` | 按名称获取 `*NatsConnection`，返回连接和是否存在 |
| `StartAll()` | 为每条连接启动后台连接状态监控 |
| `CloseAll()` | 调用所有连接的 `Stop()` 并清空 manager |

示例：

<!-- @code: manager-api -->
```go
natsMgr, err := nats.GetNatsManager()
if err != nil {
    return err
}

conn, ok := natsMgr.GetClient("default")
if !ok {
    return errors.New("nats client not found")
}
```
<!-- @end-code -->

`StartAll()` 不会自动注册业务订阅；订阅需要业务代码显式调用 `StartSubscription`。
<!-- @end-section -->

<!-- @section: connection -->
## NatsConnection

`NatsConnection` 封装单条连接：

| 方法 | 说明 |
| --- | --- |
| `Publish(subject, data)` | 发布消息；JetStream 模式使用 `js.Publish`，普通模式使用 `conn.Publish` |
| `PublishAsync(subject, data)` | JetStream 异步发布；普通模式退化为 `Publish` |
| `StartSubscription(subject, durableName, handler)` | 启动普通队列订阅或 JetStream PullSubscribe；JetStream 模式会先确保 subject 已归入 Stream |
| `StopSubscription(subject)` | 取消指定 subject 订阅 |
| `StopAllSubscriptions()` | 取消当前连接的所有订阅 |
| `AddStream(streamName, subjects)` | JetStream 模式下创建 Stream；Stream 已存在时合并缺失 subjects |
| `AddConsumer(streamName, durableName, subjects...)` | JetStream 模式下创建 durable consumer |
| `EnsureStream()` | JetStream 模式下确保配置中的 Stream 存在，并合并配置中的新增 subjects |
| `Start()` | 运行连接状态监控，处理重连后 JetStream 上下文恢复 |
| `Stop()` | 取消连接上下文并关闭底层 NATS 连接 |
| `GetJetStream()` | 返回原生 `nats.JetStreamContext` |
| `GetNativeConn()` | 返回原生 `*nats.Conn` |
| `IsConnected()` | 返回底层连接是否已连接 |
| `GetConfig()` | 返回当前连接配置 |
<!-- @end-section -->

<!-- @section: publish -->
## 发布消息

普通 NATS 和 JetStream 使用同一个发布入口：

<!-- @code: publish -->
```go
if err := conn.Publish("orders.created", []byte(`{"id":"1001"}`)); err != nil {
    return err
}
```
<!-- @end-code -->

JetStream 异步发布：

<!-- @code: publish-async -->
```go
if err := conn.PublishAsync("orders.created", []byte(`{"id":"1001"}`)); err != nil {
    return err
}
```
<!-- @end-code -->

`PublishAsync` 只提交异步发布请求，当前封装没有等待 ack；如果业务需要确认发布结果，应通过 `GetJetStream()` 使用原生 driver 能力。
<!-- @end-section -->

<!-- @section: subscribe-normal -->
## 普通订阅

当 `use_stream=false` 时，`StartSubscription` 使用 `QueueSubscribe(subject, durableName, handler)`：

<!-- @code: subscribe-normal -->
```go
err := conn.StartSubscription("events.user", "worker-a", func(msg *nats.Msg) {
    _ = msg.Data
})
if err != nil {
    return err
}
```
<!-- @end-code -->

这里的 `durableName` 在普通 NATS 模式下作为 queue group 使用，同一 queue group 内只有一个消费者处理每条消息。
<!-- @end-section -->

<!-- @section: subscribe-stream -->
## JetStream 订阅

当 `use_stream=true` 时，`StartSubscription` 会：

1. 确保当前 `subject` 已归入 `stream_name` 对应的 Stream；如果已有 Stream subjects 没有覆盖该 subject，会调用 `UpdateStream` 合并。
2. 调用 `AddConsumer(streamName, durableName, subject)` 创建 durable consumer。
3. 使用 `PullSubscribe(subject, durableName, nats.BindStream(streamName))` 订阅。
4. 后台循环 `Fetch(10)` 拉取消息。
5. 对匹配 subject 的消息调用 handler 后自动 `Ack()`。

示例：

<!-- @code: subscribe-stream -->
```go
err := conn.StartSubscription("orders.created", "order-worker", func(msg *nats.Msg) {
    // 当前封装会在 handler 返回后自动 Ack。
    _ = msg.Data
})
if err != nil {
    return err
}
```
<!-- @end-code -->

如果 handler 内部处理失败，当前封装仍会在 handler 返回后自动 Ack。需要失败重投、延迟重试或手动 Ack 语义时，应使用 `GetJetStream()` 自行订阅，或调整封装接口。
<!-- @end-section -->

<!-- @section: stream-consumer -->
## Stream 和 Consumer

JetStream 模式下可以显式创建或更新 Stream，并创建 Consumer：

<!-- @code: stream-consumer -->
```go
if err := conn.AddStream("orders", []string{"orders.>"}); err != nil {
    return err
}

if err := conn.AddConsumer("orders", "order-worker", "orders.created"); err != nil {
    return err
}
```
<!-- @end-code -->

当前封装创建 Stream 时使用 `nats.WorkQueuePolicy`，语义是每条消息只能被一个消费者消费。`AddStream` 和 `EnsureStream` 在 Stream 已存在时不会重建 Stream，只会基于现有 `StreamConfig` 合并缺失的 subjects 并调用 `UpdateStream`；已被 `orders.>` 这类通配符覆盖的 subject 不会重复追加。`AddConsumer` 默认使用 `AckExplicitPolicy` 和 `DeliverAllPolicy`；传入多个 subject 时使用 `FilterSubjects`，单个 subject 时使用 `FilterSubject`。
<!-- @end-section -->

<!-- @section: native -->
## 原生能力

封装未覆盖的能力可以通过原生连接访问：

<!-- @code: native -->
```go
nativeConn := conn.GetNativeConn()
if err := nativeConn.Flush(); err != nil {
    return err
}

js := conn.GetJetStream()
info, err := js.StreamInfo("orders")
if err != nil {
    return err
}
_ = info
```
<!-- @end-code -->

公共业务优先使用 `NatsConnection` 的封装方法；需要高级 JetStream 配置、发布确认、手动 Ack、KV/Object Store 等能力时再访问原生 driver。
<!-- @end-section -->

<!-- @section: shutdown -->
## 停止和释放

取消单个订阅：

<!-- @code: stop-subscription -->
```go
if err := conn.StopSubscription("orders.created"); err != nil {
    return err
}
```
<!-- @end-code -->

关闭所有连接：

<!-- @code: close-all -->
```go
if err := natsMgr.CloseAll(); err != nil {
    return err
}
```
<!-- @end-code -->

通过 `NatsComponent` 接入应用时，框架 Stop 阶段会调用 `manager.CloseAll()`。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `nats.Init` 当前只解析数组配置；TOML 应使用 `[[nats]]`。
- `GetNatsManager` 是全局单例，首次创建后再次调用 `Init` 不会重建已有 manager。
- `NatsConfig.Validate()` 当前没有被 manager 初始化流程调用，配置完整性需要由配置层或业务层保证。
- `StartAll()` 启动的是连接状态监控，不会自动注册业务订阅。
- 新增 topic 如果已被现有通配符覆盖，不需要更新 Stream；如果未覆盖，应用启动、重连或 `StartSubscription` 会合并 subjects，不需要重启 NATS Server。
- JetStream 订阅在 handler 返回后自动 Ack，不适合需要失败重投的处理逻辑。
- `StopSubscription(subject)` 通过 `sub.Subject` 匹配；同一 subject 多个 durable 订阅时需要谨慎管理。
- `CloseAll()` 会关闭连接并清空 manager；关闭后再通过同一个全局 manager 取连接会失败。
<!-- @end-section -->

## 相关文档

- [[reference-component-nats-001]]
- [[guide-app-components-module-001]]
