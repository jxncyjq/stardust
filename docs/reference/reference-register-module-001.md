---
id: "reference-register-module-001"
title: "register 模块使用参考"
aliases: ["register模块", "ServiceRegistry", "EtcdRegister", "ApisixGateway"]
type: "reference"
category: "backend/library"
tags: ["register", "etcd", "service-discovery", "apisix", "gateway"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-microservice-module-001"
children: []
related_docs:
  - id: "reference-component-grpc-server-001"
    relation: "related_to"
    path: "./components/reference-component-grpc-server-001.md"
  - id: "reference-microservice-module-001"
    relation: "depends_on"
    path: "./reference-microservice-module-001.md"
---

# register 模块使用参考

<!-- @section: overview -->
## 概述

`register` 模块提供服务注册、服务发现和网关注册抽象。当前包含三层能力：

| 能力 | 类型 | 说明 |
| --- | --- | --- |
| 服务注册抽象 | `Register` / `ServiceRegistry` | 统一注册、注销、发现、Watch 行为 |
| etcd 实现 | `EtcdRegister` | 将服务实例写入 etcd，并用租约维持生命周期 |
| APISIX 网关 | `Gateway` / `ApisixGateway` | 通过 APISIX Admin API 创建 upstream 和 routes |

gRPC server 和 gRPC client 已接入 etcd 注册发现：服务端启动时可注册自身，客户端可通过 resolver watch 服务列表。
<!-- @end-section -->

<!-- @section: service-info -->
## ServiceInfo

服务实例使用 `ServiceInfo` 表示：

<!-- @code: service-info -->
```go
type ServiceInfo struct {
    Name    string            `json:"name"`
    ID      string            `json:"id"`
    Address string            `json:"address"`
    Port    int               `json:"port"`
    Tags    []string          `json:"tags,omitempty"`
    Meta    map[string]string `json:"meta,omitempty"`
}
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `Name` | 服务名，用于发现同一类服务 |
| `ID` | 服务实例 ID，用于注销指定实例 |
| `Address` / `Port` | 服务实例地址 |
| `Tags` | 标签，例如 `grpc`、`http`、`prod` |
| `Meta` | 附加元数据 |
<!-- @end-section -->

<!-- @section: register-interface -->
## Register 接口

`Register` 是注册中心抽象：

<!-- @code: register-interface -->
```go
type Register interface {
    Register(ctx context.Context, info *ServiceInfo) error
    Deregister(ctx context.Context, serviceID string) error
    GetService(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
    Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)
    Close() error
}
```
<!-- @end-code -->

业务一般不直接依赖具体注册中心，而是通过 `ServiceRegistry` 包装 `Register` 实现。
<!-- @end-section -->

<!-- @section: service-registry -->
## ServiceRegistry

`ServiceRegistry` 为注册中心提供统一业务入口：

| 方法 | 说明 |
| --- | --- |
| `NewServiceRegistry(register)` | 创建服务注册管理器 |
| `Register(name, address, port, tags, meta)` | 生成实例 ID 并注册服务 |
| `Deregister()` | 注销当前实例 |
| `Discover(serviceName)` | 查询服务实例列表 |
| `Watch(serviceName)` | 监听服务实例变化 |
| `Close()` | 注销当前实例并关闭底层注册器 |

注册示例：

<!-- @code: service-registry-register -->
```go
reg, err := register.NewEtcdRegister(etcdConfigBytes)
if err != nil {
    return err
}
defer reg.Close()

registry := register.NewServiceRegistry(reg)
if err := registry.Register(
    "user-service",
    "127.0.0.1",
    9103,
    []string{"grpc"},
    map[string]string{"version": "1.0.0"},
); err != nil {
    return err
}
defer registry.Deregister()
```
<!-- @end-code -->

发现示例：

<!-- @code: service-registry-discover -->
```go
services, err := registry.Discover("user-service")
if err != nil {
    return err
}
for _, svc := range services {
    _ = fmt.Sprintf("%s:%d", svc.Address, svc.Port)
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: watch -->
## 服务监听

`Watch(serviceName)` 返回服务列表变更通道。etcd 实现会先发送当前服务列表，再监听后续变化：

<!-- @code: watch -->
```go
ch, err := registry.Watch("user-service")
if err != nil {
    return err
}

go func() {
    for services := range ch {
        _ = services
    }
}()
```
<!-- @end-code -->

`ServiceRegistry.Close()` 或底层 context 取消后，watch 通道会结束。
<!-- @end-section -->

<!-- @section: etcd-config -->
## EtcdConfig

etcd 注册配置：

<!-- @code: etcd-config -->
```go
type EtcdConfig struct {
    Endpoints   []string `json:"endpoints" yaml:"endpoints"`
    DialTimeout int      `json:"dial_timeout" yaml:"dial_timeout"`
    TTL         int64    `json:"ttl" yaml:"ttl"`
    ServiceName string   `json:"service_name" yaml:"service_name"`
    Address     string   `json:"address" yaml:"address"`
    Port        int      `json:"port" yaml:"port"`
    Tags        []string `json:"tags" yaml:"tags"`
}
```
<!-- @end-code -->

字段说明：

| 字段 | 说明 |
| --- | --- |
| `endpoints` | etcd 地址列表 |
| `dial_timeout` | 连接超时，单位秒，默认 `5` |
| `ttl` | 服务租约 TTL，单位秒，默认 `10` |
| `service_name` | 要注册或发现的服务名 |
| `address` / `port` | 当前服务实例地址 |
| `tags` | 服务标签 |

TOML 示例：

<!-- @code: etcd-toml -->
```toml
[grpc_server.etcd]
endpoints    = ["127.0.0.1:2379"]
dial_timeout = 5
ttl          = 10
service_name = "user-service"
address      = "127.0.0.1"
port         = 9103
tags         = ["grpc"]
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: etcd-register -->
## EtcdRegister

`EtcdRegister` 使用固定前缀 `/services/` 存储服务实例：

<!-- @code: etcd-key -->
```text
/services/{serviceName}/{serviceID}
```
<!-- @end-code -->

核心行为：

| 方法 | 行为 |
| --- | --- |
| `NewEtcdRegister(configBytes)` | 创建 etcd client 和注册器，不自动注册服务 |
| `Init(configBytes)` | 初始化全局 etcd 单例，并立即注册配置中的服务 |
| `GetEtcdRegister()` | 获取全局 etcd 单例 |
| `Register(ctx, info)` | 创建租约，写入服务信息，并启动 KeepAlive goroutine |
| `Deregister(ctx, serviceID)` | revoke 当前租约 |
| `GetService(ctx, serviceName)` | 按 `/services/{serviceName}/` 前缀查询实例 |
| `Watch(ctx, serviceName)` | 监听前缀变化并推送完整实例列表 |
| `Close()` | 注销服务并关闭 etcd client |

优先使用 `NewEtcdRegister` 创建局部注册器；`Init/GetEtcdRegister` 是全局单例入口，适合简单进程级注册。
<!-- @end-section -->

<!-- @section: grpc-integration -->
## gRPC 集成

`http_server` 包已经使用 `register` 完成 gRPC 服务注册和发现。

服务端：`GrpcServer.Startup()` 在配置 `Etcd` 不为空时，会创建 `EtcdRegister`，并通过 `ServiceRegistry.Register` 注册当前监听地址。

客户端：`NewGrpcClient` 在配置 `Etcd` 不为空时，会注册 etcd resolver，并使用如下 target：

<!-- @code: grpc-target -->
```text
etcd:///{service_name}
```
<!-- @end-code -->

客户端 resolver 会 watch etcd 服务列表，并把服务实例转换成 gRPC `resolver.Address`。

配置示例：

<!-- @code: grpc-client-config -->
```toml
[grpc_client.etcd]
endpoints    = ["127.0.0.1:2379"]
dial_timeout = 5
ttl          = 10
service_name = "user-service"
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: gateway-model -->
## Gateway 抽象

网关服务模型：

| 类型 | 说明 |
| --- | --- |
| `GatewayUpstream` | 上游负载均衡配置，包含 `type`、`nodes`、`scheme` |
| `GatewayRoute` | 路由规则，包含 `uri`、`methods`、`upstream_id` 或路由级 `upstream` |
| `GatewayService` | 一组 upstream 和 routes 的服务注册信息 |
| `Gateway` | 网关注入接口，包含注册、注销和关闭 |

示例：

<!-- @code: gateway-service -->
```go
svc := &register.GatewayService{
    ID:   "user-service",
    Name: "user-service",
    Upstream: &register.GatewayUpstream{
        Type:  "roundrobin",
        Nodes: map[string]int{"127.0.0.1:8080": 1},
    },
    Routes: []*register.GatewayRoute{
        {
            Name:    "user-api",
            URI:     "/api/users/*",
            Methods: []string{"GET", "POST"},
        },
    },
}
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: apisix -->
## APISIX 网关

APISIX 配置：

<!-- @code: apisix-config -->
```go
type ApisixConfig struct {
    AdminURL         string `json:"admin_url" toml:"admin_url"`
    APIKey           string `json:"api_key" toml:"api_key"`
    Timeout          int    `json:"timeout" toml:"timeout"`
    UpstreamAddr     string `json:"upstream_addr" toml:"upstream_addr"`
    GrpcUpstreamAddr string `json:"grpc_upstream_addr" toml:"grpc_upstream_addr"`
}
```
<!-- @end-code -->

TOML 示例：

<!-- @code: apisix-toml -->
```toml
[apisix]
admin_url          = "http://127.0.0.1:9180"
api_key            = "admin-api-key"
timeout            = 5
upstream_addr      = "host.docker.internal:8080"
grpc_upstream_addr = "host.docker.internal:9103"
```
<!-- @end-code -->

注册到 APISIX：

<!-- @code: apisix-register -->
```go
gw, err := register.NewApisixGateway(conf.Get("apisix"))
if err != nil {
    return err
}
defer gw.Close()

if err := gw.RegisterService(ctx, svc); err != nil {
    return err
}
defer gw.DeregisterService(ctx, svc.ID)
```
<!-- @end-code -->

`RegisterService` 会：

1. 使用 `PUT /apisix/admin/upstreams/{serviceID}` 创建或更新 upstream。
2. 为每个 route 使用 `PUT /apisix/admin/routes/{serviceID}-route-{index}` 创建路由。
3. 如果 route 带有 `Upstream`，使用路由级 upstream；否则绑定顶层 upstream。

`DeregisterService` 会先删除 routes，再删除 upstream。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- `ServiceRegistry.Register` 每次会生成新的实例 ID，格式为 `{name}-{uuid}`。
- `ServiceRegistry.Deregister` 会取消内部 context；同一个 `ServiceRegistry` 注销后不适合继续 watch。
- `EtcdRegister.Deregister` 当前 revoke 的是注册器保存的租约 ID；一个 `EtcdRegister` 多次注册时只保留最后一次租约。
- `EtcdRegister.Close` 会调用内部 registry 的 `Deregister`，如果是通过 `NewEtcdRegister` 创建且未设置内部 registry，则只关闭 client。
- `register.Init` 是 etcd 全局单例入口，受 `sync.Once` 约束，首次初始化后不会因再次调用而重建。
- `ApisixGateway` 只在内存中记录本进程创建的 route IDs；进程重启后调用 `DeregisterService` 不知道旧 route ID，需要按约定 ID 清理或重新注册覆盖。
- `ApisixConfig.UpstreamAddr` 和 `GrpcUpstreamAddr` 当前只是配置字段，`ApisixGateway.RegisterService` 不会自动用它们构造 `GatewayService`。
<!-- @end-section -->

## 相关文档

- [[reference-component-grpc-server-001]]
- [[guide-app-components-module-001]]
