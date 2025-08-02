package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
)

// NoCache is passed as TTL to indicate that no expiration should be set.
const NoCache = time.Duration(0)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Invalidate(ctx context.Context, key string) error
}

type Engine interface {
	Cache
	Shutdown(ctx context.Context) error
}

// ----

// NoopCache is a no-op implementation of Cache.
type NoopCache struct{}

func NewNoop() *NoopCache {
	return &NoopCache{}
}

func (NoopCache) Get(_ context.Context, key string) (string, error) {
	return "", nil
}

func (NoopCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (NoopCache) Invalidate(_ context.Context, key string) error { return nil }

func (NoopCache) Shutdown(_ context.Context) error {
	return nil
}

// ----

// SetEncode serializes a value of type T using gob encoding and stores it in the cache.
// Value should be passed by value, not by pointer.
func SetEncode[T any](ctx context.Context, cache Cache, key string, value T, ttl time.Duration) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return fmt.Errorf("gob encode failed: %w", err)
	}
	return cache.Set(ctx, key, buf.String(), ttl)
}

// GetDecode retrieves a value of type T from the cache, decoding it from gob format.
// If the value is not found, it returns the zero value of T and no error.
// Best to use pointer to T, to distinguish between "not found" and "found but empty" cases.
func GetDecode[T any](ctx context.Context, cache Cache, key string) (T, error) {
	var out T
	str, err := cache.Get(ctx, key)
	if err != nil {
		return out, err
	}
	if str == "" {
		return out, nil
	}
	err = gob.NewDecoder(bytes.NewBufferString(str)).Decode(&out)
	if err != nil {
		return out, fmt.Errorf("gob decode failed: %w", err)
	}
	return out, nil
}
