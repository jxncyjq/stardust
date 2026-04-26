package breaker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleBreaker_Allow(t *testing.T) {
	b := NewGoogleBreaker()
	// 初始状态应允许请求
	promise, err := b.Allow()
	assert.NoError(t, err)
	promise.Accept() // 标记成功
}

func TestGoogleBreaker_Reject(t *testing.T) {
	b := NewGoogleBreaker()
	// 连续失败触发熔断
	for i := 0; i < 110; i++ {
		p, err := b.Allow()
		if err == nil {
			p.Reject(errors.New("fail"))
		}
	}
	// 高失败率后应拒绝请求
	_, err := b.Allow()
	assert.Equal(t, ErrServiceUnavailable, err)
}

func TestGoogleBreaker_Recovery(t *testing.T) {
	b := NewGoogleBreaker()
	// 先触发熔断
	for i := 0; i < 110; i++ {
		p, err := b.Allow()
		if err == nil {
			p.Reject(errors.New("fail"))
		}
	}
	// 连续成功后恢复
	for i := 0; i < 500; i++ {
		p, err := b.Allow()
		if err == nil {
			p.Accept()
		}
	}
	_, err := b.Allow()
	assert.NoError(t, err)
}
