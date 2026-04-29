---
id: "reference-databases-module-001"
title: "databases 模块使用参考"
aliases: ["databases模块", "数据库模块", "BaseDao", "DatabaseManager"]
type: "reference"
category: "backend/library"
tags: ["databases", "gorm", "dao", "transaction", "mysql", "postgres"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-databases-001"
children: []
related_docs:
  - id: "reference-component-databases-001"
    relation: "depends_on"
    path: "./components/reference-component-databases-001.md"
---

# databases 模块使用参考

<!-- @section: overview -->
## 概述

`databases` 模块基于 GORM 封装关系型数据库连接管理、DAO、分页查询、事务和简单 Entity 操作。支持 MySQL、PostgreSQL，也支持 GORM DBResolver 的主从读写分离。

推荐在应用层通过 `components.DatabasesComponent()` 初始化；低层也可以直接调用 `databases.Init(...)` 和 `databases.GetDatabaseManager()`。
<!-- @end-section -->

<!-- @section: config -->
## 配置结构

核心配置类型是 `databases.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    Name           string   `json:"name" hcl:"name"`
    ShowSql        bool     `json:"show_sql" hcl:"show_sql"`
    MaxIdle        int      `json:"max_idle" hcl:"max_idle"`
    MaxConn        int      `json:"max_conn" hcl:"max_conn"`
    Master         string   `json:"master" hcl:"master"`
    Slaves         []string `json:"slaves" hcl:"slaves"`
    UseMasterSlave bool     `json:"use_master_slave" hcl:"use_master_slave"`
    DbType         string   `json:"db_type" hcl:"db_type"`
}
```
<!-- @end-code -->

默认值：

| 字段 | 默认值 |
| --- | --- |
| `max_idle` | `10` |
| `max_conn` | `100` |
| `db_type` | `mysql` |

校验规则：

| 字段 | 要求 |
| --- | --- |
| `name` | 必填 |
| `master` | 必填 |
| `max_idle` | `>= 0` |
| `max_conn` | `> 0` |
| `max_idle` | 不能大于 `max_conn` |
<!-- @end-section -->

<!-- @section: toml -->
## TOML 示例

<!-- @code: toml-config -->
```toml
[[databases]]
name             = "main"
db_type          = "mysql"
master           = "root:password@tcp(127.0.0.1:3306)/main_db?charset=utf8mb4&parseTime=True&loc=Local"
slaves           = [
  "root:password@tcp(127.0.0.1:3307)/main_db?charset=utf8mb4&parseTime=True&loc=Local",
]
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
<!-- @end-code -->

`Init` 支持单对象和数组两种 JSON 结构；TOML 中多实例通常使用 `[[databases]]`。
<!-- @end-section -->

<!-- @section: initialization -->
## 初始化方式

推荐使用组件：

<!-- @code: component-init -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.DatabasesComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

手动初始化：

<!-- @code: manual-init -->
```go
conf.Init()

if err := databases.Init(conf.Get("databases")); err != nil {
    return err
}

dbMgr, err := databases.GetDatabaseManager()
if err != nil {
    return err
}

mainDao := dbMgr.GetDBDao("main")
```
<!-- @end-code -->

注意：`Init` 只解析配置；`GetDatabaseManager` 才会创建连接。`DatabasesComponent` 会在 Init 阶段主动调用 `GetDatabaseManager`，提前暴露连接错误。
<!-- @end-section -->

<!-- @section: manager -->
## DatabaseManager

`DatabaseManager` 以配置中的 `name` 为 key 管理多个连接：

<!-- @code: manager-api -->
```go
dbMgr, err := databases.GetDatabaseManager()
dao := dbMgr.GetDBDao("main")
raw := dbMgr.GetDbInterface("main")
```
<!-- @end-code -->

| 方法 | 说明 |
| --- | --- |
| `GetDBDao(name)` | 返回 `BaseDao`，用于常规 CRUD |
| `GetDbInterface(name)` | 返回底层 `DBConn`，即 `*gorm.DB` 别名 |

如果名称不存在，两个方法都会返回 `nil`。
<!-- @end-section -->

<!-- @section: dao -->
## BaseDao 常用方法

`BaseDao` 封装常用 CRUD：

| 方法 | 说明 |
| --- | --- |
| `InsertOne(entry)` | 插入单条 |
| `InsertMany(entries...)` | 批量插入 |
| `Update(bean, where...)` | 按条件更新 |
| `UpdateById(id, bean)` | 按 id 更新 |
| `Upsert(where, bean)` | 存在则更新，不存在则插入 |
| `UpsertById(id, bean)` | 按 id upsert |
| `Delete(bean)` | 删除 |
| `DeleteById(id, bean)` | 按 id 删除 |
| `FindById(id, bean)` | 按主键查询 |
| `FindOne(bean)` | 按 bean 条件查询第一条 |
| `FindMany(rowsSlicePtr, sort, condiBean...)` | 列表查询 |
| `FindAndCount(rowsSlicePtr, pageable, condiBean...)` | 分页查询并返回总数 |
| `Query(rowsSlicePtr, sql, args...)` | 原生 SQL 查询 |
| `Native()` | 获取底层连接 |
| `Migrations(tables)` | `AutoMigrate` |
<!-- @end-section -->

<!-- @section: crud-example -->
## CRUD 示例

<!-- @code: crud-example -->
```go
type User struct {
    ID    int64  `gorm:"primaryKey;autoIncrement"`
    Name  string `gorm:"column:name"`
    Email string `gorm:"column:email"`
}

func (User) TableName() string { return "users" }

dao := dbMgr.GetDBDao("main")

user := &User{Name: "alice", Email: "alice@example.com"}
if _, err := dao.InsertOne(user); err != nil {
    return err
}

var got User
found, err := dao.FindById(user.ID, &got)
if err != nil {
    return err
}
if !found {
    return databases.ErrGetEmpty
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: pagination -->
## 分页查询

`NewPageable(skip, limit, sort)` 创建分页对象：

<!-- @code: pagination -->
```go
var users []User
pageable := databases.NewPageable(0, 20, "id desc")

total, err := dao.FindAndCount(&users, pageable, User{Name: "alice"})
if err != nil {
    return err
}
```
<!-- @end-code -->

`rowsSlicePtr` 必须是 slice 指针，否则会返回错误。
<!-- @end-section -->

<!-- @section: transaction -->
## 事务

### 显式 Session

<!-- @code: transaction-session -->
```go
session := dao.NewSession()
defer session.Close()

if err := session.Begin(); err != nil {
    return err
}

if _, err := session.InsertOne(user); err != nil {
    _ = session.Rollback()
    return err
}

return session.Commit()
```
<!-- @end-code -->

### SessionWrapper

<!-- @code: transaction-wrapper -->
```go
err := databases.NewSessionWrapper(dao).
    Execute(func(session databases.SessionDao) error {
        _, err := session.InsertOne(user)
        return err
    }).
    Execute(func(session databases.SessionDao) error {
        _, err := session.UpdateById(user.ID, user)
        return err
    }).
    CommitAndClose()
```
<!-- @end-code -->

`SessionWrapper` 任一步出错会回滚；无错误时提交。
<!-- @end-section -->

<!-- @section: entity -->
## Entity 包装

`NewEntity(dao, objectPtr)` 把结构体包装成带 CRUD 方法的 Entity。结构体需要实现：

<!-- @code: table-interface -->
```go
type Table interface {
    TableName() string
    PrimaryKey() interface{}
}
```
<!-- @end-code -->

示例：

<!-- @code: entity-example -->
```go
type User struct {
    ID   int64
    Name string
}

func (u *User) TableName() string { return "users" }
func (u *User) PrimaryKey() interface{} { return u.ID }

entity := databases.NewEntity(dao, &User{ID: 1})
found, err := entity.Get()
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: migration -->
## 表迁移

`Migrations` 直接调用 GORM `AutoMigrate`：

<!-- @code: migration -->
```go
err := dao.Migrations([]interface{}{
    &User{},
    &Order{},
})
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: metadata-procedure -->
## 元数据和存储过程

| 方法 | 说明 |
| --- | --- |
| `GetDBMetas()` | 查询当前数据库表列表，支持 mysql/postgres |
| `GetTableMetas(tableName)` | 查询表字段信息，支持 mysql/postgres |
| `CallProcedure(procName, args...)` | 调用存储过程并返回多结果集 |

`CallProcedure` 会校验存储过程名称，只允许字母、数字、下划线和点，降低 SQL 注入风险。
<!-- @end-section -->

<!-- @section: helpers -->
## 辅助错误处理

模块提供一组结果检查函数：

| 函数 | 空结果错误 |
| --- | --- |
| `DoGet` | `ErrGetEmpty` |
| `DoUpdate` | `ErrUpdatedEmpty` |
| `DoInsert` | `ErrInsertedEmpty` |
| `DoDelete` | `ErrDeletedEmpty` |

示例：

<!-- @code: helper-example -->
```go
err := databases.DoUpdate(func() (int64, error) {
    return dao.UpdateById(id, user)
})
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: notes -->
## 注意事项

- `GetDatabaseManager` 使用 `sync.Once`，连接创建只会执行一次。
- `Init` 支持重复解析配置，但如果 manager 已经创建，后续配置变更不会重建连接。
- `GetDBDao(name)` 找不到实例时返回 `nil`，业务代码需要检查。
- `FindById` / `FindOne` 未找到记录时返回 `(false, nil)`，不会直接返回 `gorm.ErrRecordNotFound`。
- `NewSession().Close()` 会在事务仍未结束时执行回滚。
- `ShowSql` 当前配置结构中存在，但连接创建代码没有显式应用 GORM logger。
- `UseMasterSlave=true` 且 `slaves` 为空时会退化为仅 master 连接。
<!-- @end-section -->

## 相关文档

- [[reference-component-databases-001]] — DatabasesComponent 使用说明
- [[guide-app-components-module-001]] — app/components/module 最小实践
- [[reference-docs-index-table]] — 文档索引表
