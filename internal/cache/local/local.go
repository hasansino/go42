package local

import (
	"context"
	"runtime"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type Wrapper struct {
	cache       *ttlcache.Cache[string, string]
	cleanupDone chan struct{}
}

func New(opts ...Option) *Wrapper {
	cfg := config{
		capacity:     10_000,
		maxCostBytes: 1 << 30,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	cacheOpts := []ttlcache.Option[string, string]{
		ttlcache.WithDisableTouchOnHit[string, string](),
	}
	if cfg.capacity > 0 {
		cacheOpts = append(cacheOpts, ttlcache.WithCapacity[string, string](cfg.capacity))
	}
	if cfg.maxCostBytes > 0 {
		cacheOpts = append(cacheOpts, ttlcache.WithMaxCost[string, string](
			cfg.maxCostBytes,
			func(item ttlcache.CostItem[string, string]) uint64 {
				return uint64(len(item.Key) + len(item.Value))
			},
		))
	}

	c := ttlcache.New[string, string](cacheOpts...)
	cleanupDone := make(chan struct{})

	go func() {
		c.Start()
		close(cleanupDone)
	}()

	return &Wrapper{
		cache:       c,
		cleanupDone: cleanupDone,
	}
}

func (w *Wrapper) Shutdown(ctx context.Context) error {
	for {
		w.cache.Stop()
		select {
		case <-w.cleanupDone:
			w.cache.DeleteAll()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			runtime.Gosched()
		}
	}
}

func (w *Wrapper) Get(_ context.Context, key string) (string, bool, error) {
	item := w.cache.Get(key)
	if item == nil {
		return "", false, nil
	}
	return item.Value(), true, nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	w.cache.Set(key, value, normalizedTTL(ttl))
	return nil
}

func (w *Wrapper) SetIfAbsent(
	_ context.Context,
	key string,
	value string,
	ttl time.Duration,
) (bool, error) {
	_, found := w.cache.GetOrSet(
		key,
		value,
		ttlcache.WithTTL[string, string](normalizedTTL(ttl)),
	)
	return !found, nil
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	w.cache.Delete(key)
	return nil
}

func normalizedTTL(ttl time.Duration) time.Duration {
	if ttl == 0 {
		return ttlcache.NoTTL
	}
	return ttl
}
