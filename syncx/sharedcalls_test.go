package syncx

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSharedCalls_Do(t *testing.T) {
	sc := NewSharedCalls()
	var callCount atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := sc.Do("key", func() (interface{}, error) {
				callCount.Add(1)
				time.Sleep(10 * time.Millisecond) // 确保并发请求能合并
				return "result", nil
			})
			assert.NoError(t, err)
			assert.Equal(t, "result", val)
		}()
	}

	wg.Wait()
	// 应该只调用极少次
	assert.LessOrEqual(t, int(callCount.Load()), 5)
}

func TestSharedCalls_DoEx(t *testing.T) {
	sc := NewSharedCalls()
	val, fresh, err := sc.DoEx("key", func() (interface{}, error) {
		return "hello", nil
	})
	assert.NoError(t, err)
	assert.True(t, fresh)
	assert.Equal(t, "hello", val)
}
