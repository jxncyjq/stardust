# 数据与基础设施总览

> 这个文件只做导航。要看详细使用方式，请直接读：
>
> - `databases.md`
> - `redis.md`
> - `mongodb.md`
> - `clickhouse.md`
> - `nats.md`

## 阅读顺序

1. 先看对应模块的单独文件。
2. 如果需要比较组件接入方式，再回到 `components.md`。
3. 如果要维护技能索引，再看 `reference-map.md`。

## databases

详见 `databases.md`。这里仅保留一句话：它是关系型数据库总入口，推荐通过
`components.DatabasesComponent()` 进入。

## redis

详见 `redis.md`。这里仅保留一句话：它是缓存与消息层总入口，推荐通过
`components.RedisComponent()` 进入。

## mongodb

详见 `mongodb.md`。这里仅保留一句话：它是文档型数据库总入口，推荐通过
`components.MongoDBComponent()` 进入。

## clickhouse

详见 `clickhouse.md`。这里仅保留一句话：它是列式数据库总入口，推荐通过
`components.ClickhouseComponent()` 进入。

## nats
详见 `nats.md`。这里仅保留一句话：它是消息中间件总入口，推荐通过
`components.NatsComponent()` 进入。
