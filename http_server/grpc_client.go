package httpServer

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

// GrpcClient gRPC 客户端封装，支持直连和 etcd 服务发现
type GrpcClient struct {
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewGrpcClient 创建 gRPC 客户端
func NewGrpcClient(conf GrpcClientConfig, opts ...grpc.DialOption) (*GrpcClient, error) {
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	}
	opts = append(defaultOpts, opts...)

	var target string
	if conf.Etcd != nil {
		// 通过 etcd 服务发现
		resolverBuilder := NewEtcdResolverBuilder(conf.Etcd)
		resolver.Register(resolverBuilder)
		target = fmt.Sprintf("%s:///%s", etcdScheme, conf.Etcd.ServiceName)
	} else {
		// 直连模式
		target = conf.Target
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return &GrpcClient{
		conn:   conn,
		logger: getLoggerSafe("grpc_client"),
	}, nil
}

// Conn 获取底层连接
func (c *GrpcClient) Conn() *grpc.ClientConn {
	return c.conn
}

// Close 关闭连接
func (c *GrpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
