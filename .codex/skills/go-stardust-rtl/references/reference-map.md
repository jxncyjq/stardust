# Stardust 参考地图

路由决策在 `SKILL.md` 完成；本文件用于"大图浏览"和 docs ↔ skill 维护对照。

## 组件依赖树

```
app.New → Use(components...)
  LogsComponent          (logs)
  RedisComponent         (logs → redis)
  DatabasesComponent     (logs → databases)
  MongoDBComponent       (logs → mongodb)
  ClickhouseComponent    (logs → clickhouse)
  NatsComponent          (logs → nats)
  TracingComponent       (logs → tracing)
  AuthzComponent         (logs → authz)
  I18nComponent          (logs → i18n)
  HTTPServerFromApp      (logs, tracing → http_server)
  GRPCServerComponent    (logs, tracing → grpc_server)
  Business               (依赖由 WithDependencies 设置)
```

## docs/reference ↔ skill 同步映射

正式文档更新后，按下表同步对应 skill 参考文件。

| 正式文档 | skill 参考 |
| --- | --- |
| `reference-databases-module-001.md` | `databases.md`, `data-infra.md` |
| `reference-redis-module-001.md` | `redis.md`, `data-infra.md` |
| `reference-mongodb-module-001.md` | `mongodb.md`, `data-infra.md` |
| `reference-clickhouse-module-001.md` | `clickhouse.md`, `data-infra.md` |
| `reference-nats-module-001.md` | `nats.md`, `data-infra.md` |
| `reference-metric-module-001.md` | `metric.md`, `microservice.md` |
| `reference-register-module-001.md` | `register.md`, `microservice.md` |
| `reference-microservice-module-001.md` | `microservice.md` |
| `reference-http-server-api-usage-001.md` | `http-server.md` |
| `reference-authz-module-001.md` | `authz.md`, `components.md` |
| `reference-i18n-module-001.md` | `i18n.md`, `components.md` |
| `reference-uuid-module-001.md` | `uuid.md` |
| `reference-component-grpc-server-001.md` | `grpc-server.md`, `components.md` |
| `reference-component-business-001.md` | `business.md`, `components.md` |
| `components/reference-component-*.md` | `components.md` + 对应模块文件 |
| `index.md` | `reference-map.md` |

## 关系规则（新增 / 修改文档时检查）

- `parent` 和 `children` 必须双向一致
- `related_docs.path` 必须存在
- 新增文档必须登记到 `docs/docs-index.md`
- 参考详细维护规则：`doc-maintenance.md`
