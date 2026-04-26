# stardust.mini

Go 微服务框架库，提供构建微服务应用所需的核心功能。

## 功能模块

| 模块 | 说明 |
|------|------|
| `breaker` | 熔断器（Google SRE 自适应算法，滑动窗口） |
| `clickhouse` | ClickHouse 连接管理（单例，多实例） |
| `codec` | 序列化抽象（JSON / Protobuf / FlatBuffer） |
| `conf` | TOML/YAML 配置文件管理（基于 Viper） |
| `databases` | 基于 GORM 的数据库 ORM 层，支持 MySQL/PostgreSQL，读写分离 |
| `errors` | 错误处理框架，支持堆栈跟踪和 Try-Catch 链式 API |
| `http_server` | 基于 Gin 的 HTTP 服务器 + gRPC 双协议支持 |
| `jwt` | JWT 认证（HMAC-SHA256） |
| `limit` | 限流（Redis 令牌桶 + 周期限流） |
| `load` | 自适应负载卸载（Little's Law） |
| `logs` | 基于 Zap 的日志系统，支持日志轮转（lumberjack） |
| `metric` | Prometheus 指标封装（计数器、直方图） |
| `mongodb` | MongoDB 连接管理（单例，多实例） |
| `nats` | NATS/JetStream 消息队列客户端 |
| `redis` | Redis 客户端，支持全命令和多实例管理 |
| `register` | 服务注册（APISIX 网关，基于 etcd） |
| `service` | 服务生命周期管理（ServiceGroup） |
| `syncx` | 并发工具（SharedCalls 请求合并去重） |
| `tracing` | OpenTelemetry 链路追踪（Jaeger 导出） |
| `utils` | 杂项工具（信号处理、HTTP 辅助等） |
| `uuid` | 分布式 ID 生成（随机字符串 + Snowflake） |

## 安装

```bash
go get github.com/jxncyjq/stardust
```

## 典型初始化顺序

```go
conf.Init(configBytes)
logs.Init(configBytes)
databases.Init(configBytes)
redis.Init(configBytes)
nats.Init(configBytes)
// 启动 HTTP/gRPC 服务
```

## 主要依赖

- `gorm.io/gorm` — ORM 框架
- `github.com/gin-gonic/gin` — HTTP 框架
- `github.com/redis/go-redis/v9` — Redis 客户端
- `github.com/nats-io/nats.go` — NATS 客户端
- `go.uber.org/zap` — 日志库
- `go.opentelemetry.io/otel` — 链路追踪
- `github.com/prometheus/client_golang` — 指标采集

## 要求

- Go 1.21+
