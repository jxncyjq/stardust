package limit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeriodLimiter_Allow(t *testing.T) {
	limiter := NewPeriodLimiter(1, 5, "test:period:limit") // 1秒5次
	for i := 0; i < 5; i++ {
		result, err := limiter.Take("user1")
		assert.NoError(t, err)
		assert.Equal(t, AllowedStatus, result)
	}
}

func TestPeriodLimiter_HitQuota(t *testing.T) {
	limiter := NewPeriodLimiter(1, 3, "test:period:quota")
	for i := 0; i < 3; i++ {
		limiter.Take("user2")
	}
	result, err := limiter.Take("user2")
	assert.NoError(t, err)
	assert.Equal(t, OverQuotaStatus, result)
}
