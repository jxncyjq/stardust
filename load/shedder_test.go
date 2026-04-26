package load

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdaptiveShedder_Allow(t *testing.T) {
	shedder := NewAdaptiveShedder()
	promise, err := shedder.Allow()
	assert.NoError(t, err)
	assert.NotNil(t, promise)
	promise.Pass()
}

func TestAdaptiveShedder_Overload(t *testing.T) {
	shedder := NewAdaptiveShedder(WithBuckets(5), WithWindow(time.Second))
	// 测试基本流程不 panic
	for i := 0; i < 10; i++ {
		promise, err := shedder.Allow()
		if err == nil {
			time.Sleep(time.Millisecond)
			promise.Pass()
		}
	}
}
