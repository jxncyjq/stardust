---
id: "reference-clickhouse-module-001"
title: "clickhouse 模块使用参考"
aliases: ["clickhouse模块", "ClickHouseManager", "ClickHouseCli"]
type: "reference"
category: "backend/library"
tags: ["clickhouse", "analytics", "sql", "batch", "async-insert"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-clickhouse-001"
children: []
related_docs:
  - id: "reference-component-clickhouse-001"
    relation: "depends_on"
    path: "./components/reference-component-clickhouse-001.md"
---

# clickhouse 模块使用参考

<!-- @section: overview -->
## 概述

`clickhouse` 模块基于 `github.com/ClickHouse/clickhouse-go/v2` 封装 ClickHouse 多实例连接管理和常用 SQL 操作。模块通过 `ClickHouseManager` 管理多个命名客户端，并通过 `ClickHouseCli` 暴露查询、执行、异步写入、批量写入和健康检查能力。

推荐在应用层通过 `components.ClickhouseComponent()` 初始化；业务代码通过 `clickhouse.GetClickHouseManager()` 或 `clickhouse.GetClient(name)` 获取客户端。
<!-- @end-section -->

<!-- @section: config -->
## 配置结构

核心配置类型是 `clickhouse.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    Name         string `json:"name"`
    Addr         string `json:"addr"`
    Database     string `json:"database"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    MaxConn      int    `json:"max_conn"`
    MaxIdle      int    `json:"max_idle"`
    DialTimeoutS int    `json:"dial_timeout_s"`
}
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `name` | 逻辑实例名，`ClickHouseManager` 按此名称检索客户端 |
| `addr` | ClickHouse Native 协议地址，例如 `127.0.0.1:9000` |
| `database` | 默认数据库 |
| `username` / `password` | 认证信息 |
| `max_conn` | 最大连接数，默认 `10` |
| `max_idle` | 最大空闲连接数，默认 `5` |
| `dial_timeout_s` | 连接和初始化 Ping 超时，单位秒，默认 `5` |

校验规则：

| 字段 | 要求 |
| --- | --- |
| `name` | 必填 |
| `addr` | 必填 |
| `database` | 必填 |

`Init` 入参为空会返回 `clickhouse: config is empty`；未调用 `Init` 就调用 `GetClickHouseManager` 会返回 `clickhouse: not initialized, call Init() first`。
<!-- @end-section -->

<!-- @section: toml -->
## TOML 示例

单实例：

<!-- @code: toml-single -->
```toml
[clickhouse]
name           = "default"
addr           = "127.0.0.1:9000"
database       = "analytics"
username       = "default"
password       = ""
max_conn       = 10
max_idle       = 5
dial_timeout_s = 5
```
<!-- @end-code -->

多实例：

<!-- @code: toml-multiple -->
```toml
[[clickhouse]]
name           = "default"
addr           = "127.0.0.1:9000"
database       = "analytics"
username       = "default"
password       = ""

[[clickhouse]]
name           = "archive"
addr           = "127.0.0.1:9001"
database       = "archive"
username       = "readonly"
password       = "secret"
max_conn       = 5
max_idle       = 2
dial_timeout_s = 3
```
<!-- @end-code -->

`Init` 支持单对象和数组两种 JSON 结构；TOML 中多实例通常使用 `[[clickhouse]]`。
<!-- @end-section -->

<!-- @section: initialization -->
## 初始化方式

推荐使用组件：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.ClickhouseComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动初始化：

<!-- @code: manual-init -->
```go
conf.Init()

if err := clickhouse.Init(conf.Get("clickhouse")); err != nil {
    return err
}

chMgr, err := clickhouse.GetClickHouseManager()
if err != nil {
    return err
}

client := chMgr.GetClient("default")
```
<!-- @end-code -->

注意：`Init` 只解析并保存配置；`GetClickHouseManager()` 第一次被调用时才会创建连接，并对每个实例执行 `PingContext`。`ClickhouseComponent` 会在 Init 阶段主动调用 `GetClickHouseManager()`，用于提前暴露连接错误。
<!-- @end-section -->

<!-- @section: manager -->
## ClickHouseManager

`ClickHouseManager` 以配置中的 `name` 管理多个 ClickHouse 客户端：

| 方法 | 说明 |
| --- | --- |
| `GetClickHouseManager()` | 返回全局单例 manager，首次调用时初始化所有连接 |
| `(*ClickHouseManager).GetClient(name)` | 返回指定名称的 `ClickHouseCli`，不存在时返回 `nil` |
| `clickhouse.GetClient(name)` | 快捷方法；未初始化或连接失败时返回 `nil` |

示例：

<!-- @code: manager-api -->
```go
chMgr, err := clickhouse.GetClickHouseManager()
if err != nil {
    return err
}

cli := chMgr.GetClient("default")
if cli == nil {
    return clickhouse.ErrNilClient
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: client-interface -->
## ClickHouseCli 接口

`ClickHouseCli` 封装基于 `database/sql` 的常用操作：

| 方法 | 说明 |
| --- | --- |
| `Exec(ctx, query, args...)` | 执行 DDL、INSERT 等不返回行的语句 |
| `Query(ctx, dest, query, args...)` | 查询多行并扫描到切片指针 |
| `QueryRow(ctx, dest, query, args...)` | 查询单行并扫描到结构体指针 |
| `AsyncInsert(ctx, query, wait, args...)` | 使用 ClickHouse `ASYNC INSERT` 写入 |
| `PrepareBatch(ctx, query)` | 准备批量写入语句，返回 `*sql.Stmt` |
| `DB()` | 返回底层 `*sql.DB` |
| `Ping(ctx)` | 健康检查 |

`Query` 的 `dest` 必须是 `*[]T` 或 `*[]*T`；`QueryRow` 的 `dest` 必须是结构体指针。
<!-- @end-section -->

<!-- @section: query -->
## 查询示例

查询结果按列名映射到结构体字段，优先匹配 `db` tag，未设置 tag 时匹配字段名：

<!-- @code: query-example -->
```go
type EventRow struct {
    UserID int64  `db:"user_id"`
    Event  string `db:"event"`
}

