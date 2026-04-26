package service

import (
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

func TestServiceGroup_Empty(t *testing.T) {
	sg := NewServiceGroup()
	sg.Stop() // 不应 panic
}
