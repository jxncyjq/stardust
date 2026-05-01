# BusinessComponent 参考

`components.Business(manager, services...)` 把多个业务模块纳入 `app.Component` 生命周期，并自动分发 HTTP/gRPC 路由注册。

## 什么时候读这个文件

- 需要将多个业务模块（service）接入 app 生命周期
- 需要多个模块统一向 HTTPServerFromApp 或 GRPCServerComponent 注册路由

## 推荐接入方式

```go
userModule := &UserModule{}
gameModule := &GameModule{}

businessComponent := components.Business(
    service.NewServiceGroup(),
    userModule,
    gameModule,
).WithDependencies("logs", "redis", "databases")

myApp.Use(
    components.LogsComponent(),
    components.RedisComponent(),
    components.DatabasesComponent(),
    businessComponent,
    components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
    components.GRPCServerComponent(businessComponent.SetupGRPC),
)
```

## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.Business(manager, services...)` |
| 组件名 | `business` |
| 依赖 | 通过 `WithDependencies(...)` 设置 |
| 配置 key | 无 |
| Init | `manager.AddMany(services...)` |
| Start | `manager.StartAsync()` |
| Stop | `manager.Stop()` |

## 业务模块接口

模块必须实现 `service.Service`；可选实现 HTTP/gRPC 绑定：

```go
// 必须实现
type Service interface { Start(); Stop() }

// HTTP 路由绑定（可选）
func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
    srv.Get("user/:id", "v1", httpServer.NewHandler("user/:id", []string{"user"}, m.GetUser))
    srv.Post("user", "v1", httpServer.NewHandler("user", []string{"user"}, m.CreateUser))
}

// gRPC 服务绑定（可选）
func (m *UserModule) SetupGRPC(s *grpc.Server) {
    pb.RegisterUserServiceServer(s, &UserServiceImpl{service: m})
}
```

`businessComponent.SetupHTTP` / `SetupGRPC` 会自动遍历所有实现了对应接口的模块并依次调用。

## 验证

- `go test ./app/components`
- `go test ./service`
