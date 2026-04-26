package register

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jxncyjq/stardust/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	etcdInstance *EtcdRegister
	etcdOnce     sync.Once
	etcdErr      error
)

// EtcdConfig etcd配置
type EtcdConfig struct {
	Endpoints   []string `json:"endpoints" yaml:"endpoints"`
	DialTimeout int      `json:"dial_timeout" yaml:"dial_timeout"` // 秒
	TTL         int64    `json:"ttl" yaml:"ttl"`                   // 租约TTL秒
	ServiceName string   `json:"service_name" yaml:"service_name"` // 服务名
	Address     string   `json:"address" yaml:"address"`           // 服务地址
	Port        int      `json:"port" yaml:"port"`                 // 服务端口
	Tags        []string `json:"tags" yaml:"tags"`                 // 服务标签
}

// EtcdRegister etcd服务注册实现
type EtcdRegister struct {
	client   *clientv3.Client
	leaseID  clientv3.LeaseID
	ttl      int64
	prefix   string
	registry *ServiceRegistry
	config   EtcdConfig
}

// Init 初始化etcd单例并注册服务
func Init(configBytes []byte) error {
	etcdOnce.Do(func() {
		var config EtcdConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			etcdErr = fmt.Errorf("register: parse etcd config failed: %w", err)
			return
		}

		timeout := 5
		if config.DialTimeout > 0 {
			timeout = config.DialTimeout
		}

		client, err := clientv3.New(clientv3.Config{
			Endpoints:   config.Endpoints,
			DialTimeout: time.Duration(timeout) * time.Second,
		})
		if err != nil {
			etcdErr = fmt.Errorf("register: connect etcd failed: %w", err)
			return
		}

		ttl := int64(10)
		if config.TTL > 0 {
			ttl = config.TTL
		}

		inst := &EtcdRegister{
			client: client,
			ttl:    ttl,
			prefix: "/services/",
			config: config,
		}

		// 注册服务
		inst.registry = NewServiceRegistry(inst)
		if err := inst.registry.Register(config.ServiceName, config.Address, config.Port, config.Tags, nil); err != nil {
			etcdErr = fmt.Errorf("register: register service failed: %w", err)
			return
		}
		etcdInstance = inst
	})
	return etcdErr
}

// GetEtcdRegister 获取etcd单例
func GetEtcdRegister() (*EtcdRegister, error) {
	if etcdErr != nil {
		return nil, etcdErr
	}
	return etcdInstance, nil
}

// NewEtcdRegister 创建etcd注册器
func NewEtcdRegister(etcdBytes []byte) (*EtcdRegister, error) {
	config, err := utils.Bytes2Struct[EtcdConfig](etcdBytes)
	if err != nil {
		return nil, fmt.Errorf("etcd config error: %w", err)
	}

	timeout := 5
	if config.DialTimeout > 0 {
		timeout = config.DialTimeout
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: time.Duration(timeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ttl := int64(10)
	if config.TTL > 0 {
		ttl = config.TTL
	}

	return &EtcdRegister{
		client: client,
		ttl:    ttl,
		prefix: "/services/",
		config: config,
	}, nil
}

func (r *EtcdRegister) GetConfig() EtcdConfig {
	return r.config
}

// Register 注册服务
func (r *EtcdRegister) Register(ctx context.Context, info *ServiceInfo) error {
	// 创建租约
	lease, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return err
	}
	r.leaseID = lease.ID

	// 序列化服务信息
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// 注册服务
	key := fmt.Sprintf("%s%s/%s", r.prefix, info.Name, info.ID)
	_, err = r.client.Put(ctx, key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	// 保持租约
	ch, err := r.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Deregister 注销服务
func (r *EtcdRegister) Deregister(ctx context.Context, serviceID string) error {
	if r.leaseID != 0 {
		_, err := r.client.Revoke(ctx, r.leaseID)
		return err
	}
	return nil
}

// GetService 获取服务列表
func (r *EtcdRegister) GetService(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	key := fmt.Sprintf("%s%s/", r.prefix, serviceName)
	resp, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var services []*ServiceInfo
	for _, kv := range resp.Kvs {
		var info ServiceInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			continue
		}
		services = append(services, &info)
	}
	return services, nil
}

// Watch 监听服务变化
func (r *EtcdRegister) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	key := fmt.Sprintf("%s%s/", r.prefix, serviceName)
	ch := make(chan []*ServiceInfo, 1)

	// 先获取当前服务列表
	services, err := r.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	ch <- services

	// 监听变化
	go func() {
		defer close(ch)
		watchCh := r.client.Watch(ctx, key, clientv3.WithPrefix())
		for range watchCh {
			services, err := r.GetService(ctx, serviceName)
			if err != nil {
				continue
			}
			select {
			case ch <- services:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// Close 关闭连接
func (r *EtcdRegister) Close() error {
	if r.registry != nil {
		r.registry.Deregister()
	}
	return r.client.Close()
}
