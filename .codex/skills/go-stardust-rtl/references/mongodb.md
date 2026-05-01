# mongodb 参考

`mongodb` 是 stardust 的文档型数据库层，封装了 `mongo-driver/v2`，
提供 `MongoManager`、`MongoCli` 和 `ObjectID` 工具。

## 什么时候读这个文件

- 需要初始化 MongoDB 组件
- 需要做 CRUD、计数、集合访问、健康检查
- 需要处理 `ErrNotFound` 或把 `ObjectID` 转回字符串

## 推荐接入方式

```go
components.MongoDBComponent()
```

组件会读取 `mongodb` 配置，并通过 `mongodb.GetMongoManager()` 提前暴露连接错误。

## 配置

常用字段：

- `name`: 逻辑名称
- `uri`: MongoDB 连接串
- `database`: 默认数据库
- `max_pool`: 最大连接池
- `min_pool`: 最小连接池
- `timeout_s`: 单次操作超时

## 使用流程

```go
mongodb.Init(conf.Get("mongodb"))
mgr, err := mongodb.GetMongoManager()
cli := mgr.GetClient("default")
```

`GetClient(name)` 的快捷方法也会返回 `MongoCli`，但未初始化时会返回 `nil`。

## 常用能力

`MongoCli` 提供：

- `InsertOne`
- `InsertMany`
- `FindOne`
- `FindMany`
- `UpdateOne`
- `UpdateMany`
- `DeleteOne`
- `DeleteMany`
- `CountDocuments`
- `Collection`
- `Ping`

`FindOne` 未命中时返回 `mongodb.ErrNotFound`。

## 典型用法

```go
type User struct {
    ID   string `bson:"_id,omitempty" json:"id,omitempty"`
    Name string `bson:"name" json:"name"`
}

var u User
if err := cli.FindOne(ctx, "users", bson.M{"name": "alice"}, &u); err != nil {
    if errors.Is(err, mongodb.ErrNotFound) {
        // not found
    }
}
```

## 工具函数

- `mongodb.ObjectID(hex)` 把 hex 字符串转成 `bson.ObjectID`
- `Collection(name)` 适合聚合、索引、原生管道等复杂场景

## 注意事项

- `Init` 支持单对象或数组
- `Name`、`URI`、`Database` 不能为空
- 连接建立后会先 `Ping`
- `timeout_s` 同时影响连接和操作超时

## 验证

- `go test ./mongodb`
- `go test ./app/components`
