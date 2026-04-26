package httpServer

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/jxncyjq/stardust/http_server/interceptor"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/register"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GrpcServer 增强的 gRPC 服务器，内置拦截器链、服务注册
type GrpcServer struct {
	conf     GrpcServerConfig
	server   *grpc.Server
	listener net.Listener
	registry *register.ServiceRegistry
	logger   *zap.Logger
}

// NewGrpcServer 创建 gRPC 服务器，自动注入拦截器链（metric → tracing → breaker → timeout）
func NewGrpcServer(conf GrpcServerConfig, opts ...grpc.ServerOption) (*GrpcServer, error) {
	timeout := time.Duration(conf.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	// 内置拦截器链（顺序很重要）
	chainedInterceptors := grpc.ChainUnaryInterceptor(
		interceptor.UnaryMetricInterceptor(),
		interceptor.UnaryTracingInterceptor("rpc"),
		interceptor.UnaryBreakerInterceptor(),
		interceptor.UnaryTimeoutInterceptor(timeout),
	)
	opts = append([]grpc.ServerOption{chainedInterceptors}, opts...)

	return &GrpcServer{
		conf:   conf,
		server: grpc.NewServer(opts...),
		logger: getLoggerSafe("grpc_server"),
	}, nil
}

// Server 获取原生 grpc.Server 用于注册服务
func (s *GrpcServer) Server() *grpc.Server {
	return s.server
}

// Start 实现 service.Service 接口（可直接传入 ServiceGroup.Add）
func (s *GrpcServer) Start() {
	if err := s.Startup(); err != nil {
		if s.logger != nil {
			s.logger.Fatal("grpc server start failed", zap.Error(err))
		}
	}
}

// Startup 启动服务器，实现 service.Startable 接口
func (s *GrpcServer) Startup() error {
	listener, err := net.Listen("tcp", s.conf.ListenOn)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.logger != nil {
		s.logger.Info("grpc server listening", zap.String("addr", listener.Addr().String()))
	}

	// 服务注册（etcd）
	if s.conf.Etcd != nil {
		byteConfig, err := json.Marshal(s.conf.Etcd)
		if err != nil {
			s.logger.Warn("failed to marshal etcd config", zap.Error(err))
			return err
		}
		reg, err := register.NewEtcdRegister(byteConfig)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("etcd register failed", zap.Error(err))
			}
		} else {
			s.registry = register.NewServiceRegistry(reg)
			host, port := parseAddr(listener.Addr().String())
			s.registry.Register(s.conf.Etcd.ServiceName, host, port, s.conf.Etcd.Tags, nil)
		}
	}

	go func() {
		if err := s.server.Serve(listener); err != nil {
			if s.logger != nil {
				s.logger.Error("grpc server error", zap.Error(err))
			}
		}
	}()
	return nil
}

// Stop 优雅关闭
func (s *GrpcServer) Stop() {
	if s.registry != nil {
		s.registry.Close()
	}
	s.server.GracefulStop()
	if s.logger != nil {
		s.logger.Info("grpc server stopped")
	}
}

// Address 获取实际监听地址
func (s *GrpcServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.conf.ListenOn
}

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

// getLoggerSafe 获取 logger，未初始化时返回 nop logger。
func getLoggerSafe(module string) *zap.Logger {
	return logs.GetLogger(module)
}
