---
id: "reference-mongodb-module-001"
title: "mongodb 模块使用参考"
aliases: ["mongodb模块", "MongoManager", "MongoCli", "MongoDB"]
type: "reference"
category: "backend/library"
tags: ["mongodb", "mongo-driver", "document", "crud", "manager"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-mongodb-001"
children: []
related_docs:
  - id: "reference-component-mongodb-001"
    relation: "depends_on"
    path: "./components/reference-component-mongodb-001.md"
---

# mongodb 模块使用参考

<!-- @section: overview -->
## 概述

`mongodb` 模块基于 `go.mongodb.org/mongo-driver/v2` 封装 MongoDB 多实例连接管理和常用 CRUD 操作。模块提供 `MongoManager` 管理多个命名客户端，并通过 `MongoCli` 接口让业务代码依赖抽象，方便测试时替换为 mock。

推荐在应用层通过 `components.MongoDBComponent()` 初始化；业务代码通过 `mongodb.GetMongoManager()` 或 `mongodb.GetClient(name)` 获取客户端。
<!-- @end-section -->

<!-- @section: config -->
## 配置结构

核心配置类型是 `mongodb.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    Name     string `json:"name"`
    URI      string `json:"uri"`
    Database string `json:"database"`
    MaxPool  uint64 `json:"max_pool"`
    MinPool  uint64 `json:"min_pool"`
    TimeoutS int    `json:"timeout_s"`
}
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `name` | 逻辑实例名，`MongoManager` 按此名称检索客户端 |
| `uri` | MongoDB 连接 URI，例如 `mongodb://user:pass@host:27017/db?authSource=admin` |
| `database` | 默认操作的数据库名 |
| `max_pool` | 最大连接池大小，默认 `10` |
| `min_pool` | 最小连接池大小，默认 `2` |
| `timeout_s` | 连接、默认操作和 Ping 超时时间，单位秒，默认 `5` |

校验规则：

| 字段 | 要求 |
| --- | --- |
| `name` | 必填 |
| `uri` | 必填 |
| `database` | 必填 |

`Init` 入参为空会返回 `mongodb: config is empty`；未调用 `Init` 就调用 `GetMongoManager` 会返回 `mongodb: not initialized, call Init() first`。
<!-- @end-section -->

<!-- @section: toml -->
## TOML 示例

单实例：

<!-- @code: toml-single -->
```toml
[mongodb]
name      = "default"
uri       = "mongodb://root:password@127.0.0.1:27017/mydb?authSource=admin"
database  = "mydb"
max_pool  = 10
min_pool  = 2
timeout_s = 5
```
<!-- @end-code -->

多实例：

<!-- @code: toml-multiple -->
```toml
[[mongodb]]
name      = "default"
uri       = "mongodb://root:password@127.0.0.1:27017/app?authSource=admin"
database  = "app"
max_pool  = 20
min_pool  = 2
timeout_s = 5

[[mongodb]]
name      = "analytics"
uri       = "mongodb://root:password@127.0.0.1:27017/analytics?authSource=admin"
database  = "analytics"
max_pool  = 10
min_pool  = 1
timeout_s = 3
```
<!-- @end-code -->

`Init` 支持单对象和数组两种 JSON 结构；TOML 中多实例通常使用 `[[mongodb]]`。
<!-- @end-section -->

<!-- @section: initialization -->
## 初始化方式

推荐使用组件：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.MongoDBComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动初始化：

<!-- @code: manual-init -->
```go
conf.Init()

if err := mongodb.Init(conf.Get("mongodb")); err != nil {
    return err
}

mongoMgr, err := mongodb.GetMongoManager()
if err != nil {
    return err
}

client := mongoMgr.GetClient("default")
```
<!-- @end-code -->

注意：`Init` 只解析并保存配置；`GetMongoManager()` 第一次被调用时才会创建连接，并对每个实例执行 `Ping`。`MongoDBComponent` 会在 Init 阶段主动调用 `GetMongoManager()`，用于提前暴露连接错误。
<!-- @end-section -->

<!-- @section: manager -->
## MongoManager

`MongoManager` 以配置中的 `name` 管理多个 MongoDB 客户端：

<!-- @code: manager-api -->
```go
mongoMgr, err := mongodb.GetMongoManager()
if err != nil {
    return err
}

defaultCli := mongoMgr.GetClient("default")
analyticsCli := mongodb.GetClient("analytics")
```
<!-- @end-code -->

| 方法 | 说明 |
| --- | --- |
| `GetMongoManager()` | 返回全局单例 `MongoManager`，首次调用时初始化所有连接 |
| `(*MongoManager).GetClient(name)` | 返回指定名称的 `MongoCli`，不存在时返回 `nil` |
| `mongodb.GetClient(name)` | 快捷方法；未初始化或连接失败时返回 `nil` |

业务代码需要显式判断 `GetClient` 返回值，避免实例名错误导致后续空指针。
<!-- @end-section -->

<!-- @section: client-interface -->
## MongoCli 接口

`MongoCli` 封装默认数据库下的集合操作：

| 方法 | 说明 |
| --- | --- |
| `InsertOne(ctx, collection, doc)` | 插入单个文档，返回 `_id` 字符串 |
| `InsertMany(ctx, collection, docs)` | 批量插入文档，返回 `_id` 字符串列表 |
| `FindOne(ctx, collection, filter, result)` | 查询单个文档，未找到时返回 `ErrNotFound` |
| `FindMany(ctx, collection, filter, results, opts...)` | 查询多个文档，结果写入切片指针 |
| `UpdateOne(ctx, collection, filter, update)` | 更新第一个匹配文档，返回 `ModifiedCount` |
| `UpdateMany(ctx, collection, filter, update)` | 更新所有匹配文档，返回 `ModifiedCount` |
| `DeleteOne(ctx, collection, filter)` | 删除第一个匹配文档，返回 `DeletedCount` |
| `DeleteMany(ctx, collection, filter)` | 删除所有匹配文档，返回 `DeletedCount` |
| `CountDocuments(ctx, collection, filter)` | 统计匹配文档数量 |
| `Collection(name)` | 返回原生 `*mongo.Collection`，用于聚合、索引等复杂场景 |
| `Ping(ctx)` | 检查连接健康状态 |

`MongoCli` 的 `filter`、`update`、`doc` 通常使用 `bson.M`、`bson.D` 或业务结构体。
<!-- @end-section -->

<!-- @section: crud-example -->
## CRUD 示例

<!-- @code: crud-example -->
```go
package users

import (
    "context"
    "errors"

    "github.com/jxncyjq/stardust/mongodb"
    "go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
    ID    bson.ObjectID `bson:"_id,omitempty"`
    Name  string        `bson:"name"`
    Email string        `bson:"email"`
}

func CreateUser(ctx context.Context, cli mongodb.MongoCli, user *User) (string, error) {
    return cli.InsertOne(ctx, "users", user)
}

func FindUser(ctx context.Context, cli mongodb.MongoCli, id string) (*User, error) {
    objectID, err := mongodb.ObjectID(id)
    if err != nil {
        return nil, err
    }

    var user User
    err = cli.FindOne(ctx, "users", bson.M{"_id": objectID}, &user)
    if errors.Is(err, mongodb.ErrNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: query-options -->
## 查询选项

`FindMany` 透传 MongoDB driver v2 的 `options.FindOptions`：

<!-- @code: query-options -->
```go
opts := options.Find().
    SetSort(bson.D{{Key: "created_at", Value: -1}}).
    SetLimit(20).
    SetSkip(0)

var users []User
if err := cli.FindMany(ctx, "users", bson.M{"status": "active"}, &users, opts); err != nil {
    return err
}
```
<!-- @end-code -->

查询结果参数必须传入指针。`FindOne` 的 `result` 应为结构体指针或 `*bson.M`；`FindMany` 的 `results` 应为切片指针。
<!-- @end-section -->

<!-- @section: update-delete -->
## 更新和删除

更新语句需要使用 MongoDB 原生 update 文档，例如 `$set`：

<!-- @code: update-delete -->
```go
modified, err := cli.UpdateOne(
    ctx,
    "users",
    bson.M{"_id": userID},
    bson.M{"$set": bson.M{"name": "bob"}},
)
if err != nil {
    return err
}
if modified == 0 {
    return mongodb.ErrUpdateFailed
}

deleted, err := cli.DeleteOne(ctx, "users", bson.M{"_id": userID})
if err != nil {
    return err
}
if deleted == 0 {
    return mongodb.ErrDeleteFailed
}
```
<!-- @end-code -->

`UpdateOne` 和 `UpdateMany` 返回的是实际修改数量，不是匹配数量；如果新值与旧值相同，MongoDB 可能返回 `0`。
<!-- @end-section -->

<!-- @section: native-collection -->
## 原生 Collection

当封装接口不能覆盖复杂场景时，可以通过 `Collection(name)` 使用原生 driver 能力，例如聚合查询：

<!-- @code: native-collection -->
```go
pipeline := mongo.Pipeline{
    {{"$match", bson.D{{Key: "status", Value: "active"}}}},
    {{"$group", bson.D{
        {Key: "_id", Value: "$role"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
    }}},
}

cursor, err := cli.Collection("users").Aggregate(ctx, pipeline)
if err != nil {
    return err
}
defer cursor.Close(ctx)
```
<!-- @end-code -->

公共业务优先使用 `MongoCli`，只有聚合、索引、事务等未封装能力才直接访问原生 `*mongo.Collection`。
<!-- @end-section -->

<!-- @section: object-id -->
## ObjectID

`mongodb.ObjectID(hex)` 是 `bson.ObjectIDFromHex` 的便捷封装：

<!-- @code: object-id -->
```go
objectID, err := mongodb.ObjectID("507f1f77bcf86cd799439011")
if err != nil {
    return err
}
```
<!-- @end-code -->

`InsertOne` 和 `InsertMany` 会将 driver 返回的 `bson.ObjectID` 转成十六进制字符串；如果 `_id` 使用非 `ObjectID` 类型，当前封装会返回空字符串。
<!-- @end-section -->

<!-- @section: errors -->
## 错误约定

| 错误 | 说明 |
| --- | --- |
| `ErrNotFound` | `FindOne` 未找到文档时返回 |
| `ErrInsertFailed` | 业务可用于表达插入失败 |
| `ErrUpdateFailed` | 业务可用于表达更新失败 |
| `ErrDeleteFailed` | 业务可用于表达删除失败 |
| `ErrNilClient` | 业务可用于表达客户端为空 |
| `ErrInvalidConfig` | 业务可用于表达配置非法 |

当前实现中，driver 的 `mongo.ErrNoDocuments` 会被转换成 `mongodb.ErrNotFound`；其他 driver 错误会原样返回。
<!-- @end-section -->

<!-- @section: testing -->
## 测试建议

业务代码应依赖 `MongoCli` 接口，而不是直接依赖具体实现。测试中可以用 mock 实现接口：

<!-- @code: mock-example -->
```go
type mockMongoCli struct {
    findOne func(ctx context.Context, collection string, filter interface{}, result interface{}) error
}

func (m *mockMongoCli) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
    return m.findOne(ctx, collection, filter, result)
}
```
<!-- @end-code -->

仓库内 `mongodb/mongodb_test.go` 已展示了这种 mock 方式。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `MongoManager` 是全局单例，配置初始化后再次调用 `Init` 不会重置已创建的 manager。
- 当前模块没有暴露 `Disconnect` 或 `Close` 方法，应用退出时无法通过 manager 统一释放 MongoDB 客户端。
- `newMongoClient` 会在初始化时执行 `Ping(readpref.Primary())`，因此需要保证 MongoDB 在应用启动阶段可达。
- `timeout_s` 会同时设置 `SetConnectTimeout` 和 `SetTimeout`，长查询或聚合需要结合业务 `context` 控制超时。
- `GetClient(name)` 不存在或初始化失败时返回 `nil`，业务层要显式判断。
<!-- @end-section -->

## 相关文档

- [[reference-component-mongodb-001]]
- [[guide-app-components-module-001]]
