package syncx

import "sync"

// SharedCalls 合并并发请求（参照 go-zero core/syncx.SharedCalls）
type SharedCalls interface {
	Do(key string, fn func() (interface{}, error)) (interface{}, error)
	DoEx(key string, fn func() (interface{}, error)) (interface{}, bool, error)
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type sharedCalls struct {
	mu    sync.Mutex
	calls map[string]*call
}

func NewSharedCalls() SharedCalls {
	return &sharedCalls{
		calls: make(map[string]*call),
	}
}

func (sc *sharedCalls) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	val, _, err := sc.DoEx(key, fn)
	return val, err
}

// DoEx 执行并返回是否是实际执行者
func (sc *sharedCalls) DoEx(key string, fn func() (interface{}, error)) (interface{}, bool, error) {
	sc.mu.Lock()
	if c, ok := sc.calls[key]; ok {
		sc.mu.Unlock()
		c.wg.Wait()
		return c.val, false, c.err
	}

	c := &call{}
	c.wg.Add(1)
	sc.calls[key] = c
	sc.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	sc.mu.Lock()
	delete(sc.calls, key)
	sc.mu.Unlock()

	return c.val, true, c.err
}
