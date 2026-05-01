# register 参考

`register` 负责服务注册、发现、监听和网关同步。它既可以直连 etcd，
也可以对接 APISIX 做服务路由管理。

## 什么时候读这个文件

- 需要把服务注册到 etcd
- 需要发现其他服务或监听实例变化
- 需要把服务同步到 APISIX 网关

## 基本数据

`ServiceInfo` 包含：

- `Name`
- `ID`
- `Address`
- `Port`
- `Tags`
- `Meta`

## etcd 注册

推荐显式创建注册器：

```go
reg, err := register.NewEtcdRegister(etcdBytes)
if err != nil {
    return err
}

registry := register.NewServiceRegistry(reg)
if err := registry.Register("user-service", "127.0.0.1", 9103, []string{"http"}, nil); err != nil {
    return err
}
defer registry.Close()
```

`register.Init(...)` 也可用于全局单例初始化，但显式 `NewEtcdRegister` 更适合业务层控制生命周期。

## 配置

`EtcdConfig` 常用字段：

- `endpoints`
- `dial_timeout`
- `ttl`
- `service_name`
- `address`
- `port`
- `tags`

## 常用能力

`ServiceRegistry` 提供：

- `Register`
- `Deregister`
- `Discover`
- `Watch`
- `Close`

`Register` 会生成类似 `serviceName-sessionId` 的服务 ID。

## 使用流程

```go
services, err := registry.Discover("user-service")
ch, err := registry.Watch("user-service")
```

`Watch` 会先推送当前列表，再持续监听变更。

## 网关

APISIX 相关接口通过 `Gateway` / `GatewayService` / `GatewayRoute` 抽象。
适合把 upstream 和 routes 统一管理在注册层。

## 注意事项

- etcd key 格式是 `/services/{serviceName}/{serviceID}`
- `Deregister` 需要先取消上下文并注销服务
- `Close` 会先尝试 deregister，再关闭底层连接

## 验证

- `go test ./register`
- `go test ./http_server`
