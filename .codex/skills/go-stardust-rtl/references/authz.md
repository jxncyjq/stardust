# 权限授权参考

## 适用场景

当你需要把 Casbin 授权接入 stardust 时，先看这里。它覆盖：

- `components.AuthzComponent()` 的组件初始化
- `middleware.Access()` 与 `middleware.Authz()` 的中间件接线
- `file` 与 `gorm` 两种策略存储方式
- Casbin `model.conf`、`policy.csv`、`policy_effect`、`matchers`
- RBAC 执行流程

## 组件和中间件

- `components.AuthzComponent()` 读取 `authz` 配置并初始化全局授权器。
- 当 `adapter = "gorm"` 且配置了 `database_name` 时，组件会先初始化 `[databases]`。
- `middleware.Access()` 负责认证，通常先于 `middleware.Authz()` 执行。
- `middleware.Authz()` 在 Gin 路由层执行授权检查。

## 配置

`authz.Config` 支持两种模式：

```toml
[authz]
model_path = "./example/config/casbin/model.conf"
policy_path = "./example/config/casbin/policy.csv"
adapter = "file"
subject_key = "id"
object_mode = "route"
action_mode = "method"
```

数据库模式示例：

```toml
[authz]
model_path = "./example/config/casbin/model.conf"
adapter = "gorm"
database_name = "main"
# 也可以不用 databases 组件，直接写 driver + dsn
# driver = "mysql"
# dsn = "root:password@tcp(127.0.0.1:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Local"
```

说明：

- `file` 模式下，`policy_path` 必填。
- `gorm` 模式下，`policy_path` 可不填。
- `gorm` 模式下，优先用 `database_name` 复用 `[databases]` 中的连接。
- 如果不复用 `databases`，就直接提供 `driver` 和 `dsn`。

## model.conf

推荐的默认模型：

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

语义：

- `sub`：主体，通常来自登录后的用户 ID 或用户名。
- `obj`：资源对象。`object_mode = route` 时通常是 `c.FullPath()`。
- `act`：动作，通常是 HTTP 方法。
- `policy_effect`：多条规则同时命中时如何裁决。
- `matchers`：单条策略是否命中。

### policy_effect

默认 `allow-override`：

```ini
[policy_effect]
e = some(where (p.eft == allow))
```

表示只要有一条 allow 命中就放行。

常见变体：

```ini
[policy_effect]
e = !some(where (p.eft == deny))
```

表示 deny-override，只要有一条 deny 命中就拒绝。

```ini
[policy_definition]
p = sub, obj, act, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))
```

表示同时保留显式 allow 和 deny。

### matchers

最简单的严格相等：

```ini
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

角色匹配：

```ini
[role_definition]
g = _, _

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

路径模板匹配：

```ini
[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act
```

组合匹配：

```ini
[matchers]
m = (g(r.sub, p.sub) || r.sub == p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

## RBAC 执行流程

如果你希望用角色授权，推荐把“用户”和“角色”分开写。

一个更标准的 RBAC 例子如下：

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

假设当前登录的是 `adminUser`，系统把它分配给 `admin` 角色，请求 `GET /api/v1/user` 时：

1. `middleware.Access()` 先识别请求主体，把 `adminUser` 写入 `gin.Context`。
2. `middleware.Authz()` 读取主体、路由和方法。
   - `sub = adminUser`
   - `obj = /api/v1/user`
   - `act = GET`
3. Casbin 先判断角色关系：`g(adminUser, admin)`。
4. 再判断策略是否命中：`p, admin, /api/v1/user, GET`。
5. 命中则放行，否则返回 403。

策略文件通常同时包含：

- `p`：权限规则，描述哪个角色能做什么
- `g`：分组关系，描述哪个用户属于哪个角色

示例：

```csv
g, adminUser, admin
p, admin, /api/v1/user, GET
```

## policy.csv

如果是主体直配模式：

```csv
p, roleName, /api/v1/user, GET
p, roleName, /api/v1/game/:id, GET
p, roleName, /api/v1/game/score, POST
```

如果是 RBAC 模式：

```csv
g, adminUser, admin
p, admin, /api/v1/user, GET
```

## 使用顺序

1. 先通过 `components.AuthzComponent()` 初始化授权器。
2. 再在 HTTP 组中注册 `middleware.Access()` 和 `middleware.Authz()`。
3. 业务 handler 通过 `gin.Context` 的 `id` 字段识别主体。

## 验证

- `go test ./authz`
- `go test ./app/components ./http_server/middleware`
