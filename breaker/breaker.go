package breaker

import "errors"

var ErrServiceUnavailable = errors.New("circuit breaker is open")

// Promise 请求承诺
type Promise interface {
	Accept()
	Reject(err error)
}

// Breaker 熔断器接口
type Breaker interface {
	// Allow 判断是否允许请求通过
	Allow() (Promise, error)
	// Do 执行请求并自动记录结果
	Do(req func() error) error
	// DoWithAcceptable 自定义可接受错误的判断
	DoWithAcceptable(req func() error, acceptable func(err error) bool) error
}
