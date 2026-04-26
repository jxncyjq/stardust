package service

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeStarter struct {
	startCalled atomic.Bool
	stopCalled  atomic.Bool
}

func (f *fakeStarter) Startup() error {
	f.startCalled.Store(true)
	return nil
}
func (f *fakeStarter) Stop() {
	f.stopCalled.Store(true)
}

func TestServerStarter(t *testing.T) {
	fake := &fakeStarter{}
	starter := NewServerStarter(fake)

	go starter.Start()
	time.Sleep(50 * time.Millisecond)
	assert.True(t, fake.startCalled.Load())

	starter.Stop()
	assert.True(t, fake.stopCalled.Load())
}
