# http_server 参考

`http_server` 负责 Gin HTTP 服务器、gRPC 服务器和 WebSocket 客户端管理。

## 什么时候读这个文件

- 需要使用 `components.HTTPServerComponent(...)`
- 需要在 app 里挂 HTTP group、middleware 或 handler
- 需要注册 WebSocket 路由并管理连接
- 需要使用 `GrpcServerComponent(...)` 或 `NewMultiServer(...)`

## 推荐接入方式

```go
myApp := app.New(conf.Get).WithHTTPGroup("v1", middleware.Authz())

myApp.Use(
    components.LogsComponent(),
    components.TracingComponent(),
    components.HTTPServerFromApp(myApp, setupHTTP),
    components.GRPCServerComponent(setupGRPC),
)
```

`HTTPServerFromApp` 适合把中间件组集中放在 app 层。
`HTTPServerComponent` 适合在 setup 函数里手动 `AddGroup(...)`。
`TracingComponent` 对 HTTP 服务是可选项；只有挂载 `middleware.Tracing(...)`
或需要全局 tracer provider 时才显式注册。

## HTTP 服务器

`HttpServerConfig` 常用字段：

- `port`
- `address`
- `path`
- `cors`
- `request_log`
- `access`
- `worker_id`
- `mode`

常见入口：

- `NewHttpServer(configBytes)`
- `Handle(method, path, handler)`
- `AddGroup(path, middleware...)`
- `Get` / `Post` / `Put` / `Patch` / `Delete` / `Head` / `Options` / `Connect` / `Trace`
- `AddNativeHandler(method, path, handler)`
- `Startup()` / `Stop()`

## 接口开发规范

- 业务接口通过 `HttpServer.Get/Post/...` 注册，不直接使用 `Engine().GET` 或 Gin `RouterGroup`。
- 业务 handler 使用 `httpServer.NewHandler(...)` 创建 `IHandler`，普通接口返回 `(Resp, error)`。
- `NewHandler` 必须使用明确的请求和响应类型。不要用 `any`、`map[string]any`、`interface{}` 或匿名结构体充当请求/响应模型；无请求体时使用 `struct{}`，有请求/响应时在 `dto.go` 中声明清晰的 `XxxReq` / `XxxResp` 类型。
- `NewHandler` 会统一绑定请求并调用 `httpServer.Response(...)` 输出 `protocol.go` 中的 `BaseResponse`。
- 业务代码不要直接调用 `c.JSON` / `gin.H`；确需自定义时实现 `IHandler`，并在 `GetFunc()` 内调用 `httpServer.Response(...)`。
- HTTP 模块目录结构参考 `module.go` / `handler.go` / `dto.go` / `service.go` / `model.go` 分层。
- 推荐参考正式文档 `docs/reference/reference-http-server-api-usage-001.md`。

## 统一返回

`http_server/protocol.go` 定义统一响应结构：

```go
type BaseResponse struct {
    ErrCode int    `json:"errCode"`
    ErrMsg  string `json:"errMsg,omitempty"`
    Data    any    `json:"data,omitempty"`
}
```

`NewHandler` 的返回规则：

| 处理结果 | 响应 |
| --- | --- |
| 返回 `resp, nil` | `errCode=0`，data 为 resp |
| 参数绑定失败 | `errCode=40000` |
| 返回 `*errors.StackError` | 保留业务错误码和错误信息 |
| 返回普通 `error` | 转换为 `errCode=50000` |

确需自定义 `IHandler`，`GetFunc()` 内也必须调用 `httpServer.Response(c, err, data)`，不直接使用 Gin JSON API。

## Handler

```go
srv.Get("user/:id", "v1", httpServer.NewHandler(
    "user/:id",
    []string{"user"},
    func(c *gin.Context, _ struct{}) (Resp, error) {
        return Resp{}, nil
    },
))
```

`NewHandler` 会自动完成 `ShouldBind` 和 JSON 响应包装。

推荐写法：

