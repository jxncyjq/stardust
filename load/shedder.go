package load

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var ErrServiceOverloaded = errors.New("service overloaded")

// ShedderPromise 降载承诺
type ShedderPromise interface {
	Pass()
	Fail()
}

// Shedder 降载器接口
type Shedder interface {
	Allow() (ShedderPromise, error)
}

// Option 配置选项
type Option func(*AdaptiveShedder)

func WithBuckets(buckets int) Option {
	return func(as *AdaptiveShedder) {
		as.bucketNum = buckets
	}
}

func WithWindow(window time.Duration) Option {
	return func(as *AdaptiveShedder) {
		as.windowSize = window
	}
}

func WithCpuThreshold(threshold int64) Option {
	return func(as *AdaptiveShedder) {
		as.cpuThreshold = threshold
	}
}

// AdaptiveShedder 自适应降载器（参照 go-zero core/load）
// 基于 Little's Law: L = λ * W
// 当 flying > maxPass * minRt * windows 时触发降载
type AdaptiveShedder struct {
	cpuThreshold    int64
	windowSize      time.Duration
	bucketNum       int
	flying          int64 // 当前并发数
	avgFlying       float64
	mu              sync.Mutex
	dropTime        time.Time
	droppedRecently atomic.Bool
	passCounter     *rollingCounter
	rtCounter       *rollingCounter
}

type rollingCounter struct {
	mu       sync.Mutex
	buckets  []float64
	size     int
	offset   int
	lastTime time.Time
	interval time.Duration
}

func newRollingCounter(size int, interval time.Duration) *rollingCounter {
	return &rollingCounter{
		buckets:  make([]float64, size),
		size:     size,
		interval: interval,
		lastTime: time.Now(),
	}
}

func (rc *rollingCounter) advance() {
	now := time.Now()
	elapsed := int(now.Sub(rc.lastTime) / rc.interval)
	if elapsed <= 0 {
		return
	}
	if elapsed > rc.size {
		elapsed = rc.size
	}
	for i := 0; i < elapsed; i++ {
		rc.offset = (rc.offset + 1) % rc.size
		rc.buckets[rc.offset] = 0
	}
	rc.lastTime = now
}

func (rc *rollingCounter) Add(val float64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.advance()
	rc.buckets[rc.offset] += val
}

func (rc *rollingCounter) Avg() float64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.advance()
	var sum float64
	var count int
	for _, b := range rc.buckets {
		if b > 0 {
			sum += b
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (rc *rollingCounter) Max() float64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.advance()
	var max float64
	for _, b := range rc.buckets {
		if b > max {
			max = b
		}
	}
	return max
}

// NewAdaptiveShedder 创建自适应降载器
func NewAdaptiveShedder(opts ...Option) Shedder {
	as := &AdaptiveShedder{
		cpuThreshold: 900, // 90% CPU
		windowSize:   time.Second * 5,
		bucketNum:    50,
	}
	for _, opt := range opts {
		opt(as)
	}

	bucketDuration := as.windowSize / time.Duration(as.bucketNum)
	as.passCounter = newRollingCounter(as.bucketNum, bucketDuration)
	as.rtCounter = newRollingCounter(as.bucketNum, bucketDuration)
	return as
}

func (as *AdaptiveShedder) Allow() (ShedderPromise, error) {
	if as.shouldDrop() {
		as.dropTime = time.Now()
		as.droppedRecently.Store(true)
		return nil, ErrServiceOverloaded
	}

	atomic.AddInt64(&as.flying, 1)
	return &adaptivePromise{
		shedder: as,
		start:   time.Now(),
	}, nil
}

func (as *AdaptiveShedder) shouldDrop() bool {
	if as.systemOverloaded() {
		as.mu.Lock()
		flying := atomic.LoadInt64(&as.flying)
		maxFlight := as.maxFlight()
		as.mu.Unlock()
		return float64(flying) > maxFlight
	}
	return false
}

func (as *AdaptiveShedder) systemOverloaded() bool {
	// 优先检查实际 CPU 使用率（Linux 有效，其他平台返回 0）
	if getCpuUsage() > as.cpuThreshold {
		return true
	}
	// 兜底：最近 1 秒内有丢弃记录
	if as.droppedRecently.Load() {
		if time.Since(as.dropTime) < time.Second {
			return true
		}
		as.droppedRecently.Store(false)
	}
	return false
}

// maxFlight 基于 Little's Law 计算最大并发数
func (as *AdaptiveShedder) maxFlight() float64 {
	maxPass := as.passCounter.Max()
	minRt := math.Max(as.rtCounter.Avg(), 1) // 毫秒
	return maxPass * minRt / 1e3 * float64(as.bucketNum)
}

type adaptivePromise struct {
	shedder *AdaptiveShedder
	start   time.Time
}

func (p *adaptivePromise) Pass() {
	rt := float64(time.Since(p.start).Milliseconds())
	atomic.AddInt64(&p.shedder.flying, -1)
	p.shedder.passCounter.Add(1)
	p.shedder.rtCounter.Add(rt)
}

func (p *adaptivePromise) Fail() {
	atomic.AddInt64(&p.shedder.flying, -1)
}
