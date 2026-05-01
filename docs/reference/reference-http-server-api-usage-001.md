---
id: "reference-http-server-api-usage-001"
title: "http_server 接口使用参考"
aliases: ["HTTP接口使用参考", "IHandler接口规范", "http_server API规范"]
type: "reference"
category: "backend/library"
tags: ["http_server", "http", "ihandler", "response", "api", "module"]
version: "1.1.0"
created: "2026-04-30"
updated: "2026-04-30"
author: "jxncyjq"
status: "published"
parent: "reference-component-http-server-001"
children: []
related_docs:
  - id: "reference-component-http-server-001"
    relation: "depends_on"
    path: "./components/reference-component-http-server-001.md"
  - id: "reference-component-http-server-from-app-001"
    relation: "related_to"
    path: "./components/reference-component-http-server-from-app-001.md"
  - id: "guide-app-components-module-001"
    relation: "related_to"
    path: "../guides/guide-app-components-module-001.md"
---

# http_server 接口使用参考

<!-- @section: overview -->
## 概述

业务 HTTP 接口应通过 `http_server` 包提供的框架封装注册和返回响应。业务代码不直接使用 Gin 的 `Engine`、`RouterGroup`、`Context.JSON` 或 `gin.H` 组织接口；Gin 只作为框架内部实现和 handler 上下文类型存在。

推荐入口是 `components.HTTPServerFromApp(app, setupFn)` 配合业务模块的 `SetupHTTP(*httpServer.HttpServer)`。路由 handler 必须以 `httpServer.IHandler` 形式交给框架，由框架统一完成参数绑定、错误转换和响应输出。
<!-- @end-section -->

<!-- @section: rules -->
## 基本规则

- 路由注册使用 `srv.Get`、`srv.Post`、`srv.Put`、`srv.Patch`、`srv.Delete`、`srv.Head`、`srv.Options`、`srv.Connect`、`srv.Trace`。
- 业务接口使用 `httpServer.NewHandler(...)` 创建 `IHandler`；只有框架适配层需要自定义 `IHandler`。
- WebSocket 接口使用 `httpServer.NewWebSocketHandler(...)` 创建 `IHandler`，再通过 `srv.Get(...)` 注册。
- 响应统一使用 `http_server/protocol.go` 中的 `httpServer.Response` 输出 `BaseResponse`，业务代码不直接调用 `c.JSON`。
- `NewHandler` 已经在 `GetFunc()` 内调用 `ShouldBind` 和 `Response`；普通 CRUD 接口只需要返回 `(Resp, error)`。
- 业务错误优先返回 `*errors.StackError`，普通 `error` 会被框架转换为通用错误码。
- `srv.Engine()`、`AddNativeHandler(...)` 和 Gin 原生路由只用于极少数基础设施适配场景，不用于业务接口。
<!-- @end-section -->

<!-- @section: directory-structure -->
## 推荐目录结构

HTTP 接口按业务模块组织，`main` 只做组件装配，不写业务路由细节：

<!-- @code: directory-structure -->
```text
internal/user/
├── module.go      # UserModule，聚合 service，并实现 SetupHTTP
├── handler.go     # HTTP handler 方法，返回业务 Resp 和 error
├── dto.go         # 请求 / 响应结构体
├── service.go     # 业务逻辑接口和实现
└── model.go       # 数据模型，仅在模块需要直接声明模型时保留
```
<!-- @end-code -->

目录职责：

| 文件 | 职责 |
| --- | --- |
| `module.go` | 连接 app/components 与业务模块，集中声明 HTTP 路由 |
| `handler.go` | 实现 `NewHandler` 调用的业务处理函数，不直接写 Gin 响应 |
| `dto.go` | 定义请求参数和响应数据结构 |
| `service.go` | 业务逻辑、DAO 调用、事务编排 |
| `model.go` | 数据库模型或领域对象 |
<!-- @end-section -->

<!-- @section: app-wiring -->
## 应用装配

在 app 层预声明中间件组，再用 `HTTPServerFromApp` 自动应用中间件组。`SetupHTTP` 只负责注册业务路由。

<!-- @code: app-wiring -->
```go
myApp := app.New(conf.Get).
    WithHTTPGroup("v1",
        middleware.Metrics(conf.GetAppName()),
        middleware.Timeout(0),
        middleware.Authz(),
    )

userModule := user.NewModule(userService)

businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
).WithDependencies("logs", "databases")

myApp.Use(
    components.LogsComponent(),
    components.DatabasesComponent(),
    components.AuthzComponent(),
    businessComponent,
    components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
)
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: handler -->
## IHandler 写法

普通接口使用 `httpServer.NewHandler` 创建 `IHandler`。处理函数只返回业务响应和错误，不负责写 HTTP 响应。

<!-- @code: handler-example -->
```go
type UserModule struct {
    service UserService
}

func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
    srv.Get("users/:id", "v1", httpServer.NewHandler(
        "users/:id",
        []string{"user"},
        m.GetUser,
    ))

    srv.Post("users", "v1", httpServer.NewHandler(
        "users",
        []string{"user"},
        m.CreateUser,
    ))
}

