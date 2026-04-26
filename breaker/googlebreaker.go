package breaker

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	// k 是 Google SRE 公式中的倍率，K 越大越宽容
	k          = 1.5
	buckets    = 40
	bucketTime = time.Millisecond * 250
	minReqs    = 100
)

// rollingWindow 滑动窗口
type rollingWindow struct {
	mu       sync.Mutex
	buckets  []bucket
	size     int
	offset   int
	lastTime time.Time
}

type bucket struct {
	accepts int64
	total   int64
}

func newRollingWindow(size int) *rollingWindow {
	return &rollingWindow{
		buckets:  make([]bucket, size),
		size:     size,
		lastTime: time.Now(),
	}
}

func (rw *rollingWindow) advance() {
	now := time.Now()
	elapsed := int(now.Sub(rw.lastTime) / bucketTime)
	if elapsed <= 0 {
		return
	}
	if elapsed > rw.size {
		elapsed = rw.size
	}
	for i := 0; i < elapsed; i++ {
		rw.offset = (rw.offset + 1) % rw.size
		rw.buckets[rw.offset] = bucket{}
	}
	rw.lastTime = now
}

func (rw *rollingWindow) add(accepts, total int64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.advance()
	rw.buckets[rw.offset].accepts += accepts
	rw.buckets[rw.offset].total += total
}

func (rw *rollingWindow) sum() (accepts, total int64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.advance()
	for _, b := range rw.buckets {
		accepts += b.accepts
		total += b.total
	}
	return
}

// GoogleBreaker Google SRE 风格熔断器
// 丢弃概率: max(0, (requests - K * accepts) / (requests + 1))
type GoogleBreaker struct {
	window *rollingWindow
}

func NewGoogleBreaker() Breaker {
	return &GoogleBreaker{
		window: newRollingWindow(buckets),
	}
}

func (gb *GoogleBreaker) Allow() (Promise, error) {
	if err := gb.accept(); err != nil {
		return nil, err
	}
	return &googlePromise{breaker: gb}, nil
}

func (gb *GoogleBreaker) Do(req func() error) error {
	return gb.DoWithAcceptable(req, func(err error) bool { return false })
}

func (gb *GoogleBreaker) DoWithAcceptable(req func() error, acceptable func(err error) bool) error {
	promise, err := gb.Allow()
	if err != nil {
		return err
	}

	err = req()
	if err == nil || acceptable(err) {
		promise.Accept()
	} else {
		promise.Reject(err)
	}
	return err
}

func (gb *GoogleBreaker) accept() error {
	accepts, total := gb.window.sum()
	// 请求量不足，不熔断
	if total < minReqs {
		return nil
	}

	dropRatio := gb.dropRatio(accepts, total)
	if dropRatio <= 0 {
		return nil
	}

	shouldDrop := rand.Float64() < dropRatio

	if shouldDrop {
		return ErrServiceUnavailable
	}
	return nil
}

func (gb *GoogleBreaker) dropRatio(accepts, total int64) float64 {
	ratio := math.Max(0, (float64(total)-k*float64(accepts))/float64(total+1))
	return ratio
}

func (gb *GoogleBreaker) markSuccess() {
	gb.window.add(1, 1)
}

func (gb *GoogleBreaker) markFailure() {
	gb.window.add(0, 1)
}

// googlePromise 请求承诺实现
type googlePromise struct {
	breaker *GoogleBreaker
}

func (p *googlePromise) Accept() {
	p.breaker.markSuccess()
}

func (p *googlePromise) Reject(_ error) {
	p.breaker.markFailure()
}
