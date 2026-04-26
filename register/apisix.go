package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jxncyjq/stardust/utils"
)

// ApisixConfig APISIX 网关配置
type ApisixConfig struct {
	AdminURL         string `json:"admin_url" toml:"admin_url"`                   // Admin API 地址, e.g. "http://localhost:9180"
	APIKey           string `json:"api_key" toml:"api_key"`                       // Admin API Key
	Timeout          int    `json:"timeout" toml:"timeout"`                       // 请求超时(秒), 默认 5
	UpstreamAddr     string `json:"upstream_addr" toml:"upstream_addr"`           // HTTP 上游地址, e.g. "host.docker.internal:8080"
	GrpcUpstreamAddr string `json:"grpc_upstream_addr" toml:"grpc_upstream_addr"` // gRPC 上游地址, e.g. "host.docker.internal:9103"
}

// ApisixGateway APISIX 网关实现
type ApisixGateway struct {
	config   ApisixConfig
	client   *http.Client
	mu       sync.Mutex
	routeIDs map[string][]string // serviceID -> []routeID, 用于注销时清理
}

// NewApisixGateway 创建 APISIX 网关实例
func NewApisixGateway(apisixBytes []byte) (*ApisixGateway, error) {
	config, err := utils.Bytes2Struct[ApisixConfig](apisixBytes)
	if err != nil {
		return nil, fmt.Errorf("apisix config error: %w", err)
	}

	if config.AdminURL == "" {
		return nil, fmt.Errorf("apisix admin_url is required")
	}
	timeout := 5
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	return &ApisixGateway{
		config:   config,
		client:   &http.Client{Timeout: time.Duration(timeout) * time.Second},
		routeIDs: make(map[string][]string),
	}, nil
}

func (a *ApisixGateway) GetApisixConfig() ApisixConfig {
	return a.config
}

// RegisterService 注册服务到 APISIX (创建 upstream + routes)
func (a *ApisixGateway) RegisterService(ctx context.Context, svc *GatewayService) error {
	// 1. 创建/更新 upstream
	upstreamID := svc.ID
	upstreamBody := toUpstreamBody(svc.Upstream, svc.Name)
	if err := a.putResource(ctx, "upstreams", upstreamID, upstreamBody); err != nil {
		return fmt.Errorf("create upstream failed: %w", err)
	}

	// 2. 创建 routes
	a.mu.Lock()
	a.routeIDs[svc.ID] = nil // 重置
	a.mu.Unlock()

	for i, route := range svc.Routes {
		routeID := fmt.Sprintf("%s-route-%d", svc.ID, i)
		routeBody := map[string]interface{}{
			"uri":  route.URI,
			"name": route.Name,
		}
		if route.Upstream != nil {
			routeBody["upstream"] = toUpstreamBody(route.Upstream, route.Name)
		} else {
			routeBody["upstream_id"] = upstreamID
		}
		if len(route.Methods) > 0 {
			routeBody["methods"] = route.Methods
		}
		if err := a.putResource(ctx, "routes", routeID, routeBody); err != nil {
			return fmt.Errorf("create route %s failed: %w", routeID, err)
		}

		a.mu.Lock()
		a.routeIDs[svc.ID] = append(a.routeIDs[svc.ID], routeID)
		a.mu.Unlock()
	}

	return nil
}

func toUpstreamBody(upstream *GatewayUpstream, name string) map[string]interface{} {
	body := map[string]interface{}{
		"type":  upstream.Type,
		"nodes": upstream.Nodes,
		"name":  name,
	}
	if upstream.Scheme != "" {
		body["scheme"] = upstream.Scheme
	}
	return body
}

// DeregisterService 从 APISIX 注销服务 (删除 routes + upstream)
func (a *ApisixGateway) DeregisterService(ctx context.Context, serviceID string) error {
	a.mu.Lock()
	routeIDs := a.routeIDs[serviceID]
	delete(a.routeIDs, serviceID)
	a.mu.Unlock()

	// 先删 routes, 再删 upstream (有依赖关系)
	for _, routeID := range routeIDs {
		if err := a.deleteResource(ctx, "routes", routeID); err != nil {
			return fmt.Errorf("delete route %s failed: %w", routeID, err)
		}
	}

	if err := a.deleteResource(ctx, "upstreams", serviceID); err != nil {
		return fmt.Errorf("delete upstream %s failed: %w", serviceID, err)
	}

	return nil
}

// Close 关闭
func (a *ApisixGateway) Close() error {
	a.client.CloseIdleConnections()
	return nil
}

// putResource 创建或更新 APISIX 资源 (upstream/route)
func (a *ApisixGateway) putResource(ctx context.Context, resource, id string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/apisix/admin/%s/%s", a.config.AdminURL, resource, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("X-API-KEY", a.config.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apisix %s %s/%s returned %d: %s", http.MethodPut, resource, id, resp.StatusCode, string(respBody))
	}
	return nil
}

// deleteResource 删除 APISIX 资源
func (a *ApisixGateway) deleteResource(ctx context.Context, resource, id string) error {
	url := fmt.Sprintf("%s/apisix/admin/%s/%s", a.config.AdminURL, resource, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if a.config.APIKey != "" {
		req.Header.Set("X-API-KEY", a.config.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apisix DELETE %s/%s returned %d: %s", resource, id, resp.StatusCode, string(respBody))
	}
	return nil
}
