package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryMetricInterceptor 指标采集拦截器（记录耗时和状态码）
func UnaryMetricInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := "OK"
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code().String()
			}
		}

		logger := getLoggerSafe("grpc_metric")
		if logger != nil {
			logger.Info("grpc request",
				zap.String("method", info.FullMethod),
				zap.String("code", code),
				zap.Duration("duration", duration),
			)
		}
		return resp, err
	}
}
