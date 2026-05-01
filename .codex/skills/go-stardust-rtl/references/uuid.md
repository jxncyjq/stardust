# UUID 参考

## Snowflake ID

`uuid.UuidWorker` 生成 `int64` ID：

- 41 位时间戳
- 10 位 worker ID，范围 `0-1023`
- 12 位同毫秒序列，范围 `0-4095`

默认全局 worker 为 `0`。

```go
uuid.InitWorker(1)
id := uuid.GetUuidInt64()
sid := uuid.GenSessionId()
```

HTTP server 会从 `worker_id` 调用 `uuid.InitWorker`。

## API

| API | 用途 |
| --- | --- |
| `NewUuidWorker(workerID)` | 创建独立 worker |
| `GetUuidInt64()` | 全局整数 ID |
| `GetUuidString()` | 全局字符串 ID |
| `GenSessionId()` | session 或服务实例 ID |
| `UuidWorker.Unmarshal(id)` | 调试解析时间、worker、序列 |

`register.ServiceRegistry` 使用 `GenSessionId()` 生成实例 ID：
`{serviceName}-{sessionID}`。

## 随机字符串

| API | 输出 |
| --- | --- |
| `GenString(n)` | 字母数字字符串 |
| `GenNumberString(n)` | 数字字符串 |
| `GenDateRnString(id)` | `yyyyMMdd + id + 2 位随机数字` |
| `GenBytes(n)` | 当前实现有问题，新代码避免使用 |

`GenBytes(n)` 当前返回 `2*n` 长度，前半段是零值字节。

## 注意

- 多实例必须分配不同 `worker_id`。
- 应用启动后不要反复调用 `InitWorker`。
- `UuidWorker.Get()` 并发安全。
- 随机字符串函数忽略了 `crypto/rand.Int` 错误。
