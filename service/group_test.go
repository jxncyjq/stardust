package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockService 用于测试的模拟服务
type mockService struct {
	started atomic.Bool
	stopped atomic.Bool
}

func (m *mockService) Start() { m.started.Store(true) }
func (m *mockService) Stop()  { m.stopped.Store(true) }

type orderedService struct {
	name   string
	mu     *sync.Mutex
	events *[]string
}

func (s *orderedService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	*s.events = append(*s.events, "start:"+s.name)
}

func (s *orderedService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	*s.events = append(*s.events, "stop:"+s.name)
}

func TestServiceGroup_StartStop(t *testing.T) {
	svc1 := &mockService{}
	svc2 := &mockService{}

	sg := NewServiceGroup()
	sg.Add(svc1)
	sg.Add(svc2)

	go sg.Start()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, svc1.started.Load())
	assert.True(t, svc2.started.Load())

	sg.Stop()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, svc1.stopped.Load())
	assert.True(t, svc2.stopped.Load())
}

func TestServiceGroup_ManagerInterface(t *testing.T) {
	svc1 := &mockService{}
	svc2 := &mockService{}

	var manager Manager = NewServiceGroup()
	manager.AddMany(svc1, nil, svc2)

	manager.StartAsync()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, svc1.started.Load())
	assert.True(t, svc2.started.Load())

	manager.Stop()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, svc1.stopped.Load())
	assert.True(t, svc2.stopped.Load())
}

func TestServiceGroup_StartAsyncOnlyStartsOnce(t *testing.T) {
	var starts atomic.Int32
	svc := serviceFunc{
		start: func() {
			starts.Add(1)
		},
	}

	sg := NewServiceGroup()
	sg.Add(svc)
	sg.StartAsync()
	sg.StartAsync()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, int32(1), starts.Load())
	sg.Stop()
}

func TestServiceGroup_StopReverseOrder(t *testing.T) {
	var mu sync.Mutex
	events := make([]string, 0)
	svc1 := &orderedService{name: "one", mu: &mu, events: &events}
	svc2 := &orderedService{name: "two", mu: &mu, events: &events}

	sg := NewServiceGroup()
	sg.AddMany(svc1, svc2)
	sg.Stop()

	assert.Equal(t, []string{"stop:two", "stop:one"}, events)
}

func TestServiceGroup_Empty(t *testing.T) {
	sg := NewServiceGroup()
	sg.Stop() // 不应 panic
}

type serviceFunc struct {
	start func()
	stop  func()
}

func (f serviceFunc) Start() {
	if f.start != nil {
		f.start()
	}
}

func (f serviceFunc) Stop() {
	if f.stop != nil {
		f.stop()
	}
}
