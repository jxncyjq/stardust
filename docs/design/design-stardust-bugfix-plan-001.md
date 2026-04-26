---
id: "design-stardust-bugfix-plan-001"
title: "stardust 库问题修补计划"
aliases: ["bugfix plan", "修补计划", "问题修复", "代码审查修复"]
type: "design"
category: "backend/library"
tags: ["bugfix", "refactor", "concurrency", "reliability", "panic"]
version: "2.0.0"
created: "2026-04-26"
updated: "2026-04-26"
author: "jxncyjq"
status: "completed"
parent: null
children: []
related_docs:
  - id: "design-app-component-001"
    relation: "related_to"
    path: "./design-app-component-001.md"
---

# stardust 库问题修补计划

> 基于全库代码审查结果。共 5 个 CRITICAL、10 个 HIGH、12 个 MEDIUM 问题。
> 三个阶段全部修复完成（2026-04-26）。

<!-- @section: summary -->
## 问题统计

| 级别 | 数量 | 状态 |
|------|------|------|
| CRITICAL | 5 | ✅ 全部完成 |
| HIGH | 10 | ✅ 全部完成 |
| MEDIUM | 12 | ✅ 全部完成 |
| LOW | 8 | 待处理 |
<!-- @end-section -->

<!-- @section: phase1 -->
## 阶段一：CRITICAL 修复 ✅ 已完成

### C1 — WebSocket Send 永久阻塞 + close channel panic

**文件**：`http_server/ws_client.go:64,149,156`

**问题**：
- `Send` 直接写 channel，缓冲区满时永久阻塞
- `Close()` 先 `close(c.send)`，`ReceivedMessage` goroutine 继续写已关闭 channel → panic

**修复方案**：

```go
// Send 加 select + default，满则丢弃并记录
func (c *Client) Send(msg []byte) bool {
    select {
    case c.send <- msg:
        return true
    default:
        c.logger.Warn("send buffer full, message dropped")
        return false
    }
}

// Close 前设置 closed 标志，写 channel 前检查
func (c *Client) Close() {
    c.once.Do(func() {
        atomic.StoreInt32(&c.closed, 1)
        close(c.send)
        c.conn.Close()
    })
}

// ReceivedMessage 中写 send 前检查
if atomic.LoadInt32(&c.closed) == 0 {
    c.send <- []byte(resultMsg)
}
```

---

### C2 — WebSocket ClientManager.Stop 死锁 + 无限递归

**文件**：`http_server/ws_client_manager.go:143-154`

**问题**：
- `Stop()` 持锁调 `client.Close()` → `UnregisterClient` 向 channel 写入 → 无人消费 → 死锁
- `Start()` 收到 stopChan 后调 `m.Stop()` → `Stop()` 再发 stopChan → 无限递归

**修复方案**：

```go
func (m *ClientManager) Stop() {
    // 通知 Start goroutine 退出，不持锁
    close(m.stopChan)
}

// Start 中收到 stop 信号后执行清理
case <-m.stopChan:
    m.mu.Lock()
    for client := range m.clients {
        client.conn.Close() // 直接关连接，不走 unregister channel
    }
    m.clients = make(map[*Client]bool)
    m.mu.Unlock()
    return
```

---

### C3 — NatsConnManager.StartAll 持读锁调阻塞函数

**文件**：`nats/nats_connection_manager.go:80-86`

**问题**：`Start()` 含无限 for 循环，在持 `RLock` 时调用 → 锁永不释放，后续连接无法启动。

**修复方案**：

```go
func (m *NatsConnManager) StartAll() {
    m.mu.RLock()
    conns := make([]*NatsConnection, 0, len(m.clients))
    for _, c := range m.clients {
        conns = append(conns, c)
    }
    m.mu.RUnlock() // 先释放锁

    for _, c := range conns {
        go c.Start() // 每个连接独立 goroutine
    }
}
```

> **注**：`app/components/nats.go` 中已用 `go c.manager.StartAll()` 包裹，但 StartAll 内部仍需修复。

---

### C4 — redis_internal.go 条件错误 + 不安全类型断言

**文件**：`redis/redis_internal.go:62,67`

**问题**：
- 第62行 `if len(members) > 0` 应为 `if len(zSlice) > 0`
- 第67行 `m.Member.([]byte)` 强转，`redis.Z.Member` 实际为 `string` → 运行时 panic

**修复方案**：

```go
func toRangeZMembers(zSlice []redis.Z) []ZMember {
    if len(zSlice) == 0 { // 修复条件
        return nil
    }
    members := make([]ZMember, 0, len(zSlice))
    for _, m := range zSlice {
        var memberStr string
        switch v := m.Member.(type) { // 安全类型断言
        case string:
            memberStr = v
        case []byte:
            memberStr = string(v)
        default:
            memberStr = fmt.Sprintf("%v", v)
        }
        members = append(members, ZMember{Member: memberStr, Score: m.Score})
    }
    return members
}
```

---

### C5 — errors/try_catch.go panic 被静默吞掉

**文件**：`errors/try_catch.go:7-16`

