package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const rateLimitCacheKeyPrefix = "rate_limit:"

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/rate_limit.go

type cacheAccessor interface {
	AllowRateLimit(
		ctx context.Context,
		key string,
		rate int,
		burst int,
		ttl time.Duration,
	) (allowed bool, err error)
}

// RateLimiter applies a cache-backed rate limit to hashed, namespaced client keys.
type RateLimiter struct {
	cache     cacheAccessor
	namespace string
	rate      int
	burst     int
	ttl       time.Duration
}

func NewRateLimiter(
	cache cacheAccessor,
	namespace string,
	rate int,
	burst int,
	ttl time.Duration,
) *RateLimiter {
	if namespace == "" {
		namespace = "default"
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if rate > 0 && burst > 0 {
		minimumTTL := time.Duration(burst) * time.Second / time.Duration(rate)
		if ttl < minimumTTL {
			ttl = minimumTTL
		}
	}

	return &RateLimiter{
		cache:     cache,
		namespace: namespace,
		rate:      rate,
		burst:     burst,
		ttl:       ttl,
	}
}

func (m *RateLimiter) Limit(ctx context.Context, key string) (bool, error) {
	if key == "" {
		// forbid empty keys - treat them as mistake
		return false, nil
	}
	if m.cache == nil {
		return false, fmt.Errorf("rate limiter cache is nil")
	}
	if m.rate <= 0 {
		return false, fmt.Errorf("rate must be positive")
	}
	if m.burst <= 0 {
		return false, fmt.Errorf("burst must be positive")
	}

	sum := sha256.Sum256([]byte(key))
	cacheKey := rateLimitCacheKeyPrefix + m.namespace + ":" + hex.EncodeToString(sum[:])
	return m.cache.AllowRateLimit(ctx, cacheKey, m.rate, m.burst, m.ttl)
}
