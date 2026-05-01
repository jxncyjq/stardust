# databases 参考

`databases` 是 stardust 的关系型数据库层。它提供统一的 `DatabaseManager`、`BaseDao`、
`SessionDao` 和 `BaseEntity` 包装，默认以 GORM 为底层实现。

## 什么时候读这个文件

- 需要初始化 MySQL / PostgreSQL / 其他 GORM 兼容数据库
- 需要用 `components.DatabasesComponent()` 接入 app 生命周期
- 需要 DAO、事务、迁移、分页、实体封装

## 推荐接入方式

```go
components.DatabasesComponent()
```

组件初始化时会读取 `databases` 配置，调用 `databases.Init(...)`，
并通过 `databases.GetDatabaseManager()` 提前暴露连接错误。

## 配置

支持单对象或数组两种 JSON 结构。

### Config 结构体

```go
type Config struct {
    Name           string   `json:"name"`
    ShowSql        bool     `json:"show_sql"`
    MaxIdle        int      `json:"max_idle"`
    MaxConn        int      `json:"max_conn"`
    Master         string   `json:"master"`
    Slaves         []string `json:"slaves"`
    UseMasterSlave bool     `json:"use_master_slave"`
    DbType         string   `json:"db_type"`
}
```

字段默认值：

| 字段 | 默认值 |
| --- | --- |
| `max_idle` | `10` |
| `max_conn` | `100` |
| `db_type` | `mysql` |

校验规则：`name` 和 `master` 必填；`max_idle >= 0`；`max_conn > 0`；`max_idle` 不能大于 `max_conn`。

### TOML 示例

```toml
[[databases]]
name             = "main"
db_type          = "mysql"
master           = "root:password@tcp(127.0.0.1:3306)/main_db?charset=utf8mb4&parseTime=True&loc=Local"
slaves           = ["root:password@tcp(127.0.0.1:3307)/main_db?charset=utf8mb4&parseTime=True&loc=Local"]
use_master_slave = true
max_conn         = 20
max_idle         = 5
show_sql         = false

[[databases]]
name     = "archive"
db_type  = "postgres"
master   = "host=127.0.0.1 port=5432 user=root dbname=archive_db password=password sslmode=disable"
max_conn = 10
max_idle = 2
```

## 使用流程

```go
databases.Init(conf.Get("databases"))
mgr, err := databases.GetDatabaseManager()
dao := mgr.GetDBDao("main")
```

`GetDBDao(name)` 返回 `BaseDao`，`GetDbInterface(name)` 返回底层 `DBConn`。

## 建模与访问约束

- 所有数据库表都必须包含 `ID` 字段，并将 `ID` 定义为数据库主键。
- `ID` 字段必须使用 `bigint64` 语义的自增整数；Go 结构体中使用 `int64`，数据库中使用对应的 `BIGINT` / `bigint` 自增主键类型。
- 所有时间相关字段必须使用时间戳，不使用 `datetime` / `timestamp with time zone` 等数据库时间类型。
- 时间戳字段必须使用 `bigint64` 语义；Go 结构体中使用 `int64`，数据库中使用对应的 `BIGINT` / `bigint` 类型。
- 数据库访问原则上通过 `BaseDao` / `SessionDao` / `Entity` 等 DAO 封装完成，不直接在业务代码中使用 native SQL 或底层 GORM API。
- 只有在 DAO 当前能力无法表达的场景下，才考虑使用 `Query`、`Native()` 或 `GetDbInterface(...)`，并应控制在基础设施适配层内，避免业务逻辑绑定具体 ORM，便于后续框架迁移。

## 常用 DAO 能力

`BaseDao` 和 `SessionDao` 都支持：

- `InsertOne` / `InsertMany`
- `Update` / `UpdateById` / `Upsert` / `UpsertById`
- `Delete` / `DeleteById`
- `FindById` / `FindOne` / `FindMany` / `FindAndCount`
- `Query`
- `Migrations`
- `GetDBMetas` / `GetTableMetas`
- `CallProcedure`

`SessionDao` 额外支持：

- `Begin`
- `Commit`
- `Rollback`
- `Close`
- `DB`

## 事务

推荐流程：

```go
session := dao.NewSession()
if err := session.Begin(); err != nil {
    return err
}
defer session.Close()

if _, err := session.InsertOne(obj); err != nil {
    _ = session.Rollback()
    return err
}
return session.Commit()
```

