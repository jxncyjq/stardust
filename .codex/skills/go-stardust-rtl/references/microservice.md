# 微服务支撑参考

## breaker

用于保护不稳定下游。

```go
brk := breaker.NewGoogleBreaker()
err := brk.Do(func() error {
    return callDownstream(ctx)
})
```

业务可接受错误用 `DoWithAcceptable` 排除。请求量少于 100 时不会熔断。

## limit

固定窗口：

```go
limiter := limit.NewPeriodLimiter(60, 100, "api:user")
status, _ := limiter.Take(userID)
```

令牌桶：

```go
limiter := limit.NewTokenLimiter(10, 20, "api:order")
if !limiter.Allow() { return ErrTooManyRequests }
```

优先走 Redis Lua。Redis 不可用时退回本地内存，只对当前进程生效。

## load

自适应降载：

```go
promise, err := shedder.Allow()
if errors.Is(err, load.ErrServiceOverloaded) { return err }
defer promise.Fail()

if err := work(); err != nil { return err }
promise.Pass()
```

Linux 读取 `/proc/stat`。非 Linux CPU 使用率返回 0。

## syncx

合并同 key 的并发请求。

```go
shared := syncx.NewSharedCalls()
val, err := shared.Do("user:1001", func() (interface{}, error) {
    return queryUser(ctx, 1001)
})
```

只合并进行中的请求，不缓存完成结果。

## 入口组合顺序

1. `load` 降载
2. `limit` 限流
3. `breaker` 包下游调用
4. `syncx` 合并并发请求

`metric`、`register`、`tracing` 的详细使用请分别看：

- `metric.md`
- `register.md`
- `tracing.md`
