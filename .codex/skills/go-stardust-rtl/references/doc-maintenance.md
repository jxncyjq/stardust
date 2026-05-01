# 参考文档维护

## 文件位置

- 模块文档：`docs/reference/`
- 组件文档：`docs/reference/components/`
- reference 目录索引：`docs/reference/index.md`
- 总索引：`docs/docs-index.md`
- 索引说明：`docs/reference-docs-index.md`

## Front Matter

每篇 reference 文档应包含：

```yaml
id: "reference-..."
title: "..."
type: "reference"
category: "backend/library"
tags: [...]
created: "YYYY-MM-DD"
updated: "YYYY-MM-DD"
status: "published"
parent: ...
children: [...]
related_docs:
  - id: "..."
    relation: "related_to|depends_on|parent_of"
    path: "..."
```

## 关系规则

- `parent` 和 `children` 必须双向一致。
- `related_docs.path` 必须是存在的相对路径。
- 新文档必须登记到 `docs/docs-index.md`。
- 新文档也要加入关键词索引。
- 正式 reference 文档更新后，按下表同步对应 skill 参考文件。

## docs/reference ↔ skill 同步映射

| 正式文档（docs/reference/） | 需同步的 skill 参考 |
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
| `components/reference-component-grpc-server-001.md` | `grpc-server.md`, `components.md` |
| `components/reference-component-business-001.md` | `business.md`, `components.md` |
| `components/reference-component-*.md` | `components.md` + 对应模块文件 |

## 当前父子关系

- `reference-docs-index` -> `reference-docs-index-table`、`reference-index-001`
- `design-app-component-001` -> `guide-app-components-module-001`
- `guide-app-components-module-001` -> 各组件文档
- 组件文档 -> 对应模块文档
- `reference-microservice-module-001` -> `reference-metric-module-001`、`reference-register-module-001`

## 校验重点

- 所有 ID 都能在 `docs/docs-index.md` 找到。
- 所有 parent ID 都存在。
- 父子关系互相指回。
- related path 文件存在。
- `docs/reference/` 的新增或重命名文档能映射到对应 skill 摘要。
