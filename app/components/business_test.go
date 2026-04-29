package components

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	httpServer "github.com/jxncyjq/stardust/http_server"
	"github.com/jxncyjq/stardust/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	assertWaitTimeout = time.Second
	assertWaitTick    = 10 * time.Millisecond
)

type businessTestService struct {
	started atomic.Bool
	stopped atomic.Bool
	http    atomic.Bool
	grpc    atomic.Bool
}

func (s *businessTestService) Start() { s.started.Store(true) }
func (s *businessTestService) Stop()  { s.stopped.Store(true) }
func (s *businessTestService) SetupHTTP(_ *httpServer.HttpServer) {
	s.http.Store(true)
}
func (s *businessTestService) SetupGRPC(_ *grpc.Server) {
	s.grpc.Store(true)
}

type lifecycleOnlyService struct {
	started atomic.Bool
	stopped atomic.Bool
}

func (s *lifecycleOnlyService) Start() { s.started.Store(true) }
func (s *lifecycleOnlyService) Stop()  { s.stopped.Store(true) }

func TestBusinessComponent_Lifecycle(t *testing.T) {
	svc := &businessTestService{}
	component := Business(service.NewServiceGroup(), svc).
		WithDependencies("logs", "redis")

	assert.Equal(t, "business", component.Name())
	assert.Equal(t, []string{"logs", "redis"}, component.Dependencies())

	assert.NoError(t, component.Init(context.Background(), nil))
	assert.NoError(t, component.Start(context.Background()))
	assert.Eventually(t, func() bool {
		return svc.started.Load()
	}, assertWaitTimeout, assertWaitTick)

	assert.NoError(t, component.Stop(context.Background()))
	assert.True(t, svc.stopped.Load())
}

func TestBusinessComponent_Binders(t *testing.T) {
	svc := &businessTestService{}
	component := Business(service.NewServiceGroup(), svc)

	component.SetupHTTP(nil)
	component.SetupGRPC(nil)

	assert.True(t, svc.http.Load())
	assert.True(t, svc.grpc.Load())
}

func TestBusinessComponent_LifecycleOnlyService(t *testing.T) {
	svc := &lifecycleOnlyService{}
	component := Business(service.NewServiceGroup(), svc)

	component.SetupHTTP(nil)
	component.SetupGRPC(nil)

	assert.NoError(t, component.Init(context.Background(), nil))
	assert.NoError(t, component.Start(context.Background()))
	assert.Eventually(t, func() bool {
		return svc.started.Load()
	}, assertWaitTimeout, assertWaitTick)

	assert.NoError(t, component.Stop(context.Background()))
	assert.True(t, svc.stopped.Load())
}
