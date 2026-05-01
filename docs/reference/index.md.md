以下是 `reference` 目录下所有 Markdown 文件的索引，按功能模块分类整理，便于快速查阅：

### 1. 核心基础模块 (Core Modules)
| 文件路径                                                                          | 标题           | 核心功能                                      |
| :---------------------------------------------------------------------------- | :----------- | :---------------------------------------- |
| `reference/reference-uuid-module-001.md`[[reference-uuid-module-001\|uuid模块]] | uuid 模块使用参考  | Snowflake ID 生成、随机字符串/字节生成、Worker 管理      |
| `reference/reference-i18n-module-001.md`                                      | i18n 模块使用参考  | 本地化、Bundle 管理、请求上下文语言切换                   |
| `reference/reference-authz-module-001.md`                                     | authz 模块使用参考 | Casbin 授权、RBAC 模型、策略文件配置、GORM 适配器<br><br> |


---

## 🛡️ 2. 生产环境最佳实践

| `reference/reference-microservice-module-001.md` | microService 模块使用参考                         | 熔断 (breaker)、限流 (limit)、降载 (load)、合并请求 (syncx) |
| ------------------------------------------------ | ------------------------------------------- | ---------------------------------------------- |
| `reference/reference-metric-module-001.md`       | metric 模块使用参考                               | Prometheus 指标封装 (Counter/Gauge/Histogram)      |
| `reference/reference-register-module-001.md`     | register 模块使用参考                             | 服务注册/发现 (etcd)、APISIX 网关注册、gRPC 集成             |
| `reference/reference-http-server-api-usage-001.md` | http_server 接口使用参考                     | IHandler 接口注册、Response 统一返回、HTTP 模块目录结构 |

### 3. 数据存储模块 (Data Storage)
| 文件路径 | 标题 | 核心功能 |
| :--- | :--- | :--- |
| `reference/reference-redis-module-001.md` | redis 模块使用参考 | 连接管理、RedisCli 视图、Pub/Sub、Stream、Notification |
| `reference/reference-mongodb-module-001.md` | mongodb 模块使用参考 | 多实例管理、CRUD 封装、ObjectID、错误约定 |
| `reference/reference-databases-module-001.md` | databases 模块使用参考 | GORM 封装、DAO、分页、事务、主从读写分离 |
| `reference/reference-clickhouse-module-001.md` | clickhouse 模块使用参考 | 多实例管理、SQL 查询/执行、异步写入、批量写入 |

### 4. 消息队列模块 (Messaging)
| 文件路径 | 标题 | 核心功能 |
| :--- | :--- | :--- |
| `reference/reference-nats-module-001.md` | nats 模块使用参考 | 连接管理、普通发布订阅、JetStream 工作队列、Stream/Consumer |

### 5. 组件使用说明 (Components)
| 文件路径 | 标题 | 核心功能 |
| :--- | :--- | :--- |
| `reference/components/reference-component-logs-001.md` | LogsComponent 使用说明 | 日志初始化与配置 |
| `reference/components/reference-component-tracing-001.md` | TracingComponent 使用说明 | OpenTelemetry/Jaeger 链路追踪初始化 |
| `reference/components/reference-component-redis-001.md` | RedisComponent 使用说明 | Redis 连接管理器初始化 |
| `reference/components/reference-component-nats-001.md` | NatsComponent 使用说明 | NATS/JetStream 连接管理器初始化 |
| `reference/components/reference-component-mongodb-001.md` | MongoDBComponent 使用说明 | MongoDB 连接管理器初始化 |
| `reference/components/reference-component-databases-001.md` | DatabasesComponent 使用说明 | GORM 数据库管理器初始化 |
| `reference/components/reference-component-clickhouse-001.md` | ClickhouseComponent 使用说明 | ClickHouse 连接管理器初始化 |
| `reference/components/reference-component-i18n-001.md` | I18nComponent 使用说明 | go-i18n 国际化支撑层初始化 |
| `reference/components/reference-component-authz-001.md` | AuthzComponent 使用说明 | Casbin 授权器初始化 |
| `reference/components/reference-component-http-server-001.md` | HTTPServerComponent 使用说明 | Gin HTTP 服务器创建与启动 |
| `reference/components/reference-component-http-server-from-app-001.md` | HTTPServerFromApp 使用说明 | 绑定 Application 的 HTTP 组件写法 |
| `reference/components/reference-component-grpc-server-001.md` | GRPCServerComponent 使用说明 | gRPC 服务器创建与 Protobuf 注册 |
| `reference/components/reference-component-business-001.md` | BusinessComponent 使用说明 | 业务模块聚合、HTTP/gRPC 路由自动分发 |

### 6. 关键依赖关系速查
*   **UUID** 被 `register` (服务实例ID) 和 `http_server` (Worker ID) 使用。
*   **Register** 依赖 `uuid` (生成ID) 和 `redis` (分布式锁/限流兜底)。
*   **Microservice** 模块是 `breaker`, `limit`, `load`, `syncx` 的聚合文档，并引用 `register` 和 `metric`。
*   **Metric** 模块被 `http_server` 中间件和 `microservice` 模块引用。
*   **Authz** 模块依赖 `databases` (GORM适配器) 或 `file` 适配器。
*   **所有 Component** 均依赖 `logs` 组件。
