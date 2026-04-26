package register

import "context"

// GatewayUpstream 上游服务配置
type GatewayUpstream struct {
	Type   string         `json:"type"`             // 负载均衡: roundrobin, chash, ewma
	Nodes  map[string]int `json:"nodes"`            // "host:port": weight
	Scheme string         `json:"scheme,omitempty"` // upstream 协议: http/https/grpc/grpcs
}

// GatewayRoute 路由规则
type GatewayRoute struct {
	Name       string           `json:"name"`                  // 路由名称
	URI        string           `json:"uri"`                   // 匹配路径, e.g. "/api/*"
	Methods    []string         `json:"methods,omitempty"`     // HTTP 方法
	UpstreamID string           `json:"upstream_id,omitempty"` // 关联上游ID
	Upstream   *GatewayUpstream `json:"upstream,omitempty"`    // 路由级上游覆盖（用于 grpc 等差异协议）
}

// GatewayService 网关服务注册信息
type GatewayService struct {
	ID       string           `json:"id"`       // 服务唯一ID (用作 APISIX upstream/route ID 前缀)
	Name     string           `json:"name"`     // 服务名称
	Upstream *GatewayUpstream `json:"upstream"` // 上游配置
	Routes   []*GatewayRoute  `json:"routes"`   // 路由规则列表
}

// Gateway 网关注册抽象接口
type Gateway interface {
	// RegisterService 注册服务到网关 (创建 upstream + routes)
	RegisterService(ctx context.Context, svc *GatewayService) error
	// DeregisterService 从网关注销服务 (删除 routes + upstream)
	DeregisterService(ctx context.Context, serviceID string) error
	// Close 关闭连接/清理资源
	Close() error
}