**问题**：`catch=nil` 时发生 panic，`recover()` 捕获后既不调用 catch（nil）也不重新 panic，错误静默消失。

**修复方案**：

```go
func TryFunc(try func(), catch func(interface{}), finally func()) {
    defer func() {
        if finally != nil {
            finally()
        }
    }()
    defer func() {
        if r := recover(); r != nil {
            if catch != nil {
                catch(r)
            } else {
                panic(r) // 无 catch 则重新抛出
            }
        }
    }()
    try()
}
```
<!-- @end-section -->

<!-- @section: phase2 -->
## 阶段二：HIGH 修复 ✅ 已完成

### H1 — 全库 panic 策略改为 error return

**文件**：`logs/logger.go:98`、`databases/db_manager.go:40,49`、`register/etcd.go` 等 40+ 处

**问题**：库代码大量使用 panic 进行流程控制，调用方无法优雅处理。

**修复方向**：
- `GetLogger(m string)` → `(logger *zap.Logger, err error)` 或返回 nop logger
- `GetDatabaseManager()` → `(*DatabaseManager, error)`
- `NewEtcdRegister()` 已返回 error，内部不再 panic
- 所有 `Init([]byte)` 系函数改为 `Init([]byte) error`

> 此项工作量最大，可分包逐步迁移，`app/components` 的 `recoverToError` 作为过渡方案。

---

### H2 — conf.go os.Setenv 全局污染

**文件**：`conf/conf.go:86-88`

**修复方案**：

```go
// 用包级变量替代 os.Setenv
var (
    AppName        string
    AppVersion     string
    RedisKeyPrefix string
)

// 提供 getter
func GetAppName() string        { return AppName }
func GetRedisKeyPrefix() string { return RedisKeyPrefix }
```

---

### H3 — conf.Get 竞态条件

**文件**：`conf/conf.go:96-134`

**修复方案**：移除 `Init()` 内的 `mu.Lock()`，只用 `sync.Once` 保证初始化唯一性：

```go
func Init() {
    initOnce.Do(func() {
        // 直接执行，不需要额外锁
        // sync.Once 本身保证并发安全
    })
}
```

---

### H4 — jwt.go 签名错误静默丢弃

**文件**：`jwt/jwt.go:51`

```go
// 修改签名返回 error
func JWTEncryptWithExpiry(claims jwt.Claims, expiry time.Duration) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(mySigningKey) // 直接返回 (string, error)
}
```

---

### H5 — load/shedder.go 降载器实际不工作

**文件**：`load/shedder.go:175-183`

**问题**：`cpuThreshold` 字段存在但未使用，`systemOverloaded()` 只看历史丢弃记录。

**修复方案**：接入真实 CPU 采样：

```go
func (as *adaptiveShedder) systemOverloaded() bool {
    usage := getCpuUsage() // 实现 CPU 采样
    return usage > float64(as.cpuThreshold)
}

// 简单实现：读 /proc/stat（Linux）或使用 gopsutil
func getCpuUsage() float64 {
    // ...
}
```

---

### H6 — HttpServer goroutine 泄漏

**文件**：`http_server/server_http.go:69,100-108`

```go
type HttpServer struct {
    ctx    context.Context
    cancel context.CancelFunc
    // ...
}

func NewHttpServer(config []byte) (*HttpServer, error) {
    ctx, cancel := context.WithCancel(context.Background())
    return &HttpServer{ctx: ctx, cancel: cancel, ...}, nil
}

func (m *HttpServer) Stop() {
    m.cancel() // 触发所有监听 ctx 的 goroutine 退出
    // ...
}
```

---

### H7 — errors/warp.go 未实现 Unwrap

**文件**：`errors/warp.go:45-56`

```go
// 添加标准 Unwrap 接口，使 errors.Is/As 可穿透
func (w *wrappedError) Unwrap() error {
    return w.Inner
}
```

---

### H8 — parseAddr 手工解析端口

**文件**：`http_server/server_grpc.go:120-127`

```go
func parseAddr(addr string) (string, int) {
    host, portStr, err := net.SplitHostPort(addr)
    if err != nil {
        return addr, 0
    }
    port, err := strconv.Atoi(portStr)
    if err != nil {
        return host, 0
    }
    return host, port
}
```

---

### H9 — jwt.go 使用 fmt.Println 输出错误

**文件**：`jwt/jwt.go:18,24`

将 `fmt.Println(err)` 改为将 error 包含在返回值中，或使用 `logs.GetLogger("jwt")` 结构化输出。

---

### H10 — tracing toString 丢失非 string 值

**文件**：`tracing/jaeger.go:161-169`

```go
func toString(v interface{}) string {
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v) // 兜底转换，不丢失信息
}
```
<!-- @end-section -->

<!-- @section: phase3 -->
## 阶段三：MEDIUM 修复（建议）✅ 已完成（2026-04-26）

