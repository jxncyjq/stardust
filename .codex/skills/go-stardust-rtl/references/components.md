# 组件参考

## 基本接口

```go
type Component interface {
    Name() string
    Dependencies() []string
    Init(ctx context.Context, configFn ConfigFunc) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

Container 按依赖顺序 Init/Start，Stop 时反向执行。

## 组件表

| 组件 | 名称 | 依赖 | 配置 |
| --- | --- | --- | --- |
| `LogsComponent()` | `logs` | 无 | `logs` |
| `RedisComponent()` | `redis` | `logs` | `redis` |
| `DatabasesComponent()` | `databases` | `logs` | `databases` |
| `MongoDBComponent()` | `mongodb` | `logs` | `mongodb` |
| `ClickhouseComponent()` | `clickhouse` | `logs` | `clickhouse` |
| `NatsComponent()` | `nats` | `logs` | `nats` |
| `TracingComponent()` | `tracing` | `logs` | `tracing` |
| `AuthzComponent()` | `authz` | `logs` | `authz` |
| `I18nComponent()` | `i18n` | `logs` | `i18n` |
| `HTTPServerComponent(setup)` | `http_server` | `logs` | `http_server` |
| `HTTPServerFromApp(app, setup)` | `http_server` | `logs` | `http_server` |
| `GRPCServerComponent(setup)` | `grpc_server` | `logs`, `tracing` | `grpc_server` |
| `Business(manager, services...)` | `business` | 调用方设置 | 无 |

## 推荐 app 写法

```go
myApp := app.New(conf.Get).
    WithHTTPGroup("v1", middleware.I18n(), middleware.Metrics(appName), middleware.Authz())

businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
).WithDependencies("logs", "redis", "databases")

myApp.Use(
    components.LogsComponent(),
    // 使用 middleware.Tracing(...) 时显式注册；纯 HTTP 服务可不注册。
    components.TracingComponent(),
    components.RedisComponent(),
    components.DatabasesComponent(),
    components.I18nComponent(),
    components.AuthzComponent(),
    businessComponent,
    components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
    components.GRPCServerComponent(businessComponent.SetupGRPC),
)
```

## BusinessComponent

业务模块至少实现：

```go
type Service interface {
    Start()
    Stop()
}
```

可选实现：

```go
type HTTPBinder interface { SetupHTTP(*httpServer.HttpServer) }
type GRPCBinder interface { SetupGRPC(*grpc.Server) }
```

生命周期：

- `Init`: `manager.AddMany`
- `Start`: `manager.StartAsync`
- `Stop`: `manager.Stop`

依赖用 `WithDependencies(...)` 设置，不要写死在框架里。

## AuthzComponent

`components.AuthzComponent()` 负责读取 `authz` 配置并初始化 Casbin 授权器。
当 `adapter = "gorm"` 且配置了 `database_name` 时，它会先初始化 `[databases]`，再构建授权器。

配置示例：

```toml
[authz]
model_path  = "./example/config/casbin/model.conf"
policy_path = "./example/config/casbin/policy.csv"
adapter     = "file"
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
# 或者直接写 driver + dsn
# driver = "mysql"
# dsn = "root:password@tcp(127.0.0.1:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Local"
```

HTTP 组中通常配合 `middleware.Access()` 和 `middleware.Authz()` 使用；前者负责认证并写入主体，后者负责授权判定。

## I18nComponent

`components.I18nComponent()` 负责读取 `i18n` 配置并加载嵌入式翻译文件。

配置示例：

```toml
[i18n]
default_language = "zh"
supported_languages = ["zh", "en"]
query_key = "lang"
header_key = "Accept-Language"
```

HTTP 组里通常配合 `middleware.I18n()` 使用；handler 再通过 `i18n.Localize(c.Request.Context(), ...)` 取出本地化消息。

## ServiceGroup

`service.ServiceGroup` 实现 `service.Manager`：

- `Add` / `AddMany`
- `StartAsync`
- `Start`
- `Stop`

`StartAsync` 只启动一次并立即返回。`Stop` 按服务注册顺序反向执行。

## 验证

- `go test ./app/components`
- `go test ./example`
- `go test ./service`
