---
id: "reference-uuid-module-001"
title: "uuid 模块使用参考"
aliases: ["uuid模块", "Snowflake ID", "GenSessionId", "UuidWorker"]
type: "reference"
category: "backend/library"
tags: ["uuid", "snowflake", "session-id", "random-string", "worker-id"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: null
children: []
related_docs:
  - id: "reference-component-http-server-001"
    relation: "related_to"
    path: "./components/reference-component-http-server-001.md"
  - id: "reference-register-module-001"
    relation: "related_to"
    path: "./reference-register-module-001.md"
  - id: "reference-component-http-server-from-app-001"
    relation: "related_to"
    path: "./components/reference-component-http-server-from-app-001.md"
---

# uuid 模块使用参考

<!-- @section: overview -->
## 概述

`uuid` 模块提供两类 ID/随机值能力：

| 能力 | API | 说明 |
| --- | --- | --- |
| Snowflake 风格整数 ID | `UuidWorker`、`GetUuidInt64`、`GetUuidString`、`GenSessionId` | 用于 session、服务实例 ID 等全局唯一 ID |
| 随机字符串/字节 | `GenString`、`GenNumberString`、`GenDateRnString`、`GenBytes` | 用于短随机码、数字串、日期前缀编码等 |

HTTP server 会根据 `worker_id` 调用 `uuid.InitWorker`，WebSocket session ID 和服务注册实例 ID 当前都使用 `uuid.GenSessionId()`。
<!-- @end-section -->

<!-- @section: snowflake -->
## Snowflake ID

核心类型是 `UuidWorker`：

<!-- @code: worker-struct -->
```go
type UuidWorker struct {
    // 内部包含 timestamp、workerID、number 和互斥锁。
}
```
<!-- @end-code -->

ID 位布局：

| 部分 | 位数 | 说明 |
| --- | --- | --- |
| 时间戳 | 41 bits | 当前毫秒时间减去自定义 epoch |
| workerID | 10 bits | 节点 ID，范围 `0-1023` |
| number | 12 bits | 同毫秒内自增序列，最大 `4095` |

自定义 epoch 是 `1652751577000`，即 `2022-05-17 09:39:37`。
<!-- @end-section -->

<!-- @section: worker -->
## Worker 初始化

创建独立 worker：

<!-- @code: new-worker -->
```go
worker, err := uuid.NewUuidWorker(1)
if err != nil {
    return err
}

id := worker.Get()
```
<!-- @end-code -->

全局 worker 默认在包初始化时使用 `workerID=0`。应用可通过 `InitWorker` 设置进程级 worker：

<!-- @code: init-worker -->
```go
if err := uuid.InitWorker(1); err != nil {
    return err
}

id := uuid.GetUuidInt64()
sid := uuid.GenSessionId()
```
<!-- @end-code -->

`workerID` 必须在 `0-1023` 范围内，超出范围会返回 `worker ID excess of quantity`。

HTTP server 初始化时会读取 `HttpServerConfig.WorkerID`：

<!-- @code: http-worker-id -->
```toml
[http_server]
worker_id = 1
```
<!-- @end-code -->
<!-- @end-section -->

<!-- @section: global-api -->
## 全局 ID API

| 方法 | 返回值 | 说明 |
| --- | --- | --- |
| `GetUuidInt64()` | `int64` | 从全局 worker 生成整数 ID |
| `GetUuidString()` | `string` | 从全局 worker 生成十进制字符串 ID |
| `GenSessionId()` | `string` | 从全局 worker 生成 session ID 字符串 |

示例：

<!-- @code: global-api -->
```go
sessionID := uuid.GenSessionId()
orderID := uuid.GetUuidString()
numericID := uuid.GetUuidInt64()
```
<!-- @end-code -->

`register.ServiceRegistry` 会使用 `GenSessionId()` 生成服务实例 ID，格式为 `{serviceName}-{sessionID}`。
<!-- @end-section -->

<!-- @section: unmarshal -->
## 解析 ID

`UuidWorker.Unmarshal(id)` 可拆解 ID：

<!-- @code: unmarshal -->
```go
worker, _ := uuid.NewUuidWorker(1)
id := worker.Get()

parts := worker.Unmarshal(id)
_ = parts["time"]
_ = parts["worker"]
_ = parts["number"]
```
<!-- @end-code -->

返回字段：

| key | 说明 |
| --- | --- |
| `id` | 原始 ID |
| `time` | 秒级时间戳 |
| `worker` | worker ID |
| `number` | 同毫秒内序列号 |
<!-- @end-section -->

<!-- @section: random-string -->
## 随机字符串

生成字母数字混合字符串：

<!-- @code: gen-string -->
```go
token := uuid.GenString(16)
```
<!-- @end-code -->

生成数字字符串：

<!-- @code: gen-number-string -->
```go
code := uuid.GenNumberString(6)
```
<!-- @end-code -->

生成日期前缀字符串：

<!-- @code: gen-date-rn-string -->
```go
value := uuid.GenDateRnString("01")
// 格式: yyyyMMdd + idString + 2位随机数字
```
<!-- @end-code -->

这些方法使用 `crypto/rand` 生成随机下标，字符集分别为：

| 方法 | 字符集 |
| --- | --- |
| `GenString` | `a-zA-Z1-9` 和 `0` |
| `GenNumberString` | `0-9` |
| `GenDateRnString` | 日期 + 调用方传入字符串 + 两位数字 |
<!-- @end-section -->

<!-- @section: gen-bytes -->
## 随机字节

`GenBytes(len)` 设计意图是生成字母数字字节序列：

<!-- @code: gen-bytes -->
```go
bs := uuid.GenBytes(16)
```
<!-- @end-code -->

当前实现需要注意：函数先创建了长度为 `len` 的零值切片，然后 append `len` 个随机字符，因此返回切片长度为 `2*len`，前半段为零值字节。现阶段不建议在新业务中直接使用 `GenBytes` 作为 token 或密钥材料；需要字节随机值时应先修正实现或自行使用 `crypto/rand`。
<!-- @end-section -->

<!-- @section: usage -->
## 使用场景

| 场景 | 推荐 API |
| --- | --- |
| WebSocket session ID | `GenSessionId()` |
| 服务注册实例 ID | `GenSessionId()` |
| 数据库 bigint 主键 | `GetUuidInt64()` 或独立 `UuidWorker.Get()` |
| 外部展示的纯数字 ID | `GetUuidString()` |
| 短验证码 | `GenNumberString(len)` |
| 业务流水号日期前缀 | `GenDateRnString(idString)` |

多实例部署时必须为每个进程分配不同 `worker_id`，否则同一毫秒内不同实例可能生成重复 ID。
<!-- @end-section -->

<!-- @section: cautions -->
## 注意事项

- 全局 `uuidWorker` 是包级变量，`InitWorker` 会替换全局 worker；应用启动后不要在运行期反复调用。
- `NewUuidWorker` 只校验 workerID 范围，不处理跨进程 workerID 分配。
- `UuidWorker.Get()` 内部加锁，单 worker 并发安全。
- 同一毫秒内序列超过 `4095` 时会自旋等待下一毫秒。
- `Unmarshal` 返回的是 map，适合调试，不建议作为高频路径解析方式。
- `GenString`、`GenNumberString` 当前忽略了 `crypto/rand.Int` 的错误。
- `GenBytes` 当前返回长度和内容存在实现问题，新代码应避免使用。
<!-- @end-section -->

## 相关文档

- [[reference-component-http-server-001]]
- [[reference-register-module-001]]
