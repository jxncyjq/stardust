---
id: "reference-redis-module-001"
title: "redis 模块使用参考"
aliases: ["redis模块", "RedisCli", "RedisManager", "RedisView"]
type: "reference"
category: "backend/library"
tags: ["redis", "cache", "pubsub", "stream", "notification"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-redis-001"
children: []
related_docs:
  - id: "reference-component-redis-001"
    relation: "depends_on"
    path: "./components/reference-component-redis-001.md"
---

# redis 模块使用参考

<!-- @section: overview -->
## 概述

`redis` 模块基于 `github.com/redis/go-redis/v9` 封装 Redis 连接、实例管理和带业务前缀的 `RedisCli` 访问视图。模块支持单实例、多实例和 Redis Cluster。

推荐在应用层通过 `components.RedisComponent()` 初始化；业务代码通过 `redis.GetRedisManager()` 获取 `RedisManager`，再按实例名创建 `RedisCli`。
<!-- @end-section -->

<!-- @section: config -->
## 配置结构

核心配置类型是 `redis.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    Name               string
    Addrs              []string
    Password           string
    KeyPrefix          string
    DbIndex            int
    DialTimeout        time.Duration
    ReadTimeout        time.Duration
    WriteTimeout       time.Duration
    ReadOnly           bool
    PoolSize           int
    PoolTimeout        time.Duration
    IdleTimeout        time.Duration
    IdleCheckFrequency time.Duration
    UseCluster         bool
    TLSConfig          *tls.Config
}
```
<!-- @end-code -->

关键字段：

| 字段 | 说明 |
| --- | --- |
| `name` | Redis 实例名，`RedisManager` 按此名称检索客户端 |
| `addrs` | Redis 地址列表；普通模式使用第一个地址，Cluster 模式使用全部地址 |
| `password` | Redis 密码 |
| `key_prefix` | 配置层记录的 key 前缀；业务通常也会通过 `conf.GetRedisKeyPrefix()` 传给 `GetRedisView` |
| `db_index` | 普通 Redis 客户端的 DB index，Cluster 模式不使用 |
| `use_cluster` | 为 `true` 时创建 `redis.ClusterClient` |
| `read_only` | Cluster 只读模式 |
| `pool_size` | 连接池大小，未配置或小于等于 0 时默认 `10` |
| `pool_timeout` | 连接池等待超时，未配置或小于等于 0 时默认 `4s` |

校验规则：

| 项 | 行为 |
| --- | --- |
| `addrs` 为空 | `NewRedisCmd` 返回 `ErrRedisAddrsEmpty` |
| `Init` 入参为空 | 返回 `redis config is empty` |
| 未调用 `Init` 就调用 `GetRedisManager` | 返回 `redis: not initialized, call Init() first` |
<!-- @end-section -->

<!-- @section: toml -->
## TOML 示例

<!-- @code: toml-config -->
```toml
[[redis]]
name         = "default"
addrs        = ["127.0.0.1:6379"]
password     = ""
key_prefix   = "stardust"
db_index     = 0
pool_size    = 10
pool_timeout = "4s"
use_cluster  = false

[[redis]]
name        = "session"
addrs       = ["127.0.0.1:6380"]
key_prefix  = "stardust-session"
db_index    = 1
pool_size   = 20
```
<!-- @end-code -->

Cluster 示例：

<!-- @code: cluster-config -->
```toml
[[redis]]
name        = "cluster"
addrs       = ["127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002"]
password    = ""
use_cluster = true
read_only   = true
```
<!-- @end-code -->

`Init` 支持单对象和数组两种 JSON 结构；TOML 中多实例通常使用 `[[redis]]`。
<!-- @end-section -->

<!-- @section: initialization -->
## 初始化方式

推荐使用组件：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.RedisComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动初始化：

<!-- @code: manual-init -->
```go
conf.Init()

if err := redis.Init(conf.Get("redis")); err != nil {
    return err
}

redisMgr, err := redis.GetRedisManager()
if err != nil {
    return err
}

cache := redisMgr.GetRedisView("default", conf.GetRedisKeyPrefix(), logger)
```
<!-- @end-code -->

注意：`Init` 只解析并保存配置；`GetRedisManager()` 或 `GetRedisDb()` 第一次被调用时才会创建底层连接。`RedisComponent` 会在 Init 阶段主动调用 `GetRedisManager()`，用于提前暴露连接错误。
<!-- @end-section -->

<!-- @section: manager -->
## RedisManager

`RedisManager` 以配置中的 `name` 管理多个 Redis 客户端：

<!-- @code: manager-api -->
```go
redisMgr, err := redis.GetRedisManager()
if err != nil {
    return err
}

cmd := redisMgr.GetRedisCmd("default")
cache := redisMgr.GetRedisView("default", "api", logger)
```
<!-- @end-code -->

| 方法 | 说明 |
| --- | --- |
| `GetRedisCmd(name)` | 返回底层 `RedisCmd`，可直接使用 go-redis 命令 |
| `GetRedisView(name, prefix, logger)` | 返回带 key prefix 的 `RedisCli` 业务视图 |

如果实例名不存在，两个方法都会返回 `nil`。

`GetRedisDb()` 是兼容入口，返回配置列表中第一个 Redis 客户端。新业务更推荐使用 `RedisManager` 按实例名显式获取。
<!-- @end-section -->

<!-- @section: view -->
## RedisCli 视图

`RedisCli` 是业务常用接口，封装 key 前缀和常用命令。所有通过 `RedisCli` 访问的 key 会被展开成：

<!-- @code: key-prefix -->
```go
fullKey := prefix + ":" + key
```
<!-- @end-code -->

示例：

<!-- @code: basic-cache -->
```go
cache := redisMgr.GetRedisView("default", "user", logger)

if err := cache.Set(ctx, "profile:1001", []byte(`{"name":"alice"}`), "10m"); err != nil {
    return err
}

value, err := cache.Get(ctx, "profile:1001")
if err != nil {
    if errors.Is(err, redis.Nil) {
        return nil
    }
    return err
}
```
<!-- @end-code -->

`Set`、`SetNX`、`Expire` 的 `duration` 参数使用 Go `time.ParseDuration` 格式，例如 `500ms`、`10s`、`5m`、`1h`。`Set` 和 `SetNX` 的 `duration` 为空字符串时表示不设置过期时间。
<!-- @end-section -->

<!-- @section: command-groups -->
## 常用命令族

| 命令族 | 方法 |
| --- | --- |
| Key/String | `Get`、`Set`、`SetNX`、`Del`、`Expire`、`Exists`、`Scan` |
| Hash | `HSetNX`、`HSet`、`HMSet`、`HGet`、`HMGet`、`HGetAll`、`HDel`、`HLen`、`HKeys`、`HValues`、`HExists` |
| List | `LPush`、`LAppend`、`LPop`、`LRPop`、`LRange`、`LLen`、`LIndex`、`LSet`、`LTrim`、`LRem`、`LInsert` |
| Set | `SAdd`、`SRem`、`SPop`、`SPopN`、`SLen`、`SDiff`、`SDiffMerge`、`SInter`、`SInterMerge`、`SUnion`、`SUnionMerge` |
| ZSet | `ZAdd`、`ZRem`、`ZLen`、`ZCount`、`ZRange`、`ZRangeByScore`、`ZRangeByLex`、`ZRank`、`ZIncr`、`ZIncrNX`、`ZInterMerge`、`ZUnionMerge` |
| Geo | `GeoAdd`、`GeoRadius`、`GeoRadiusByMember`、`GeoDist`、`GeoHash`、`GeoPos`、`GeoCalculateDistance` |
| Pub/Sub | `Subscribe`、`PSubscribe`、`Publish` |
| Stream | `XAdd` |

`RedisCli` 以 `[]byte` 作为主要 value 类型，调用方负责 JSON、protobuf 或其他业务编码。
<!-- @end-section -->

<!-- @section: hash-example -->
## Hash 示例

<!-- @code: hash-example -->
```go
fields := map[string][]byte{
    "name":  []byte("alice"),
    "email": []byte("alice@example.com"),
}

if err := cache.HMSet(ctx, "user:1001", fields); err != nil {
    return err
}

got, err := cache.HGetAll(ctx, "user:1001")
if err != nil {
    return err
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: zset-example -->
## ZSet 示例

<!-- @code: zset-example -->
```go
_, err := cache.ZAdd(ctx, "game:rank", &redis.ZMember{
    Score:  100,
    Member: []byte("user:1001"),
})
if err != nil {
    return err
}

members, err := cache.ZRange(ctx, "game:rank", 0, 9, true, true)
if err != nil {
    return err
}
```
<!-- @end-code -->

`ZRangeBy` 的 `Min`、`Max` 使用 Redis 原生边界表达式，例如 `"-inf"`、`"+inf"`、`"(10"`、`"[a"`。
<!-- @end-section -->

<!-- @section: pubsub-stream -->
## Pub/Sub 与 Stream

发布订阅：

<!-- @code: pubsub -->
```go
pubsub, err := cache.Subscribe(ctx, "events:user")
if err != nil {
    return err
}
defer pubsub.Close()

if _, err := cache.Publish(ctx, "events:user", []byte("created")); err != nil {
    return err
}
```
<!-- @end-code -->

Stream 写入：

<!-- @code: stream -->
```go
cmd := cache.XAdd(ctx, redislib.XAddArgs{
    Stream: "events",
    Values: map[string]any{
        "type": "user.created",
        "id":   "1001",
    },
})
if err := cmd.Err(); err != nil {
    return err
}
```
<!-- @end-code -->

上例中 `redislib` 是 `github.com/redis/go-redis/v9` 的导入别名，用于避免与本仓库 `redis` 包重名。
<!-- @end-section -->

<!-- @section: notification -->
## Notification 重试通知

`NewNotification(node, key, cache, policies)` 基于 Redis key 过期事件实现延迟重试通知：

<!-- @code: notification -->
```go
notifier := redis.NewNotification("node-1", "email", cache, nil)
if notifier == nil {
    return errors.New("notification init failed")
}

if err := notifier.Subscribe(ctx, func(p *redis.Payload, err error) redis.PutNext {
    if err != nil {
        return true
    }
    return false
}); err != nil {
    return err
}
```
<!-- @end-code -->

`policies` 为 `nil` 时使用默认退避序列：`1m`、`5m`、`10m`、`30m`、`60m`、`120m`。

使用前需要确认 Redis 已启用 keyspace notification，否则过期事件不会被推送。
<!-- @end-section -->

<!-- @section: native -->
## 底层客户端

如果封装接口未覆盖某个 Redis 命令，可以通过 `NativeCmd()` 或 `GetRedisCmd(name)` 访问 go-redis 原生命令：

<!-- @code: native -->
```go
cmd := cache.NativeCmd()
result, err := cmd.Do(ctx, "PING").Result()
if err != nil {
    return err
}
_ = result
```
<!-- @end-code -->

公共业务优先使用 `RedisCli`，只有缺少封装能力或需要 go-redis 特有参数时再访问原生命令。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `RedisCli` 会无条件使用 `prefix + ":" + key` 展开 key，传入空 prefix 会生成以 `:` 开头的 key。
- `Scan(ctx, cursor, match, count)` 会对 `match` 也加前缀，调用方通常传入业务侧 pattern，例如 `user:*`。
- `SDiffMerge`、`SInterMerge`、`SUnionMerge`、`ZInterMerge`、`ZUnionMerge` 的 destination 当前不会自动加前缀；如需统一命名，应传入完整目标 key 或改造封装。
- `ZLexCount` 当前返回 `not implemented`。
- `redis.Nil` 已通过本包导出为 `redis.Nil`，用于判断 key 不存在。
<!-- @end-section -->

## 相关文档

- [[reference-component-redis-001]]
- [[guide-app-components-module-001]]