`Close()` 会在事务未提交时回滚。

## SessionWrapper

`SessionWrapper` 是一层更轻的事务编排封装，适合把一组写操作串成一个连续流程。

```go
sw := databases.NewSessionWrapper(dao)
err := sw.
    Execute(func(session databases.SessionDao) error {
        _, err := session.InsertOne(obj)
        return err
    }).
    Execute(func(session databases.SessionDao) error {
        _, err := session.UpdateById(id, obj)
        return err
    }).
    CommitAndClose()
if err != nil {
    return err
}
```

使用规则：

- `NewSessionWrapper(dao)` 会先创建 `SessionDao` 并调用 `Begin()`
- `Execute(...)` 返回新的 wrapper，任一步出错都会保留错误状态
- `CommitAndClose()` 在没有错误时提交并关闭事务
- `Close()` 在有错误时会回滚，再关闭 session

适合场景：

- 一个业务动作里有多次数据库写入
- 想把事务提交 / 回滚逻辑收敛到一个出口
- 想减少手写 `Begin` / `Rollback` / `Commit` 的样板代码

## Entity 模式

如果业务更偏向实体对象，可以用 `databases.NewEntity(dao, &Model{})`。
它提供：

- `Create`
- `Update`
- `Delete`
- `Get`
- `Exists`
- `Count`
- `Begin` / `Commit` / `Rollback`

## BaseDao 使用边界

### 优先使用 DAO 方法的场景

| 操作 | 推荐写法 | 说明 |
|------|----------|------|
| 单主键查找 | `found, err := dao.FindById(id, &bean)` | 返回 `(bool, error)`，`false+nil` 表示未找到 |
| 结构体字段精确 Count | `count, err := dao.Count(&Model{Field: val})` | GORM 以非零字段作 WHERE |
| 非零字段更新 | `dao.UpdateById(id, &Model{Status: "x"})` | 零值字段会被跳过 |
| 插入单条记录 | `dao.InsertOne(&bean)` | |

### 必须保留 `NewSession().DB()` 的场景

以下场景 DAO 接口无法表达，必须落回原生 GORM：

| 场景 | 示例 | 原因 |
|------|------|------|
| SQL 函数 / 聚合 | `DATE(created_at)`、`SUM(...)`、`COALESCE(...)` | BaseDao 不支持 SQL 函数 |
| 零值字段更新 | `Update("point_balance", 0)` | `UpdateById` 底层用 `Updates(struct)`，**零值字段被跳过** |
| 多字段 map 更新 | `Updates(map[string]any{...})` | UpdateById 只接受结构体 |
| NOT IN / 子查询 | `Where("id NOT IN (?)", ids)` | DAO 无法表达否定集合 |
| 关联预加载 | `Preload("Items").Find(...)` | DAO 不支持 Preload |
| 排序 / LIMIT 组合 | `Order("created_at DESC").Limit(n)` | FindMany 参数有限 |

### 零值字段更新的正确处理

```go
// ❌ 错误：newBal 可能为 0，UpdateById 会跳过该字段
dao.UpdateById(userID, &model.User{PointBalance: newBal})

// ✅ 正确：使用 NewSession().DB() 强制更新
dao.NewSession().DB().Model(&user).Update("point_balance", newBal)
```

### FindById 返回值处理

```go
found, err := dao.FindById(id, &bean)
if err != nil {
    return fmt.Errorf("db error: %w", err)
}
if !found {
    return ErrNotFound
}
// 此时 bean 已填充
```

## 辅助错误处理

模块提供一组结果检查函数，封装"操作有没有生效"的判断：

| 函数 | 空结果错误 |
| --- | --- |
| `DoGet` | `ErrGetEmpty` |
| `DoUpdate` | `ErrUpdatedEmpty` |
| `DoInsert` | `ErrInsertedEmpty` |
| `DoDelete` | `ErrDeletedEmpty` |

```go
err := databases.DoUpdate(func() (int64, error) {
    return dao.UpdateById(id, user)
})
```

## 注意事项

- `Init` 支持单对象或数组，但 `GetDatabaseManager()` 之前必须先 `Init`
- `max_idle` 不能大于 `max_conn`
- `master` 不能为空
- 如果需要主从，`use_master_slave` 不是装饰字段，实际连接逻辑会按配置走

## 验证

- `go test ./databases`
- `go test ./app/components`

