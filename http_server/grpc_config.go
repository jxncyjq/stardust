package httpServer

import (
	"github.com/jxncyjq/stardust/register"
)

// GrpcServerConfig gRPC 服务器配置
type GrpcServerConfig struct {
	ListenOn string               `json:"listen_on" toml:"listen_on"` // 监听地址，如 "0.0.0.0:9090"
	Timeout  int64                `json:"timeout" toml:"timeout"`     // 超时(ms)
	Etcd     *register.EtcdConfig `json:"etcd,omitempty" toml:"etcd,omitempty"`
}

// GrpcClientConfig gRPC 客户端配置
type GrpcClientConfig struct {
	Target  string               `json:"target" toml:"target"`   // 直连地址（与 Etcd 二选一）
	Timeout int64                `json:"timeout" toml:"timeout"` // 超时(ms)
	Etcd    *register.EtcdConfig `json:"etcd,omitempty" toml:"etcd,omitempty"`
}
