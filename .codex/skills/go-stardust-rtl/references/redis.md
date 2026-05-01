# redis 参考

`redis` 是 stardust 的缓存与消息层。它封装了 `go-redis/v9`，
同时提供带 key 前缀的 `RedisCli` 和管理器入口。

## 什么时候读这个文件

- 需要初始化 Redis 组件
- 需要读写缓存、集合、排序集、地理位置、PubSub、Stream
- 需要把 Redis 作为通知、锁、队列或缓存层使用

## 推荐接入方式

```go
components.RedisComponent()
```

组件会读取 `redis` 配置，并通过 `redis.GetRedisManager()` 触发初始化。

## 配置

常用字段：

- `name`: 逻辑名称
- `addrs`: 地址列表
- `password`: 密码
- `key_prefix`: key 前缀
- `db_index`: 单机模式数据库索引
- `use_cluster`: 是否启用集群
- `pool_size`: 连接池大小
- `dial_timeout` / `read_timeout` / `write_timeout`

## 使用流程

```go
redis.Init(conf.Get("redis"))
mgr, err := redis.GetRedisManager()
cli := mgr.GetRedisView("default", conf.GetRedisKeyPrefix(), logger)
```

`GetRedisDb()` 返回默认 `RedisCmd`，适合直接调用原生命令。
`GetRedisView()` 返回带前缀包装的 `RedisCli`，适合业务缓存。

## key 前缀规则

- `RedisView` 会把业务 key 展开成 `prefix + ":" + key`
- 空前缀会生成 `:key`
- `KeyPrefix()` 可查看当前前缀
- `NativeCmd()` 可回到底层 `redis.Cmdable`

## 常用能力

`RedisCli` 按数据结构分组：

- String / Key: `Get`、`Set`、`SetNX`、`Del`、`Expire`、`Exists`
- Hash: `HSet`、`HGet`、`HMSet`、`HGetAll`
- List: `LPush`、`LPop`、`LRange`、`LLen`
- Set: `SAdd`、`SRem`、`SPop`、`SUnion`
- ZSet: `ZAdd`、`ZRange`、`ZRangeByScore`、`ZUnionMerge`
- Geo: `GeoAdd`、`GeoRadius`、`GeoDist`
- PubSub: `Subscribe`、`PSubscribe`、`Publish`
- Stream: `XAdd`

## 典型用法

```go
cache := redisMgr.GetRedisView("default", "user", logger)
_ = cache.Set(ctx, "profile:1001", []byte("..."), "10m")
data, err := cache.Get(ctx, "profile:1001")
```

## 注意事项

- `Set` / `SetNX` 的 duration 使用 `time.ParseDuration` 语义字符串
- `GetRedisDb()` 在未初始化时返回 `nil`
- `GetRedisView()` 找不到名称时返回 `nil`
- `ZLexCount` 目前未实现

## 验证

- `go test ./redis`
- `go test ./app/components`
