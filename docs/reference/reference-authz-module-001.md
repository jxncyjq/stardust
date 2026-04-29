---
id: "reference-authz-module-001"
title: "authz 模块使用参考"
aliases: ["authz模块", "Casbin授权", "权限控制"]
type: "reference"
category: "backend/library"
tags: ["authz", "casbin", "authorization", "middleware", "rbac"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-29"
author: "jxncyjq"
status: "published"
parent: "reference-component-authz-001"
children: []
related_docs:
  - id: "reference-component-authz-001"
    relation: "depends_on"
    path: "./components/reference-component-authz-001.md"
---

# authz 模块使用参考

<!-- @section: overview -->
## 概述

`authz` 模块基于 Casbin v3 提供授权能力。它只负责“已认证主体是否允许访问资源”，不负责认证本身。认证通常由 `middleware.Access()` 或其他登录态中间件完成。
<!-- @end-section -->

<!-- @section: config -->
## 配置

核心配置类型是 `authz.Config`：

<!-- @code: config-struct -->
```go
type Config struct {
    ModelPath  string
    PolicyPath string
    Adapter    string
    SubjectKey string
    ObjectMode string
    ActionMode string
}
```
<!-- @end-code -->

默认值：

| 字段 | 默认值 |
| --- | --- |
| `adapter` | `file` |
| `subject_key` | `id` |
| `object_mode` | `route` |
| `action_mode` | `method` |

默认模式仍是 file policy。`object_mode` 支持 `route` 和 `path`，`action_mode` 只支持 `method`。
同时，`authz` 也支持基于数据库的 Casbin GORM adapter：

- `adapter = "gorm"` 时，`policy_path` 不再必填。
- 如果配置了 `database_name`，组件会优先复用 `[databases]` 里的同名数据库实例。
- 如果没有 `database_name`，则可以直接配置 `driver` 和 `dsn`，由 `authz` 自己创建 GORM 连接。

下面是一份可直接用于 `example/config.toml` 的完整配置示例：

<!-- @code: config-example -->
```toml
[authz]
# Casbin 模型文件，定义 request / policy / matcher / effect
# 示例可放在 ./example/config/casbin/model.conf
model_path = "./example/config/casbin/model.conf"

# Casbin 策略文件，按 CSV 格式编写授权规则
# 示例可放在 ./example/config/casbin/policy.csv
policy_path = "./example/config/casbin/policy.csv"

# 当前首版仅支持 file 适配器
# 这里保留字段是为了后续扩展数据库、远程存储等适配器
adapter = "file"

# 从请求中读取主体标识的键名
# 通常由 middleware.Access() 写入 gin.Context
# 默认值是 id
subject_key = "id"

# 授权对象的生成方式
# route: 使用 gin 的 FullPath()，例如 /api/v1/users/:id
# path: 使用请求真实路径，例如 /api/v1/users/123
object_mode = "route"

# 授权动作的生成方式
# 当前首版只支持 method，对应 HTTP 方法，例如 GET / POST / PUT
action_mode = "method"
```
<!-- @end-code -->

数据库模式示例：

<!-- @code: config-example-gorm -->
```toml
[authz]
model_path = "./example/config/casbin/model.conf"
adapter = "gorm"

# 方式一：复用 [databases] 里的同名连接
# 这个名字必须和 [databases] 里的 name 一致
database_name = "main"

# 方式二：不依赖 [databases]，直接给出 GORM driver 和 DSN
# driver = "mysql"
# dsn = "root:password@tcp(127.0.0.1:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Local"

# gorm 模式下 policy_path 不是必需项，保留也不会使用
# policy_path = "./example/config/casbin/policy.csv"
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `model_path` | Casbin 模型文件路径，不能为空。 |
| `policy_path` | Casbin 策略文件路径，`file` 模式必填，`gorm` 模式可不填。 |
| `adapter` | 适配器类型，支持 `file` 和 `gorm`。 |
| `database_name` | `gorm` 模式下从 `[databases]` 中选择的数据库实例名。 |
| `driver` | `gorm` 模式下直接连接数据库时使用的驱动名，例如 `mysql`。 |
| `dsn` | `gorm` 模式下直接连接数据库时使用的 DSN。 |
| `subject_key` | 从 `gin.Context` 中读取主体 ID 的键名。 |
| `object_mode` | 授权对象提取方式，`route` 或 `path`。 |
| `action_mode` | 授权动作提取方式，当前仅支持 `method`。 |

### model.conf 语义

`model.conf` 是 Casbin 的模型定义文件，决定“请求是什么、策略长什么样、怎么匹配”。

在本项目里，推荐使用下面这一套固定结构：

<!-- @code: model-conf-example -->
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```
<!-- @end-code -->

每一段的含义如下：

| 段名 | 说明 |
| --- | --- |
| `request_definition` | 定义一次授权请求携带哪些信息。这里的 `r` 代表 request，`sub` 是主体，`obj` 是资源对象，`act` 是动作。 |
| `policy_definition` | 定义策略规则的字段结构。这里的 `p` 代表 policy，字段顺序必须和 `policy.csv` 一致。 |
| `policy_effect` | 定义策略命中后如何做决策。`some(where (p.eft == allow))` 表示只要有一条 allow 规则命中就放行。 |
| `matchers` | 定义请求和策略的匹配表达式。当前示例采用严格相等匹配，适合路由级 RBAC。 |

这份模型文件对应的授权含义是：

- `sub`：谁在访问，通常是登录后的用户 ID。
- `obj`：访问哪个资源，默认使用 Gin 的 `FullPath()`，例如 `/api/v1/users/:id`。
- `act`：执行什么动作，默认使用 HTTP 方法，例如 `GET`、`POST`。

如果你把 `object_mode` 改成 `path`，那么 `obj` 会变成请求真实路径，例如 `/api/v1/users/123`，这时 `policy.csv` 里也必须写真实路径，不能再写路由模板。

### policy_effect 语义

`policy_effect` 决定“当多条策略都命中时，最终要不要放行，以及如何合并结果”。

当前示例使用的是：

<!-- @code: policy-effect-allow-override -->
```ini
[policy_effect]
e = some(where (p.eft == allow))
```
<!-- @end-code -->

这表示 allow-override，也就是“只要有任意一条命中的策略允许，就放行”。
在最简单的 RBAC/ACL 场景里，这通常是最容易理解、也最容易落地的选择。

示例：

- 规则 A：`roleName` 可以访问 `/orders`
- 规则 B：`roleName` 也可以访问 `/reports`
- 请求：`roleName` 访问 `/orders`

只要规则 A 命中，结果就是允许。规则 B 是否命中，不影响这次结果。

如果你需要更严格的拒绝模型，可以改成 deny-override：

<!-- @code: policy-effect-deny-override -->
```ini
[policy_effect]
e = !some(where (p.eft == deny))
```
<!-- @end-code -->

这表示“只要有任何一条 deny 命中，就拒绝；否则放行”。
要使用这种模式，`policy_definition` 也要扩展出 `eft` 字段：

<!-- @code: policy-definition-with-eft -->
```ini
[policy_definition]
p = sub, obj, act, eft
```
<!-- @end-code -->

对应的 `policy.csv` 也要多一列：

<!-- @code: policy-csv-deny-example -->
```csv
p, roleName, /orders, GET, allow
p, roleName, /orders, POST, deny
```
<!-- @end-code -->

这时：

- `roleName` 访问 `/orders` 的 `GET`，会被允许
- `roleName` 访问 `/orders` 的 `POST`，会被拒绝

如果你想表达“既要有 allow，又不能有 deny”，也可以使用 allow-and-deny：

<!-- @code: policy-effect-allow-and-deny -->
```ini
[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))
```
<!-- @end-code -->

这类模型适合需要显式授权、同时还要保留黑名单拒绝能力的场景。

`priority(p.eft) || deny` 适合按优先级决策的模型，常见于你想让更具体的规则覆盖更宽泛的规则时。
`subjectPriority(p.eft)` 则用于基于主体优先级的决策，通常和角色树、层级角色一起使用。

在当前 stardust 示例里，我们没有把这些高级 effect 做成默认配置，是因为最小实践场景更适合先用 allow-override，把“能不能匹配上”讲清楚。

### matchers 语义

`matchers` 决定“请求和策略怎么匹配”。

当前示例使用的是最简单的严格相等匹配：

<!-- @code: matcher-exact -->
```ini
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```
<!-- @end-code -->

这表示：

- 主体必须相同
- 资源必须相同
- 动作必须相同

它适合路由级 RBAC，原因是规则简单，行为可预测，和 `middleware.Access()` / `middleware.Authz()` 的接线方式也最直接。

几个常见变体如下。

第一种，角色匹配：

<!-- @code: matcher-rbac -->
```ini
[role_definition]
g = _, _

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```
<!-- @end-code -->

这表示 `r.sub` 不再直接和 `p.sub` 比较，而是先判断主体是否属于某个角色。
例如 `roleName` 属于 `admin` 角色时，只要策略写的是 `p, admin, /orders, GET`，`roleName` 也会被允许。

第二种，路径前缀匹配：

<!-- @code: matcher-prefix -->
```ini
[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act
```
<!-- @end-code -->

这类写法适合资源层级明显的 API。
例如策略写成 `/api/v1/users/:id`，请求路径是 `/api/v1/users/123` 时，也能匹配上。
你当前的示例默认用 `route` 模式，其实就是为了配合这类模板式匹配。

第三种，动作放宽：

<!-- @code: matcher-action-all -->
```ini
[matchers]
m = r.sub == p.sub && r.obj == p.obj
```
<!-- @end-code -->

这表示同一个主体对同一个资源的所有动作都统一授权。
它简单，但通常过于宽泛，适合很小的内部系统，不适合精细化权限控制。

第四种，组合匹配：

<!-- @code: matcher-composite -->
```ini
[matchers]
m = (g(r.sub, p.sub) || r.sub == p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```
<!-- @end-code -->

这种写法可以同时支持角色授权、路径模板和动作表达式。
但匹配条件越复杂，规则越难维护，也越需要明确约定 `policy.csv` 每一列写什么。

简化理解：

- `policy_effect` 负责“多条命中后怎么裁决”
- `matchers` 负责“这一条命中不命中”

前者决定结果聚合，后者决定单条规则是否成立。

### RBAC 执行流程

如果你希望用 `admin` 这类角色来做授权，推荐把“用户”和“角色”分开写。

一个更标准的 RBAC 例子如下：

<!-- @code: rbac-model-example -->
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```
<!-- @end-code -->

假设你当前登录的是 `adminUser`，系统给它分配了 `admin` 角色，那么一次 `GET /api/v1/user` 的执行流程通常是这样的：

1. `middleware.Access()` 先识别请求主体，把 `adminUser` 写入 `gin.Context`。
2. `middleware.Authz()` 读取主体、路由和方法。
   - `sub = adminUser`
   - `obj = /api/v1/user`，当 `object_mode = "route"` 时通常来自 `c.FullPath()`
   - `act = GET`
3. Casbin 先判断角色关系：
   - `g(adminUser, admin)` 是否为真
4. Casbin 再判断策略是否命中：
   - `p, admin, /api/v1/user, GET`
5. 如果角色关系和策略规则都成立，就放行；否则返回 403。

对应的策略文件可以写成这样：

<!-- @code: rbac-policy-example -->
```csv
g, adminUser, admin
p, admin, /api/v1/user, GET
```
<!-- @end-code -->

这个流程里，`adminUser` 是“用户主体”，`admin` 是“角色名”。  
`roleName` 这种写法只适合做主体标识示例，不代表自动启用 RBAC。真正的 RBAC 检查必须依赖 `g(...)`。

### policy.csv 语义

`policy.csv` 是 Casbin 的策略数据文件，描述“谁能做什么”。
如果是 RBAC 模型，这个文件里通常会同时出现：

- `p`：权限规则，描述“哪个角色能做什么”
- `g`：分组关系，描述“哪个用户属于哪个角色”

在本项目里，最小可用格式是：

<!-- @code: policy-csv-example -->
```csv
p, roleName, /api/v1/user/:id, GET
p, roleName, /api/v1/game/:id, GET
p, roleName, /api/v1/game/score, POST
```
<!-- @end-code -->

每一列的含义如下：

| 列 | 说明 |
| --- | --- |
| 第 1 列 | 固定为 `p`，表示一条策略记录。 |
| 第 2 列 | `sub`，主体标识，也就是允许访问的用户或角色。 |
| 第 3 列 | `obj`，资源对象，必须和 `object_mode` 的取值一致。 |
| 第 4 列 | `act`，动作，通常是 HTTP 方法。 |

上面的示例表示：

- `roleName` 可以 `GET /api/v1/user/:id`
- `roleName` 可以 `GET /api/v1/game/:id`
- `roleName` 可以 `POST /api/v1/game/score`

如果你采用角色模型，也可以把第二列写成角色名，例如 `admin`、`operator`，但前提是你的 Casbin 模型和 matcher 也要随之调整。当前这版示例是最简单的主体直配模式，便于和 `middleware.Access()` 的 `id` 字段直接对接。
<!-- @end-section -->

<!-- @section: init -->
## 初始化

推荐通过组件初始化：

<!-- @code: init-component -->
```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.AuthzComponent(),
    ).
    Run(context.Background())
