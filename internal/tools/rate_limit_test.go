package tools_test

import (
	"testing"
	"time"

	"github.com/hasansino/go42/internal/tools"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := tools.NewRateLimiter(2, 2) // 2 req/sec, burst 2
	key := "user1"
	allowed := 0
	for i := 0; i < 2; i++ {
		if rl.Limit(key) {
			allowed++
		}
	}
	if allowed != 2 {
		t.Errorf("expected 2 allowed, got %d", allowed)
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	rl := tools.NewRateLimiter(1, 1) // 1 req/sec, burst 1
	key := "user2"
	if !rl.Limit(key) {
		t.Error("first request should be allowed")
	}
	if rl.Limit(key) {
		t.Error("second request should be rate limited")
	}
}

func TestRateLimiter_ResetsAfterTime(t *testing.T) {
	rl := tools.NewRateLimiter(1, 1)
	key := "user3"
	if !rl.Limit(key) {
		t.Error("first request should be allowed")
	}
	if rl.Limit(key) {
		t.Error("second request should be rate limited")
	}
	time.Sleep(time.Second)
	if !rl.Limit(key) {
		t.Error("should allow after 1 second")
	}
}

func TestRateLimiter_EmptyKey(t *testing.T) {
	rl := tools.NewRateLimiter(1, 1)
	if rl.Limit("") {
		t.Error("empty string key should not be allowed")
	}
	if rl.Limit(nil) {
		t.Error("nil key should not be allowed")
	}
}
