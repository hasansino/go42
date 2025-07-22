package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
	Invalidate(ctx context.Context, key string) error
	Shutdown(ctx context.Context) error
}

// ---

// NoopCache is a no-op implementation of Cache.
type NoopCache struct{}

func NewNoop() *NoopCache {
	return &NoopCache{}
}

func (NoopCache) Get(_ context.Context, key string) (string, error) {
	return "", nil
}

func (NoopCache) Set(_ context.Context, key string, value string) error {
	return nil
}

func (NoopCache) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (NoopCache) Invalidate(_ context.Context, key string) error { return nil }

func (NoopCache) Shutdown(_ context.Context) error {
	return nil
}
