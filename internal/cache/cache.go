package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
	Shutdown(ctx context.Context) error
}

// NoopCache is a cache that does nothing.
type NoopCache struct{}

func NewNoop() *NoopCache {
	return &NoopCache{}
}

func (NoopCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (NoopCache) Set(ctx context.Context, key string, value string) error {
	return nil
}

func (NoopCache) SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (NoopCache) Shutdown(_ context.Context) error {
	return nil
}
