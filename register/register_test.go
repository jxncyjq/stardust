package register

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jxncyjq/stardust/logs"
)

func TestMain(m *testing.M) {
	logs.Init([]byte(`{"level":-1}`))
	os.Exit(m.Run())
}

// MockRegister 模拟注册器用于测试
type MockRegister struct {
	mu       sync.RWMutex
	services map[string][]*ServiceInfo
	watchers map[string][]chan []*ServiceInfo
}

func NewMockRegister() *MockRegister {
	return &MockRegister{
		services: make(map[string][]*ServiceInfo),
		watchers: make(map[string][]chan []*ServiceInfo),
	}
}

func (m *MockRegister) Register(ctx context.Context, info *ServiceInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services[info.Name] = append(m.services[info.Name], info)
	m.notifyWatchers(info.Name)
	return nil
}

func (m *MockRegister) Deregister(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, services := range m.services {
		for i, s := range services {
			if s.ID == serviceID {
				m.services[name] = append(services[:i], services[i+1:]...)
				m.notifyWatchers(name)
				return nil
			}
		}
	}
	return nil
}

func (m *MockRegister) GetService(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services[serviceName], nil
}

func (m *MockRegister) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan []*ServiceInfo, 10)
	m.watchers[serviceName] = append(m.watchers[serviceName], ch)
	if services, ok := m.services[serviceName]; ok {
		ch <- services
	}
	return ch, nil
}

func (m *MockRegister) notifyWatchers(serviceName string) {
	if watchers, ok := m.watchers[serviceName]; ok {
		services := m.services[serviceName]
		for _, ch := range watchers {
			select {
			case ch <- services:
			default:
			}
		}
	}
}

func (m *MockRegister) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, watchers := range m.watchers {
		for _, ch := range watchers {
			close(ch)
		}
	}
	return nil
}

func TestServiceInfo(t *testing.T) {
	info := &ServiceInfo{
		Name:    "test-service",
		ID:      "test-1",
		Address: "127.0.0.1",
		Port:    8080,
		Tags:    []string{"http", "api"},
		Meta:    map[string]string{"version": "1.0"},
	}

	if info.Name != "test-service" {
		t.Errorf("Expected name 'test-service', got '%s'", info.Name)
	}
	if info.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", info.Port)
	}
}

func TestMockRegister(t *testing.T) {
	ctx := context.Background()
	reg := NewMockRegister()
	defer reg.Close()

	// 测试注册
	info := &ServiceInfo{
		Name:    "test-service",
		ID:      "test-1",
		Address: "127.0.0.1",
		Port:    8080,
	}

	err := reg.Register(ctx, info)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试获取服务
	services, err := reg.GetService(ctx, "test-service")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if services[0].ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", services[0].ID)
	}

	// 测试注销
	err = reg.Deregister(ctx, "test-1")
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	services, _ = reg.GetService(ctx, "test-service")
	if len(services) != 0 {
		t.Errorf("Expected 0 services after deregister, got %d", len(services))
	}
}

func TestServiceRegistry(t *testing.T) {
	reg := NewMockRegister()
	defer reg.Close()

	registry := NewServiceRegistry(reg)

	// 测试注册服务
	err := registry.Register("my-service", "127.0.0.1", 8080, []string{"http"}, nil)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试发现服务
	services, err := registry.Discover("my-service")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	// 测试注销
	err = registry.Deregister()
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	services, _ = registry.Discover("my-service")
	if len(services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(services))
	}
}

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg := NewMockRegister()
	defer reg.Close()

	// 开始监听
	ch, err := reg.Watch(ctx, "watch-service")
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// 注册服务
	info := &ServiceInfo{
		Name:    "watch-service",
		ID:      "watch-1",
		Address: "127.0.0.1",
		Port:    8080,
	}
	reg.Register(ctx, info)

	// 接收通知
	services := <-ch
	if len(services) != 1 {
		t.Errorf("Expected 1 service in watch, got %d", len(services))
	}
}

// --- Gateway + APISIX tests ---

type MockGateway struct {
	mu       sync.Mutex
	services map[string]*GatewayService
}

func NewMockGateway() *MockGateway {
	return &MockGateway{services: make(map[string]*GatewayService)}
}

func (g *MockGateway) RegisterService(ctx context.Context, svc *GatewayService) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.services[svc.ID] = svc
	return nil
}

func (g *MockGateway) DeregisterService(ctx context.Context, serviceID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.services, serviceID)
	return nil
}

func (g *MockGateway) Close() error { return nil }