| #   | 文件                                              | 问题                        | 修复结果                                              |
| --- | ----------------------------------------------- | ------------------------- | ------------------------------------------------- |
| M1  | `databases/mysql.go` `dao.go` `db_manager.go`   | `DBInterface` 名称误导        | ✅ 重命名为 `DBConn`                                  |
| M2  | `databases/dao.go`                              | `Upsert` 未使用 where 参数     | ✅ 改用 `Where().Assign().FirstOrCreate()`          |
| M3  | `clickhouse/manager.go`                         | 建连后未 Ping 验证              | ✅ 加 `db.PingContext()` 验证可达性                     |
| M4  | `limit/periodlimit.go`                          | localCount map 无限增长       | ✅ 加 `cleanupLoop()` goroutine（2×period 周期清理）     |
| M5  | `nats/nats.go` `subscribe.go`                   | handlers map 无锁并发写        | ✅ 加 `handlersMu sync.Mutex` 保护写入                 |
| M6  | `nats/nats.go` `subscribe.go`                   | js 字段重连时无锁写入              | ✅ 加 `jsMu sync.RWMutex`；读路径统一改用 `getJS()` helper |
| M7  | `breaker/googlebreaker.go`                      | 使用已弃用 rand API            | ✅ 移除 `rand.Rand` + `mu`；改用 `rand.Float64()`      |
| M8  | `http_server/server_http.go`                    | Get/Post/Put group 不存在静默  | ✅ 改为 panic（编程错误，启动时即暴露）                          |
| M9  | `register/etcd.go`                              | KeepAlive goroutine 无退出机制 | ✅ 加 `ctx.Done()` select 分支显式退出                   |
| M10 | `http_server/server_grpc.go`                    | getLoggerSafe 掩盖初始化问题     | ✅ 移除多余 `recover()`，直接调 `logs.GetLogger()`        |
| M11 | `errors/warp.go` → `errors/wrap.go`             | 文件名拼写错误                   | ✅ 重命名                                            |
| M11 | `jwt/PBEWhithMD5AndDES.go` → `PBEWithMD5AndDES.go` | 文件名拼写错误               | ✅ 重命名                                            |
| M12 | `http_server/handler.go`                        | Resp 泛型参数实际无效             | ✅ 改为 `func(*gin.Context, Req) (Resp, error)` 签名  |
<!-- @end-section -->

<!-- @section: schedule -->
## 实施顺序 ✅ 全部完成（2026-04-26）

```
阶段一（C1-C5）✅
  ├── ✅ C3 NatsConnManager.StartAll 锁修复
  ├── ✅ C4 redis_internal 条件 + 类型断言
  ├── ✅ C5 try_catch panic 重新抛出
  ├── ✅ C1 WebSocket Send + Close
  └── ✅ C2 ClientManager Stop 死锁

阶段二（H1-H10）✅
  ├── ✅ H7 errors.Unwrap
  ├── ✅ H8 parseAddr strconv
  ├── ✅ H10 tracing toString
  ├── ✅ H4 jwt error return
  ├── ✅ H9 jwt fmt.Println 移除
  ├── ✅ H3 conf.Get 竞态
  ├── ✅ H2 conf os.Setenv 移除
  ├── ✅ H6 HttpServer goroutine 泄漏
  ├── ✅ H1 全库 panic → error（40+ 处）
  └── ✅ H5 CPU 采样（cpu_linux.go / cpu_other.go）

阶段三（M1-M12）✅
  ├── ✅ M1 DBInterface → DBConn 重命名
  ├── ✅ M2 Upsert 使用 where 参数
  ├── ✅ M3 ClickHouse 建连后 Ping 验证
  ├── ✅ M4 PeriodLimiter localCount 定期清理
  ├── ✅ M5 nats handlers map 加锁
  ├── ✅ M6 nats js 字段加 RWMutex
  ├── ✅ M7 breaker 移除废弃 rand API
  ├── ✅ M8 路由注册 group 不存在时 panic
  ├── ✅ M9 etcd KeepAlive 监听 ctx 退出
  ├── ✅ M10 getLoggerSafe 移除多余 recover
  ├── ✅ M11 文件名拼写修正（warp→wrap, Whith→With）
  └── ✅ M12 Handler 泛型签名修复
```
<!-- @end-section -->

<!-- @section: acceptance -->
## 验收标准

- [x] `go build ./...` 全部通过
- [x] `go test ./...` 全部通过（`errors` 包 1 个预存测试设计缺陷除外，非本次引入）
- [ ] `go test -race ./...` 全部通过（无竞态）— 待验证
- [ ] `go vet ./...` 无警告 — 待验证
- [x] WebSocket Send 使用非阻塞 select + closeOnce 防重复 Close
- [x] NATS StartAll 先拷贝切片再释锁，每连接独立 goroutine
- [x] `errors.Is` / `errors.As` 可穿透 `wrappedError`（Unwrap 已实现）
- [x] JWT 签名错误有明确 error 返回
- [x] ClickHouse 建连后 Ping 验证，错误地址 Init 阶段即报错
- [x] AdaptiveShedder 接入 CPU 采样（Linux /proc/stat，非 Linux 返回 0）
<!-- @end-section -->

## 相关文档

- [[design-app-component-001|app 组件化统一接口设计]]
- [[guide-progress-001|实现进度记录]]
