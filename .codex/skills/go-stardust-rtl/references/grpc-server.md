# grpc_server 参考

通过 `components.GRPCServerComponent(setupFn)` 创建并启动 gRPC 服务器。

## 什么时候读这个文件

- 需要注册 protobuf gRPC 服务
- 需要将 gRPC server 接入 app 生命周期

## 推荐接入方式

直接注册 proto 服务：

```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.TracingComponent(),
        components.GRPCServerComponent(func(s *grpc.Server) {
            pb.RegisterUserServiceServer(s, &UserServiceImpl{})
        }),
    ).
    Run(context.Background())
```

通过 `BusinessComponent` 自动分发（推荐多模块场景）：

```go
businessComponent := components.Business(service.NewServiceGroup(), userModule)
myApp.Use(
    businessComponent,
    components.GRPCServerComponent(businessComponent.SetupGRPC),
)
```

## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.GRPCServerComponent(setupFn)` |
| 组件名 | `grpc_server` |
| 依赖 | `logs`, `tracing` |
| 配置 key | `grpc_server` |
| Init | 创建 gRPC server，注入拦截器链，执行 proto 注册 |
| Start | `srv.Startup()` |
| Stop | `srv.Stop()` |

内置拦截器链：metric → tracing → breaker → timeout（由 `NewGrpcServer` 自动注入）。

## 配置

```toml
[grpc_server]
listen_on = "0.0.0.0:9090"
timeout   = 5000
```

`timeout` 单位毫秒。`[etcd]` 块存在时自动接入 `register.NewEtcdRegister(...)` 进行服务注册。

## 常见入口

- `NewGrpcServer(conf)` — 创建 server
- `srv.Server()` — 获取底层 `*grpc.Server` 用于注册 proto 服务
- `srv.Startup()` / `srv.Stop()`

## 验证

- `go test ./http_server`
- `go test ./app/components`