var rows []EventRow
if err := cli.Query(
    ctx,
    &rows,
    "SELECT user_id, event FROM events WHERE event_date = ?",
    "2026-04-27",
); err != nil {
    return err
}
```
<!-- @end-code -->

单行查询：

<!-- @code: query-row-example -->
```go
type CountRow struct {
    Count int64 `db:"count"`
}

var row CountRow
err := cli.QueryRow(ctx, &row, "SELECT count() AS count FROM events")
if errors.Is(err, sql.ErrNoRows) {
    return nil
}
if err != nil {
    return err
}
```
<!-- @end-code -->

如果查询返回了结构体中不存在的列，当前实现会丢弃该列。
<!-- @end-section -->

<!-- @section: exec -->
## 执行 DDL 和普通写入

<!-- @code: exec-example -->
```go
err := cli.Exec(ctx, `
    CREATE TABLE IF NOT EXISTS events (
        event_date Date,
        user_id Int64,
        event String
    ) ENGINE = MergeTree()
    ORDER BY (event_date, user_id)
`)
if err != nil {
    return err
}

if err := cli.Exec(
    ctx,
    "INSERT INTO events (event_date, user_id, event) VALUES (?, ?, ?)",
    "2026-04-27",
    int64(1001),
    "play",
); err != nil {
    return err
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: async-insert -->
## 异步写入

`AsyncInsert(ctx, query, wait, args...)` 会在调用方传入的 SQL 前拼接：

| `wait` | 实际前缀 |
| --- | --- |
| `false` | `ASYNC INSERT` |
| `true` | `ASYNC INSERT WAIT` |

示例：

<!-- @code: async-insert -->
```go
err := cli.AsyncInsert(
    ctx,
    "INTO events (event_date, user_id, event) VALUES (?, ?, ?)",
    false,
    "2026-04-27",
    int64(1001),
    "play",
)
if err != nil {
    return err
}
```
<!-- @end-code -->

注意传入的 query 通常从 `INTO ...` 开始；如果传入完整 `INSERT INTO ...`，会形成重复的 `ASYNC INSERT INSERT INTO ...`。
<!-- @end-section -->

<!-- @section: batch -->
## 批量写入

`PrepareBatch` 返回 `*sql.Stmt`，调用方负责执行和关闭：

<!-- @code: batch-example -->
```go
stmt, err := cli.PrepareBatch(ctx, "INSERT INTO events (event_date, user_id, event) VALUES (?, ?, ?)")
if err != nil {
    return err
}
defer stmt.Close()

for _, event := range events {
    if _, err := stmt.ExecContext(ctx, event.Date, event.UserID, event.Name); err != nil {
        return err
    }
}
```
<!-- @end-code -->

高吞吐埋点和事件写入优先使用批量写入；单条低频写入可直接使用 `Exec`。
<!-- @end-section -->

<!-- @section: native-db -->
## 底层 DB

复杂场景可通过 `DB()` 获取底层 `*sql.DB`：

<!-- @code: native-db -->
```go
db := cli.DB()
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
    return err
}
```
<!-- @end-code -->

公共业务优先使用 `ClickHouseCli`；只有事务、连接统计、driver 特定能力等场景再访问底层 `*sql.DB`。
<!-- @end-section -->

<!-- @section: errors -->
## 错误约定

| 错误 | 说明 |
| --- | --- |
| `ErrQueryFailed` | 业务可用于表达查询失败 |
| `ErrExecFailed` | 业务可用于表达执行失败 |
| `ErrNilClient` | 业务可用于表达客户端为空 |
| `ErrInvalidConfig` | 业务可用于表达配置非法 |
| `ErrBatchFailed` | 业务可用于表达批量写入失败 |

当前实现中，driver 或 `database/sql` 错误通常原样返回；例如 `QueryRow` 无数据时返回 `sql.ErrNoRows`。
<!-- @end-section -->

<!-- @section: testing -->
## 测试建议

业务代码应依赖 `ClickHouseCli` 接口，便于测试中替换为 mock：

<!-- @code: mock-example -->
```go
type mockClickHouseCli struct {
    query func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (m *mockClickHouseCli) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
    return m.query(ctx, dest, query, args...)
}
```
<!-- @end-code -->

仓库内 `clickhouse/clickhouse_test.go` 已展示了这种 mock 方式。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `ClickHouseManager` 是全局单例，配置初始化后再次调用 `Init` 不会重置已创建的 manager。
- 当前模块没有暴露统一 `Close` 方法；需要释放连接时只能通过 `cli.DB().Close()` 自行处理。
- 初始化时会执行 `PingContext`，因此 ClickHouse 需要在应用启动阶段可达。
- `Query` 和 `QueryRow` 依赖反射扫描，只支持按列名映射到 struct 字段；复杂类型需确认 driver 能正确 Scan。
- `structFields` 会对未匹配列进行丢弃，不会报错；字段遗漏可能被静默忽略。
- `AsyncInsert` 会自动拼接 `ASYNC INSERT` 前缀，调用方不要传完整 `INSERT` 前缀。
<!-- @end-section -->

## 相关文档

- [[reference-component-clickhouse-001]]
- [[guide-app-components-module-001]]