```
<!-- @end-code -->

也可以直接初始化：

<!-- @code: init-manual -->
```go
if err := authz.Init(conf.Get("authz")); err != nil {
    return err
}
```
<!-- @end-code -->

`authz.Init` 接收 JSON 编码的配置字节；空配置或非法 JSON 会返回 `ErrInvalidConfig`。

如果你使用的是 `adapter = "gorm"` 且配置了 `database_name`，更推荐通过 `components.AuthzComponent()` 启动，这样组件会先读取 `[databases]` 配置并完成数据库初始化。
<!-- @end-section -->

<!-- @section: usage -->
## 使用方式

包级全局授权：

<!-- @code: enforce -->
```go
allowed, err := authz.Enforce(ctx, "roleName", "/api/v1/users/:id", http.MethodGet)
```
<!-- @end-code -->

Gin 中间件：

<!-- @code: middleware -->
```go
myApp.WithHTTPGroup("v1", middleware.Access(), middleware.Authz())
```
<!-- @end-code -->

`middleware.Authz()` 会优先读取 `gin.Context` 中的 `id`，对象默认使用 `c.FullPath()`，方法默认使用 `c.Request.Method`。
<!-- @end-section -->

## 相关文档

- [[reference-component-authz-001]]
- [[reference-docs-index]]