type CreateUserReq struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UserResp struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (m *UserModule) GetUser(c *gin.Context, _ struct{}) (UserResp, error) {
    id := c.Param("id")
    user, err := m.service.GetUser(c.Request.Context(), id)
    if err != nil {
        return UserResp{}, err
    }
    return UserResp{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

func (m *UserModule) CreateUser(c *gin.Context, req CreateUserReq) (UserResp, error) {
    user, err := m.service.CreateUser(c.Request.Context(), req.Name, req.Email)
    if err != nil {
        return UserResp{}, err
    }
    return UserResp{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}
```
<!-- @end-code -->

`*gin.Context` 只作为框架传入的上下文使用。业务 handler 可以读取 path/query/header/context，但不要调用 `c.JSON`、`c.AbortWithStatusJSON` 或直接注册 Gin 路由。
<!-- @end-section -->

<!-- @section: websocket -->
## WebSocket 接口

WebSocket 路由也走框架 `IHandler`。业务模块只声明路由和消息处理器，不直接调用 `srv.Engine().GET`、`websocket.Upgrader` 或手工注册 `Client`。

<!-- @code: websocket-handler -->
```go
type ChatModule struct {
    logger  *zap.Logger
    manager httpServer.IClientManager
    handler codec.IMessageProcessor
}

func NewChatModule(logger *zap.Logger, handler codec.IMessageProcessor) *ChatModule {
    manager := httpServer.NewClientManager(logger)
    go manager.Start()

    return &ChatModule{
        logger:  logger,
        manager: manager,
        handler: handler,
    }
}

func (m *ChatModule) SetupHTTP(srv *httpServer.HttpServer) {
    srv.Get("ws", "v1", httpServer.NewWebSocketHandler(
        "ws",
        httpServer.WebSocketOptions{
            Codec:   codec.NewJsonCodec(),
            Logger:  m.logger,
            Manager: m.manager,
            Handler: m.handler,
        },
    ))
}
```
<!-- @end-code -->

`NewWebSocketHandler` 内部负责连接升级、生成 session ID、创建 `Client`、注册到 `ClientManager` 并启动监听。默认行为：

| 选项 | 默认值 |
| --- | --- |
| `Codec` | `codec.NewJsonCodec()` |
| `Logger` | `logs.GetLogger("websocket")` |
| `UserID` | `c.Query("userId")` |
| `SessionID` | `uuid.GenSessionId()` |
| `CheckOrigin` | 允许所有 origin |

`Manager` 和 `Handler` 必须显式传入。升级前发现配置缺失时，框架使用 `httpServer.Response` 返回统一错误；升级失败后只记录日志，因为连接协议已经进入 WebSocket 握手流程。
<!-- @end-section -->

<!-- @section: response -->
## 统一返回

`http_server/protocol.go` 定义统一响应结构：

<!-- @code: response-shape -->
```go
type BaseResponse struct {
    ErrCode int    `json:"errCode"`
    ErrMsg  string `json:"errMsg,omitempty"`
    Data    any    `json:"data,omitempty"`
}
```
<!-- @end-code -->

`NewHandler` 的返回规则：

| 处理结果 | 响应 |
| --- | --- |
| 返回 `resp, nil` | `Response(c, nil, resp)`，`errCode=0` |
| 参数绑定失败 | `Response(c, errors.New(err.Error(), 40000), nil)` |
| 返回 `*errors.StackError` | 保留业务错误码和错误信息 |
| 返回普通 `error` | 转换为 `errors.New(err.Error(), 50000)` |

如果确实需要自定义 `IHandler`，也必须调用 `httpServer.Response(c, err, data)` 生成响应，不直接使用 Gin JSON API。
<!-- @end-section -->

<!-- @section: custom-ihandler -->
## 自定义 IHandler

大多数接口不需要自定义 `IHandler`。只有在框架内置 `NewHandler` 无法表达时，才实现 `IHandler`：

<!-- @code: custom-ihandler -->
```go
type HealthHandler struct{}

func (h *HealthHandler) GetName() string { return "health" }
func (h *HealthHandler) GetTags() []string { return []string{"system"} }
func (h *HealthHandler) GetFunc() gin.HandlerFunc {
    return func(c *gin.Context) {
        httpServer.Response(c, nil, map[string]string{"status": "ok"})
    }
}
```
<!-- @end-code -->

自定义实现仍然只通过 `srv.Get/Post/...` 注册：

<!-- @code: custom-ihandler-register -->
```go
srv.Get("health", "", &HealthHandler{})
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: anti-patterns -->
## 禁止写法

以下写法不作为业务接口规范：

<!-- @code: anti-patterns -->
```go
srv.Engine().GET("/api/v1/users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id")})
})

srv.Engine().GET("/api/v1/ws", func(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    client := httpServer.NewClient(c.Query("userId"), uuid.GenSessionId(), conn, jsonCodec, logger, c.Request.Context(), handler, manager)
    manager.RegisterClient(client)
    client.Listen()
})

srv.AddNativeHandler("GET", "users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id")})
})
```
<!-- @end-code -->

问题在于它绕过 `IHandler`、统一绑定、统一错误转换和 `Response` 输出，后续替换底层 HTTP 框架时迁移成本更高。
<!-- @end-section -->

<!-- @section: checklist -->
## 检查清单

- HTTP 路由是否都集中在模块的 `SetupHTTP` 中注册。
- 路由注册是否使用 `srv.Get/Post/Put/Patch/Delete/...`。
- handler 是否通过 `httpServer.NewHandler` 或自定义 `IHandler` 暴露。
- WebSocket 路由是否通过 `httpServer.NewWebSocketHandler` 暴露。
- 业务 handler 是否只返回 `(Resp, error)`，不直接写 JSON。
- 响应是否统一经过 `httpServer.Response`。
- 目录是否按模块拆分为 `module.go`、`handler.go`、`dto.go`、`service.go`。
<!-- @end-section -->

## 相关文档

- [[reference-component-http-server-001]]
- [[reference-component-http-server-from-app-001]]
- [[guide-app-components-module-001]]
