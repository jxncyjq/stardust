package httpServer

import (
	"github.com/jxncyjq/stardust/utils"
)

// Server 服务器接口（HTTP 和 gRPC 都实现此接口）
type Server interface {
	Startup() error
	Stop()
}

// NewServer 根据配置创建 HTTP 服务器（保持向后兼容）
func NewServer(configByte []byte) (Server, error) {
	return NewHttpServer(configByte)
}

// MultiServerConfig 多协议服务器配置
type MultiServerConfig struct {
	Http *HttpServerConfig  `json:"http,omitempty" toml:"http,omitempty"`
	Grpc *GrpcServerConfig  `json:"grpc,omitempty" toml:"grpc,omitempty"`
}

// MultiServer 多协议服务器管理器，统一管理 HTTP + gRPC 生命周期
type MultiServer struct {
	httpServer *HttpServer
	grpcServer *GrpcServer
}

// NewMultiServer 根据配置创建多协议服务器
func NewMultiServer(httpConfigBytes []byte, grpcConf *GrpcServerConfig) (*MultiServer, error) {
	ms := &MultiServer{}

	if httpConfigBytes != nil {
		httpSrv, err := NewHttpServer(httpConfigBytes)
		if err != nil {
			return nil, err
		}
		ms.httpServer = httpSrv
	}

	if grpcConf != nil {
		grpcSrv, err := NewGrpcServer(*grpcConf)
		if err != nil {
			return nil, err
		}
		ms.grpcServer = grpcSrv
	}

	return ms, nil
}

// NewMultiServerFromBytes 从字节配置创建多协议服务器
func NewMultiServerFromBytes(httpConfigBytes, grpcConfigBytes []byte) (*MultiServer, error) {
	ms := &MultiServer{}

	if httpConfigBytes != nil {
		httpSrv, err := NewHttpServer(httpConfigBytes)
		if err != nil {
			return nil, err
		}
		ms.httpServer = httpSrv
	}

	if grpcConfigBytes != nil {
		grpcConf, err := utils.Bytes2Struct[GrpcServerConfig](grpcConfigBytes)
		if err != nil {
			return nil, err
		}
		grpcSrv, err := NewGrpcServer(grpcConf)
		if err != nil {
			return nil, err
		}
		ms.grpcServer = grpcSrv
	}

	return ms, nil
}

// HttpServer 获取 HTTP 服务器实例
func (ms *MultiServer) HttpServer() *HttpServer {
	return ms.httpServer
}

// GrpcServer 获取 gRPC 服务器实例
func (ms *MultiServer) GrpcServer() *GrpcServer {
	return ms.grpcServer
}

// Startup 启动所有协议服务器
func (ms *MultiServer) Startup() error {
	if ms.httpServer != nil {
		if err := ms.httpServer.Startup(); err != nil {
			return err
		}
	}
	if ms.grpcServer != nil {
		if err := ms.grpcServer.Startup(); err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止所有协议服务器
func (ms *MultiServer) Stop() {
	if ms.grpcServer != nil {
		ms.grpcServer.Stop()
	}
	if ms.httpServer != nil {
		ms.httpServer.Stop()
	}
}