```go
type GetUserResp struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

srv.Get("user/:id", "v1", httpServer.NewHandler(
    "user/:id",
    []string{"user"},
    func(c *gin.Context, _ struct{}) (GetUserResp, error) {
        return GetUserResp{ID: 1, Name: "alice"}, nil
    },
))
```

避免写法：

```go
srv.Get("user/:id", "v1", httpServer.NewHandler(
    "user/:id",
    []string{"user"},
    func(c *gin.Context, _ any) (map[string]any, error) {
        return map[string]any{"id": 1, "name": "alice"}, nil
    },
))
```

## WebSocket

WebSocket 相关能力主要在这几个类型里：

- `NewWebSocketHandler`
- `WebSocketOptions`
- `IClient`
- `Client`
- `IClientManager`
- `ClientManager`

常见流程应通过框架层 `IHandler` 注册，不直接使用 `srv.Engine().GET`：

```go
manager := httpServer.NewClientManager(logger)
go manager.Start()

srv.Get("ws", "v1", httpServer.NewWebSocketHandler(
    "ws",
    httpServer.WebSocketOptions{
        Codec:   codec.NewJsonCodec(),
        Logger:  logger,
        Manager: manager,
        Handler: handler,
    },
))
```

`NewWebSocketHandler` 内部负责 `websocket.Upgrader`、连接升级、`uuid.GenSessionId()`、
`NewClient(...)`、`RegisterClient(...)` 和 `client.Listen()`。默认从 query
参数 `userId` 读取用户 ID，默认 `CheckOrigin` 返回 `true`；如需覆盖可在
`WebSocketOptions` 中设置 `UserID`、`SessionID`、`CheckOrigin`。

## WebSocket 客户端

`Client`：`SendMessage` / `ReceivedMessage` / `Listen` / `Close`

`ClientManager`：`RegisterClient` / `UnregisterClient` / `BroadcastMessage` / `KickClientByUserId` / `KickClientBySessionId` / `ClientCount` / `ClientKeepLive` / `Stop`

`ClientManager.Start()` 是阻塞循环，需单独 goroutine 启动。

## gRPC

`GrpcServerConfig` 常用字段：

- `listen_on`
- `timeout`
- `etcd`

常见入口：

- `NewGrpcServer(conf)`
- `Server()` 获取底层 `*grpc.Server`
- `Startup()` / `Start()` / `Stop()`

服务注册会在 `EtcdConfig` 存在时自动接入 `register.NewEtcdRegister(...)`。

## 禁止写法

以下写法绕过 `IHandler` 统一绑定和 `Response` 输出，不用于业务接口：

```go
// ❌ 禁止：直接使用 Engine().GET 和 c.JSON
srv.Engine().GET("/api/v1/users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id")})
})

// ❌ 禁止：AddNativeHandler 绕过 IHandler
srv.AddNativeHandler("GET", "users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id")})
})

// ❌ 禁止：直接升级 WebSocket 连接
srv.Engine().GET("/api/v1/ws", func(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    client := httpServer.NewClient(...)
    ...
})
```

## 检查清单

- HTTP 路由是否都集中在模块的 `SetupHTTP` 中注册
- 路由注册是否使用 `srv.Get/Post/Put/Patch/Delete/...`
- handler 是否通过 `httpServer.NewHandler` 或自定义 `IHandler` 暴露
- WebSocket 路由是否通过 `httpServer.NewWebSocketHandler` 暴露
- 业务 handler 是否只返回 `(Resp, error)`，不直接写 JSON
- 响应是否统一经过 `httpServer.Response`
- 目录是否按模块拆分为 `module.go`、`handler.go`、`dto.go`、`service.go`

## 注意事项

- `path` 必须以 `/` 开头
- `HTTPServerFromApp` 会在 app 层先声明中间件组
- WebSocket route 优先使用 `NewWebSocketHandler(...)` 并通过 `srv.Get(...)` 注册
- `ClientManager.Start()` 是阻塞循环，通常单独 goroutine 启动

## 验证

- `go test ./http_server`
- `go test ./http_server/middleware`