func TestGatewayInterface(t *testing.T) {
	gw := NewMockGateway()
	ctx := context.Background()

	svc := &GatewayService{
		ID:   "example-service",
		Name: "example-service",
		Upstream: &GatewayUpstream{
			Type:  "roundrobin",
			Nodes: map[string]int{"127.0.0.1:8080": 1},
		},
		Routes: []*GatewayRoute{
			{Name: "hello", URI: "/api/*", Methods: []string{"GET", "POST"}},
		},
	}

	if err := gw.RegisterService(ctx, svc); err != nil {
		t.Fatalf("RegisterService failed: %v", err)
	}
	if len(gw.services) != 1 {
		t.Errorf("expected 1 service, got %d", len(gw.services))
	}

	if err := gw.DeregisterService(ctx, "example-service"); err != nil {
		t.Fatalf("DeregisterService failed: %v", err)
	}
	if len(gw.services) != 0 {
		t.Errorf("expected 0 services, got %d", len(gw.services))
	}
}

func TestApisixGateway_RegisterAndDeregister(t *testing.T) {
	var mu sync.Mutex
	upstreams := make(map[string]json.RawMessage)
	routes := make(map[string]json.RawMessage)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/apisix/admin/"), "/")
		if len(parts) != 2 {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		resource, id := parts[0], parts[1]

		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			switch resource {
			case "upstreams":
				upstreams[id] = body
			case "routes":
				routes[id] = body
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"key":"` + id + `"}`))

		case http.MethodDelete:
			switch resource {
			case "upstreams":
				delete(upstreams, id)
			case "routes":
				delete(routes, id)
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	apisixBytes, err := json.Marshal(&ApisixConfig{
		AdminURL: server.URL,
		APIKey:   "test-key",
		Timeout:  5,
	})
	if err != nil {
		t.Fatalf("json.Marshal(apisix config) error = %v", err)
	}

	gw, err := NewApisixGateway(apisixBytes)
	if err != nil {
		t.Fatalf("NewApisixGateway failed: %v", err)
	}
	defer gw.Close()

	ctx := context.Background()

	svc := &GatewayService{
		ID:   "example-svc",
		Name: "example-service",
		Upstream: &GatewayUpstream{
			Type:  "roundrobin",
			Nodes: map[string]int{"127.0.0.1:8080": 1},
		},
		Routes: []*GatewayRoute{
			{Name: "hello-api", URI: "/api/*", Methods: []string{"GET", "POST"}},
			{Name: "health", URI: "/health", Methods: []string{"GET"}},
		},
	}

	if err := gw.RegisterService(ctx, svc); err != nil {
		t.Fatalf("RegisterService failed: %v", err)
	}

	if len(upstreams) != 1 {
		t.Errorf("expected 1 upstream, got %d", len(upstreams))
	}
	if _, ok := upstreams["example-svc"]; !ok {
		t.Error("upstream 'example-svc' not found")
	}

	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}

	var routeBody map[string]interface{}
	json.Unmarshal(routes["example-svc-route-0"], &routeBody)
	if routeBody["uri"] != "/api/*" {
		t.Errorf("expected route uri '/api/*', got %v", routeBody["uri"])
	}
	if routeBody["upstream_id"] != "example-svc" {
		t.Errorf("expected upstream_id 'example-svc', got %v", routeBody["upstream_id"])
	}

	if err := gw.DeregisterService(ctx, "example-svc"); err != nil {
		t.Fatalf("DeregisterService failed: %v", err)
	}

	if len(upstreams) != 0 {
		t.Errorf("expected 0 upstreams after deregister, got %d", len(upstreams))
	}
	if len(routes) != 0 {
		t.Errorf("expected 0 routes after deregister, got %d", len(routes))
	}
}

func TestApisixGateway_InvalidConfig(t *testing.T) {
	apisixBytes, err := json.Marshal(&ApisixConfig{})
	if err != nil {
		t.Fatalf("json.Marshal(empty apisix config) error = %v", err)
	}

	_, err = NewApisixGateway(apisixBytes)
	if err == nil {
		t.Error("expected error for empty AdminURL")
	}
}

func TestApisixGateway_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	apisixBytes, err := json.Marshal(&ApisixConfig{AdminURL: server.URL, APIKey: "wrong-key"})
	if err != nil {
		t.Fatalf("json.Marshal(apisix wrong-key config) error = %v", err)
	}

	gw, _ := NewApisixGateway(apisixBytes)
	err = gw.RegisterService(context.Background(), &GatewayService{
		ID:       "test",
		Upstream: &GatewayUpstream{Type: "roundrobin", Nodes: map[string]int{"127.0.0.1:80": 1}},
	})
	if err == nil {
		t.Error("expected auth error")
	}
}
