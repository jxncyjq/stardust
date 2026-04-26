package httpServer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jxncyjq/stardust/register"
	"google.golang.org/grpc/resolver"
)

const etcdScheme = "etcd"

type etcdResolverBuilder struct {
	config *register.EtcdConfig
}

// NewEtcdResolverBuilder 创建 etcd 服务发现的 resolver
func NewEtcdResolverBuilder(config *register.EtcdConfig) resolver.Builder {
	return &etcdResolverBuilder{config: config}
}

func (b *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	byteConfig, err := json.Marshal(b.config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal etcd config: %w", err)
	}
	reg, err := register.NewEtcdRegister(byteConfig)
	if err != nil {
		return nil, err
	}

	r := &etcdResolver{
		cc:          cc,
		register:    reg,
		serviceName: target.Endpoint(),
	}

	go r.watch()
	return r, nil
}

func (b *etcdResolverBuilder) Scheme() string {
	return etcdScheme
}

type etcdResolver struct {
	cc          resolver.ClientConn
	register    *register.EtcdRegister
	serviceName string
	cancel      context.CancelFunc
}

func (r *etcdResolver) watch() {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	ch, err := r.register.Watch(ctx, r.serviceName)
	if err != nil {
		return
	}

	for services := range ch {
		addrs := make([]resolver.Address, 0, len(services))
		for _, svc := range services {
			addrs = append(addrs, resolver.Address{
				Addr:       fmt.Sprintf("%s:%d", svc.Address, svc.Port),
				ServerName: svc.Name,
			})
		}
		r.cc.UpdateState(resolver.State{Addresses: addrs})
	}
}

func (r *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *etcdResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	r.register.Close()
}
