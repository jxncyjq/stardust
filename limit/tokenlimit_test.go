package limit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenLimiter_Allow(t *testing.T) {
	limiter := NewTokenLimiter(10, 10, "test:token:limit")
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.AllowN(time.Now(), 1))
	}
}

func TestTokenLimiter_Deny(t *testing.T) {
	limiter := NewTokenLimiter(5, 5, "test:token:deny")
	// 消耗所有令牌
	for i := 0; i < 5; i++ {
		limiter.AllowN(time.Now(), 1)
	}
	assert.False(t, limiter.AllowN(time.Now(), 1))
}
