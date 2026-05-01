# clickhouse 参考

`clickhouse` 是 stardust 的列式数据库层，封装了 `clickhouse-go/v2`，
提供 `ClickHouseManager` 和 `ClickHouseCli`。

## 什么时候读这个文件

- 需要初始化 ClickHouse 组件
- 需要写分析型 SQL、异步写入或批量写入
- 需要把 ClickHouse 当作独立读写存储使用

## 推荐接入方式

```go
components.ClickhouseComponent()
```

组件会读取 `clickhouse` 配置，并通过 `clickhouse.GetClickHouseManager()` 提前暴露连接错误。

## 配置

常用字段：

- `name`: 逻辑名称
- `addr`: `host:port`
- `database`: 默认数据库
- `username` / `password`
- `max_conn`: 最大连接数
- `max_idle`: 最大空闲连接数
- `dial_timeout_s`: 连接超时

## 使用流程

```go
clickhouse.Init(conf.Get("clickhouse"))
mgr, err := clickhouse.GetClickHouseManager()
cli := mgr.GetClient("default")
```

`GetClient(name)` 的快捷方法也会返回 `ClickHouseCli`，但未初始化时会返回 `nil`。

## 常用能力

`ClickHouseCli` 提供：

- `Exec`
- `Query`
- `QueryRow`
- `AsyncInsert`
- `PrepareBatch`
- `DB`
- `Ping`

## 查询映射

- `Query` / `QueryRow` 通过字段名或 `db` tag 扫描
- `Query` 的 `dest` 必须是切片指针
- `QueryRow` 的 `dest` 必须是结构体指针

## 典型用法

```go
var rows []Event
if err := cli.Query(ctx, &rows, "select id, name from events where kind = ?", "login"); err != nil {
    return err
}
```

异步写入示例：

```go
err := cli.AsyncInsert(ctx, "insert into events (id, name) values", true, 1, "login")
```

批量写入推荐：

```go
stmt, err := cli.PrepareBatch(ctx, "insert into events (id, name)")
```

## 注意事项

- `AsyncInsert` 通常配合 `INSERT INTO ...` 语句前缀使用
- `PrepareBatch` 适合高吞吐埋点 / 事件场景
- `Ping` 会在组件初始化阶段提前验证连接
- `SetMaxOpenConns`、`SetMaxIdleConns` 在初始化时直接生效

## 验证

- `go test ./clickhouse`
- `go test ./app/components`
