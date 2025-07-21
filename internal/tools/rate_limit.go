package tools

import (
	"sync"

	ratepkg "golang.org/x/time/rate"
)

// RateLimiter is a simple rate limiter that allows limiting requests based on a key.
type RateLimiter struct {
	sync.RWMutex
	rate     int
	burst    int
	limiters map[any]*ratepkg.Limiter
}

func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		burst:    burst,
		limiters: make(map[any]*ratepkg.Limiter),
	}
}

func (m *RateLimiter) Limit(key any) bool {
	if key == nil {
		if str, ok := key.(string); ok && str == "" {
			// forbid empty keys - treat it as mistake
			return false
		}
	}

	m.Lock()
	_, ok := m.limiters[key]
	if !ok {
		m.limiters[key] = ratepkg.NewLimiter(ratepkg.Limit(m.rate), m.burst)
	}
	m.Unlock()

	// limiter itself is thread-safe
	return m.limiters[key].Allow()
}
