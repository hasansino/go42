package memcached

import (
	"testing"
	"time"
)

func TestExpirationSeconds(t *testing.T) {
	const maxInt32 = 1<<31 - 1

	tests := []struct {
		name string
		ttl  time.Duration
		want int32
	}{
		{name: "negative", ttl: -time.Second, want: 0},
		{name: "zero", ttl: 0, want: 0},
		{name: "sub-second", ttl: time.Nanosecond, want: 1},
		{name: "whole second", ttl: time.Second, want: 1},
		{name: "partial second", ttl: time.Second + time.Nanosecond, want: 2},
		{name: "maximum", ttl: maxInt32 * time.Second, want: maxInt32},
		{name: "overflow", ttl: maxInt32*time.Second + time.Nanosecond, want: maxInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expirationSeconds(tt.ttl); got != tt.want {
				t.Errorf("expirationSeconds(%s) = %d, want %d", tt.ttl, got, tt.want)
			}
		})
	}
}
