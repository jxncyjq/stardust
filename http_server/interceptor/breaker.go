package interceptor

import (
	"context"
	"sync"

	"github.com/jxncyjq/stardust/breaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	breakerMap = make(map[string]breaker.Breaker)
	breakerMu  sync.RWMutex
)

func getBreaker(method string) breaker.Breaker {
	breakerMu.RLock()
	b, ok := breakerMap[method]
	breakerMu.RUnlock()
	if ok {
		return b
	}

	breakerMu.Lock()
	defer breakerMu.Unlock()
	b, ok = breakerMap[method]
	if ok {
		return b
	}
	b = breaker.NewGoogleBreaker()
	breakerMap[method] = b
	return b
}

// UnaryBreakerInterceptor 熔断器拦截器
func UnaryBreakerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		b := getBreaker(info.FullMethod)
		promise, err := b.Allow()
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}

		resp, err := handler(ctx, req)
		if err != nil {
			promise.Reject(err)
		} else {
			promise.Accept()
		}
		return resp, err
	}
}
